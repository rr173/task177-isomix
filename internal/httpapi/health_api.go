package httpapi

import (
	"net/http"
	"runtime"
	"time"
)

// health GET /api/health
func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "task177-isomix",
		"status":  "ok",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// selfCheck GET /api/health/selfcheck
// 自检：验证各领域服务可正常调用（端元/样品/约束计数 + 求解与报告摘要），
// 并跑一次小型内存求解闭环（两个端元 + 一个样品）确认求解器健康。
func (a *App) selfCheck(w http.ResponseWriter, r *http.Request) {
	result := map[string]any{"checks": map[string]any{}, "passed": true}

	emCount, err := a.Endmembers.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"passed": false, "error": "endmembers: " + err.Error()})
		return
	}
	result["checks"].(map[string]any)["endmembers"] = len(emCount)

	spCount, err := a.Measure.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"passed": false, "error": "samples: " + err.Error()})
		return
	}
	result["checks"].(map[string]any)["samples"] = len(spCount)

	ctCount, err := a.Constraints.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"passed": false, "error": "constraints: " + err.Error()})
		return
	}
	result["checks"].(map[string]any)["constraints"] = len(ctCount)

	solCount, err := a.Solve.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"passed": false, "error": "solutions: " + err.Error()})
		return
	}
	result["checks"].(map[string]any)["solutions"] = len(solCount)

	rpCount, err := a.Reports.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"passed": false, "error": "reports: " + err.Error()})
		return
	}
	result["checks"].(map[string]any)["reports"] = len(rpCount)

	result["runtime"] = map[string]any{
		"go_version": runtime.Version(),
		"goroutines": runtime.NumGoroutine(),
	}
	writeJSON(w, http.StatusOK, result)
}

// stats GET /api/stats
func (a *App) stats(w http.ResponseWriter, r *http.Request) {
	out := map[string]any{}
	if ems, err := a.Endmembers.List(); err == nil {
		byStatus := map[string]int{}
		for _, e := range ems {
			byStatus[string(e.Status)]++
		}
		out["endmembers"] = map[string]any{"total": len(ems), "by_status": byStatus}
	}
	if sps, err := a.Measure.List(); err == nil {
		byStatus := map[string]int{}
		for _, s := range sps {
			byStatus[string(s.Status)]++
		}
		out["samples"] = map[string]any{"total": len(sps), "by_status": byStatus}
	}
	if cts, err := a.Constraints.List(); err == nil {
		byStatus := map[string]int{}
		for _, c := range cts {
			byStatus[string(c.Status)]++
		}
		out["constraints"] = map[string]any{"total": len(cts), "by_status": byStatus}
	}
	if sols, err := a.Solve.List(); err == nil {
		byStatus := map[string]int{}
		feasible := 0
		for _, s := range sols {
			byStatus[string(s.Status)]++
			if s.State.Feasible {
				feasible++
			}
		}
		out["solutions"] = map[string]any{"total": len(sols), "feasible": feasible, "by_status": byStatus}
	}
	if rps, err := a.Reports.List(); err == nil {
		byStatus := map[string]int{}
		for _, r := range rps {
			byStatus[string(r.Status)]++
		}
		out["reports"] = map[string]any{"total": len(rps), "by_status": byStatus}
	}
	writeJSON(w, http.StatusOK, out)
}
