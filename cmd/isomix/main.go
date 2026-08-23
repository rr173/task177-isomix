// Command isomix 启动同位素混合来源约束求解服务。
//
// 用法：
//
//	go run ./cmd/isomix --addr :8080 --db ./isomix.db   # 启动 HTTP 服务
//	go run ./cmd/isomix --smoke-test                     # 端到端自检（不启动服务）
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"task177-isomix/internal/constraint"
	"task177-isomix/internal/diagnostics"
	"task177-isomix/internal/endmember"
	"task177-isomix/internal/export"
	"task177-isomix/internal/httpapi"
	"task177-isomix/internal/measure"
	"task177-isomix/internal/model"
	"task177-isomix/internal/provenance"
	"task177-isomix/internal/report"
	"task177-isomix/internal/sensitivity"
	"task177-isomix/internal/solve"
	"task177-isomix/internal/store"
	"task177-isomix/internal/uncertainty"
)

// signalChan 返回进程信号通道（SIGINT/SIGTERM 触发优雅关闭）。
func signalChan() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "", "SQLite database path (default: in-memory)")
	smoke := flag.Bool("smoke-test", false, "run end-to-end self test and exit")
	flag.Parse()

	if *smoke {
		if err := runSmokeTest(); err != nil {
			log.Fatalf("SMOKE TEST FAILED: %v", err)
		}
		fmt.Println("SMOKE TEST PASSED")
		return
	}

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	app := buildApp(st)
	if err := resumeIncomplete(app); err != nil {
		log.Printf("resume incomplete solutions: %v", err)
	}

	// 启动时对存量已发布报告做一次过期替代检查（幂等维护）。
	if _, err := app.Reports.SupersedeStale(); err != nil {
		log.Printf("supersede stale reports: %v", err)
	}

	mux := app.Routes()
	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 优雅退出：收到 SIGINT/SIGTERM 时关闭。
	go func() {
		<-signalChan()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	log.Printf("task177-isomix listening on %s", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}

// buildApp 组装领域服务与 HTTP 层。
func buildApp(st *store.Store) *httpapi.App {
	em := endmember.NewService(st.EndmemberStore, nil, nil)
	ms := measure.NewService(st.SampleStore, nil, nil)
	cs := constraint.NewService(st.ConstraintStore, nil, nil)
	sv := solve.NewService(st.SolutionStore, em, ms, cs, nil, nil)
	rp := report.NewService(st.ReportStore, sv, em, cs, nil, nil)
	return httpapi.NewApp(em, ms, cs, sv, rp, sensitivity.NewService(sv, nil))
}

// resumeIncomplete 恢复未完成求解并记录日志。
func resumeIncomplete(app *httpapi.App) error {
	recovered, errs := app.Solve.ResumeIncomplete()
	if recovered > 0 {
		log.Printf("resumed %d incomplete solution(s)", recovered)
	}
	if len(errs) > 0 {
		return fmt.Errorf("resume errors: %v", errs)
	}
	return nil
}

// runSmokeTest 端到端自检：真实创建实体、求解、发布报告，
// 关闭并重新打开数据库验证持久化与重启恢复，全部通过返回 nil。
func runSmokeTest() error {
	dir, err := os.MkdirTemp("", "isomix-smoke-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	dbFile := filepath.Join(dir, "smoke.db")

	// ---- 阶段 A：写入并求解 ----
	st, err := store.Open(dbFile)
	if err != nil {
		return err
	}
	app := buildApp(st)

	// 1. 三个端元（同位素 d18O / d2H），组成区间。
	em1, err := app.Endmembers.Create("mantle", map[string]model.Range{
		"d18O": {Lo: 5.2, Hi: 5.8}, "d2H": {Lo: -90, Hi: -70},
	}, nil)
	assert(err == nil, "create endmember mantle")
	_, err = app.Endmembers.Validate(em1.ID)
	assert(err == nil, "validate mantle")
	_, err = app.Endmembers.MakeAvailable(em1.ID)
	assert(err == nil, "make mantle available")

	em2, err := app.Endmembers.Create("crust", map[string]model.Range{
		"d18O": {Lo: 8.0, Hi: 9.5}, "d2H": {Lo: -60, Hi: -40},
	}, nil)
	assert(err == nil, "create endmember crust")
	_, _ = app.Endmembers.Validate(em2.ID)
	_, _ = app.Endmembers.MakeAvailable(em2.ID)

	em3, err := app.Endmembers.Create("organic", map[string]model.Range{
		"d18O": {Lo: 14.0, Hi: 18.0}, "d2H": {Lo: -160, Hi: -110},
	}, nil)
	assert(err == nil, "create endmember organic")
	_, _ = app.Endmembers.Validate(em3.ID)
	_, _ = app.Endmembers.MakeAvailable(em3.ID)

	// 2. 样品测量（区间设计为端元混合可行域内部，且相对宽度 < 0.5 判为可求解）。
	sp, err := app.Measure.Create("river-water", map[string]model.Range{
		"d18O": {Lo: 6.0, Hi: 8.0}, "d2H": {Lo: -80, Hi: -55},
	}, []float64{0.04, 0.002, 3.2}, nil)
	assert(err == nil, "create sample")
	spc, err := app.Measure.Check(sp.ID)
	assert(err == nil, "check sample solvable")
	assert(spc.Status == model.SampleSolvable, "sample judged solvable, got "+string(spc.Status))

	// 3. 求解：三端元可解释样品。
	sol1, err := app.Solve.Submit(sp.ID)
	assert(err == nil, "submit solution 1")
	assert(sol1.Status == model.SolutionFeasible, "solution 1 feasible, got "+string(sol1.Status))
	assert(sol1.State.Feasible, "solution 1 state feasible")
	b1 := sol1.State.EndmemberBounds
	assert(len(b1) == 3, "solution 1 bounds cover 3 endmembers")
	// 比例非负且和为一（抽查：各端元区间中值求和接近 1）。
	sumOK := checkSumOne(b1)
	assert(sumOK, "endmember bounds consistent with mass conservation")
	// 派生解释能力必须与同一持久化求解协作，不能只存在于孤立工具包。
	sensitivityReport, err := sensitivity.NewService(app.Solve, nil).Analyze(sol1.ID)
	assert(err == nil && sensitivityReport.Feasible, "sensitivity report available")
	assert(len(sensitivityReport.Bounds) == 3, "sensitivity covers each endmember")
	uncertaintyReport := uncertainty.Assess(*sp)
	assert(uncertaintyReport.PositiveDefinite, "sample uncertainty report accepts covariance")
	diagnosticReport := diagnostics.Analyze(sol1)
	assert(diagnosticReport.Risk == "normal" || diagnosticReport.Risk == "sparse-system", "solver diagnostics report low risk")

	// 幂等：同一样品再次提交返回同一输入哈希结果。
	sol1b, err := app.Solve.Submit(sp.ID)
	assert(err == nil, "resubmit solution 1 idempotent")
	assert(sol1b.ID == sol1.ID, "idempotent submit returns same solution")

	// 4. 排除约束：排除 organic -> 可行域收缩。
	_, err = app.Constraints.Create(model.ConstraintExclude, "exclude-organic", em3.ID, model.Range{Lo: 0, Hi: 0}, "field evidence")
	assert(err == nil, "create exclude constraint")
	sol2, err := app.Solve.Submit(sp.ID)
	assert(err == nil, "submit solution 2")
	assert(sol2.State.Feasible, "solution 2 feasible")
	b2 := sol2.State.EndmemberBounds
	assert(b2[em3.ID].Lo == 0 && b2[em3.ID].Hi == 0, "excluded endmember bounds collapse to zero")

	// 5. 矛盾约束：ratio(mantle:crust)=4 与 d18O 下界（>=6.0）冲突，
	//    服务返回最小冲突集（含 ratio 约束）。
	rc, err := app.Constraints.Create(model.ConstraintRatio, "strict-ratio", em1.ID+":"+em2.ID, model.Range{Lo: 4, Hi: 4}, "synthetic")
	assert(err == nil, "create ratio constraint")
	sol3, err := app.Solve.Submit(sp.ID)
	assert(err == nil, "submit solution 3")
	assert(!sol3.State.Feasible, "solution 3 infeasible")
	assert(sol3.Status == model.SolutionInfeasible, "solution 3 status infeasible")
	core := sol3.State.ConflictCore
	assert(len(core) >= 1 && contains(core, rc.ID), "conflict core contains the contradictory ratio constraint, got "+fmt.Sprint(core))

	// 6. 报告：对可行解 sol2 创建并发布。
	rp1, err := app.Reports.Create(sol2.ID, "river-water mixing report v1")
	assert(err == nil, "create report v1")
	_, err = app.Reports.Publish(rp1.ID)
	assert(err == nil, "publish report v1")
	assert(rp1.Status == model.ReportPublished || rp1.Status == model.ReportDraft, "report created")
	provenanceSnapshot := provenance.BuildSnapshot(*rp1, rp1.CreatedAt)
	assert(len(provenanceSnapshot.Events) == 5, "report provenance captures sample, endmembers and constraints")
	jsonExport, err := export.JSON(*rp1)
	assert(err == nil && len(jsonExport) > 0, "report JSON export available")
	csvExport, err := export.CSV(*rp1)
	assert(err == nil && len(csvExport) > 0, "report CSV export available")

	// 7. 修订端元 -> 旧报告过期；重新提交产生新解（不改写旧报告）。
	//    修订使端元回退为草拟，须重新校验并置为可用后才能再次参与求解。
	_, err = app.Endmembers.Revise(em1.ID, map[string]model.Range{
		"d18O": {Lo: 5.0, Hi: 6.0}, "d2H": {Lo: -95, Hi: -65},
	})
	assert(err == nil, "revise endmember mantle")
	_, err = app.Endmembers.Validate(em1.ID)
	assert(err == nil, "re-validate revised mantle")
	_, err = app.Endmembers.MakeAvailable(em1.ID)
	assert(err == nil, "re-available revised mantle")
	sol4, err := app.Solve.Submit(sp.ID)
	assert(err == nil, "submit solution 4 after revision")
	assert(sol4.ID != sol2.ID, "endmember revision triggers a new solution")
	// 修订收紧 d2H 下界后与 ratio>=4 冲突 -> 新求解不可行（修订只触发新求解）。
	assert(sol4.Status == model.SolutionInfeasible, "solution 4 infeasible after revision, got "+string(sol4.Status))
	stale, changes := app.Reports.CheckSuperseded(rp1)
	assert(stale, "report v1 stale after revision, changes="+fmt.Sprint(changes))
	superseded, err := app.Reports.SupersedeStale()
	assert(err == nil, "supersede stale reports")
	assert(superseded >= 1, "at least one report superseded")

	// 8. 在库中注入一条 queued 求解（模拟求解中崩溃），供重启恢复验证。
	injectQueued(st, sol4)

	// 9. 关闭数据库。
	if err := st.Close(); err != nil {
		return err
	}

	// ---- 阶段 B：重开同一数据库，验证持久化与恢复 ----
	st2, err := store.Open(dbFile)
	if err != nil {
		return err
	}
	defer st2.Close()
	app2 := buildApp(st2)

	// 数据仍在。
	ems2, err := app2.Endmembers.List()
	assert(err == nil && len(ems2) == 3, "endmembers persisted after reopen")
	sps2, err := app2.Measure.List()
	assert(err == nil && len(sps2) == 1, "samples persisted after reopen")
	sols2, err := app2.Solve.List()
	assert(err == nil && len(sols2) >= 3, "solutions persisted after reopen")
	rps2, err := app2.Reports.List()
	assert(err == nil && len(rps2) == 1, "reports persisted after reopen")

	// 幂等恢复：重开后同一样品提交（活动约束 [exclude, ratio] + 端元 v2 不变），
	// 仍命中已持久化的输入哈希（sol4）。
	sol1c, err := app2.Solve.Submit(sp.ID)
	assert(err == nil, "resubmit after reopen")
	assert(sol1c.ID == sol4.ID, "reopen idempotent hit existing solution, got "+sol1c.ID+" want "+sol4.ID)

	// 重启恢复：queued 记录被恢复为终态。
	recovered, errs := app2.Solve.ResumeIncomplete()
	assert(len(errs) == 0, "resume no errors")
	assert(recovered >= 1, "resume recovered queued solution, got %d", recovered)

	// 已发布报告在新库中仍为 published 或 superseded（不可变）。
	rp1c, err := app2.Reports.Get(rp1.ID)
	assert(err == nil, "report v1 still exists after reopen")
	assert(rp1c.Status == model.ReportSuperseded, "report v1 superseded after revision+reopen")

	return nil
}

// injectQueued 注入一条 queued 求解记录（模拟求解中进程崩溃，
// 其输入与既有可行解完全一致，重开后可经输入哈希恢复重算）。
func injectQueued(st *store.Store, base *model.Solution) {
	queued := &model.Solution{
		ID:            "so_injected_queued",
		SampleID:      base.SampleID,
		Status:        model.SolutionQueued,
		InputHash:     base.InputHash,
		EndmemberIDs:  append([]string(nil), base.EndmemberIDs...),
		ConstraintIDs: append([]string(nil), base.ConstraintIDs...),
		CreatedAt:     time.Now(),
	}
	if err := st.SolutionStore.CreateSolution(queued); err != nil {
		panic("inject queued solution: " + err.Error())
	}
}

// assert 断言，失败即 panic 抛出测试失败。
func assert(cond bool, format string, args ...any) {
	if !cond {
		panic(fmt.Sprintf("assertion failed: "+format, args...))
	}
}

// checkSumOne 抽查各端元可行区间是否与质量守恒一致（取中值求和接近 1）。
func checkSumOne(bounds map[string]model.Range) bool {
	sum := 0.0
	for _, r := range bounds {
		sum += (r.Lo + r.Hi) / 2
	}
	return sum > 0.9 && sum < 1.1
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
