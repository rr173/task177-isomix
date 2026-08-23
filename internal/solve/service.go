// Package solve 编排混合求解：校验输入、构建矩阵、执行 LP、
// 计算可行域/活跃约束/最小不可满足子集，并维护求解生命周期与重启恢复。
package solve

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"task177-isomix/internal/constraint"
	"task177-isomix/internal/endmember"
	"task177-isomix/internal/hashutil"
	"task177-isomix/internal/lp"
	"task177-isomix/internal/measure"
	"task177-isomix/internal/model"
)

// IDGen 生成求解 ID。
type IDGen func() string

// TimeNow 返回当前时间。
type TimeNow func() time.Time

// Store 求解持久化接口。
type Store interface {
	CreateSolution(s *model.Solution) error
	UpdateSolution(s *model.Solution) error
	GetSolution(id string) (*model.Solution, error)
	ListSolutions() ([]model.Solution, error)
	GetSolutionByInputHash(h string) (*model.Solution, error)
	CountSolutions() (int, error)
}

// Service 求解领域服务。
type Service struct {
	solutions Store
	endm      *endmember.Service
	meas      *measure.Service
	cons      *constraint.Service
	nextID    IDGen
	now       TimeNow
}

// NewService 构造求解服务。
func NewService(st Store, em *endmember.Service, ms *measure.Service, cs *constraint.Service, idg IDGen, now TimeNow) *Service {
	if idg == nil {
		idg = func() string { return "so_" + strings.ToLower(model.RandToken(10)) }
	}
	if now == nil {
		now = time.Now
	}
	return &Service{solutions: st, endm: em, meas: ms, cons: cs, nextID: idg, now: now}
}

// Submit 提交样品求解。幂等：相同输入哈希只返回已有结果（复用活动任务）。
func (s *Service) Submit(sampleID string) (*model.Solution, error) {
	sp, err := s.meas.Get(sampleID)
	if err != nil {
		return nil, err
	}
	if sp.Status == model.SampleSealed {
		return nil, model.ErrSealedSample
	}

	ems, err := s.endm.ListAvailable()
	if err != nil {
		return nil, err
	}
	if len(ems) == 0 {
		return nil, model.ErrNoAvailableEndmember
	}
	// 维度一致性检查：端元内部同位素名集合必须一致。
	if !endmemberConsistent(ems) {
		return nil, model.NewError("DIMENSION_MISMATCH", "available endmembers have inconsistent isotope dimensions")
	}
	// 端元与样品必须共享完全一致的同位素名集合（精确比较，而非仅比基数）：
	// 缺失 = 样品有端元无（无法建模），多余 = 端元有样品无（约束无法落地），
	// 任一存在均无法构建有效的质量守恒混合系统。
	missing, extra := checkSampleDimension(ems, *sp)
	if len(missing) > 0 || len(extra) > 0 {
		return nil, model.NewError("DIMENSION_MISMATCH",
			"sample and endmembers have inconsistent isotope dimensions: missing=[%s] extra=[%s]",
			strings.Join(missing, ","), strings.Join(extra, ","))
	}

	cs, err := s.cons.ListActive()
	if err != nil {
		return nil, err
	}

	// 预校验：活动约束引用的端元必须都在候选集中（避免编译期未知端元）。
	if err := validateConstraintTargets(ems, cs); err != nil {
		return nil, err
	}

	// 输入哈希幂等。
	h, err := hashutil.InputHash(*sp, ems, cs)
	if err != nil {
		return nil, err
	}
	if existing, err := s.solutions.GetSolutionByInputHash(h); err == nil && existing != nil {
		return existing, nil
	}

	ids := make([]string, len(ems))
	for i, e := range ems {
		ids[i] = e.ID
	}
	cids := make([]string, len(cs))
	for i, c := range cs {
		cids[i] = c.ID
	}

	sol := &model.Solution{
		ID:           s.nextID(),
		SampleID:     sampleID,
		Status:       model.SolutionQueued,
		InputHash:    h,
		EndmemberIDs: ids,
		ConstraintIDs: cids,
		CreatedAt:    s.now(),
	}
	if err := s.solutions.CreateSolution(sol); err != nil {
		return nil, err
	}

	// 同步执行求解（状态在内存中推进，持久化最终态）。
	s.execute(sol)
	if err := s.solutions.UpdateSolution(sol); err != nil {
		return nil, err
	}
	return sol, nil
}

// validateConstraintTargets 校验活动约束引用的端元均在候选集中。
func validateConstraintTargets(ems []model.Endmember, cs []model.Constraint) error {
	has := func(id string) bool {
		for _, e := range ems {
			if e.ID == id {
				return true
			}
		}
		return false
	}
	for _, c := range cs {
		switch c.Type {
		case model.ConstraintExclude, model.ConstraintBound:
			if !has(c.Target) {
				return model.NewError("BAD_REQUEST", "constraint %q references unknown endmember %q (not in candidate set)", c.Name, c.Target)
			}
		case model.ConstraintRatio:
			parts := strings.Split(c.Target, ":")
			if len(parts) != 2 || !has(strings.TrimSpace(parts[0])) || !has(strings.TrimSpace(parts[1])) {
				return model.NewError("BAD_REQUEST", "constraint %q references unknown endmember in target %q (not in candidate set)", c.Name, c.Target)
			}
		}
	}
	return nil
}

// execute 执行一次求解，填充 sol.State 与终态状态。
func (s *Service) execute(sol *model.Solution) {
	sp, err := s.meas.Get(sol.SampleID)
	if err != nil {
		sol.Status = model.SolutionInfeasible
		sol.Error = "sample vanished: " + err.Error()
		return
	}
	ems := make([]model.Endmember, 0, len(sol.EndmemberIDs))
	for _, id := range sol.EndmemberIDs {
		e, err := s.endm.Get(id)
		if err != nil {
			sol.Status = model.SolutionUnstable
			sol.Error = "endmember lookup failed: " + err.Error()
			return
		}
		ems = append(ems, *e)
	}
	cs := make([]model.Constraint, 0, len(sol.ConstraintIDs))
	for _, id := range sol.ConstraintIDs {
		c, err := s.cons.Get(id)
		if err != nil {
			sol.Status = model.SolutionUnstable
			sol.Error = "constraint lookup failed: " + err.Error()
			return
		}
		cs = append(cs, *c)
	}

	sol.Status = model.SolutionRunning
	compiled, err := s.compileAndSolve(ems, *sp, cs)
	if err != nil {
		sol.Status = model.SolutionUnstable
		sol.Error = err.Error()
		return
	}
	sol.State = *compiled
	sol.Status = model.SolutionFeasible
	if !compiled.Feasible {
		if compiled.UnstableEvidence != "" {
			sol.Status = model.SolutionUnstable
		} else {
			sol.Status = model.SolutionInfeasible
		}
	}
	now := s.now()
	sol.FinishedAt = &now
	if sol.Status == model.SolutionFeasible {
		_ = s.meas.MarkSolved(sol.SampleID)
	}
}

// compileAndSolve 构建矩阵并求解，返回完整求解状态。
func (s *Service) compileAndSolve(ems []model.Endmember, sp model.Sample, cs []model.Constraint) (*model.SolutionState, error) {
	b, err := constraint.NewBuilder(ems, sp)
	if err != nil {
		return nil, err
	}
	compiled, err := BuildProblem(b, cs)
	if err != nil {
		return nil, err
	}

	summary := model.MatrixSummary{
		Endmembers: b.NumVars(),
		Isotopes:   len(b.IsotopeNames),
		Rows:       len(compiled.Problem.A),
		NonZeros:   NonZeros(compiled.Problem),
	}

	// 可行性检查。
	ok, res := lp.Feasible(compiled.Problem)
	if res.Status == lp.StatusUnstable {
		return &model.SolutionState{
			Feasible:         false,
			MatrixSummary:    summary,
			UnstableEvidence: "feasibility check: " + res.Evidence,
		}, nil
	}
	if !ok {
		// 不可行：计算最小不可满足子集。
		core := s.minimalConflictCore(compiled)
		return &model.SolutionState{
			Feasible:      false,
			MatrixSummary: summary,
			ConflictCore:  core,
		}, nil
	}

	// 可行：计算各端元比例可行区间。
	lows, highs, unstable, evidence := lp.Bounds(compiled.Problem)
	if unstable {
		return &model.SolutionState{
			Feasible:         false,
			MatrixSummary:    summary,
			UnstableEvidence: "bounds computation: " + evidence,
		}, nil
	}
	// 活跃约束：用第一个可行解（目标 0 的解）。
	sol := lp.Solution(compiled.Problem)
	if sol.Status != lp.StatusOptimal {
		return &model.SolutionState{
			Feasible:         false,
			MatrixSummary:    summary,
			UnstableEvidence: "solution extraction: " + sol.Evidence,
		}, nil
	}
	active, _ := lp.ActiveRows(compiled.Problem, sol.X)
	activeIDs := map[string]bool{}
	var activeList []string
	for i, a := range active {
		if a && compiled.Rows[i].Origin != "" {
			if !activeIDs[compiled.Rows[i].Origin] {
				activeIDs[compiled.Rows[i].Origin] = true
				activeList = append(activeList, compiled.Rows[i].Origin)
			}
		}
	}
	sort.Strings(activeList)

	bounds := map[string]model.Range{}
	for i, eid := range b.EndmemberIDs {
		bounds[eid] = model.Range{Lo: lows[i], Hi: highs[i]}
	}

	return &model.SolutionState{
		Feasible:          true,
		EndmemberBounds:   bounds,
		Residual:          0,
		ActiveConstraints: activeList,
		MatrixSummary:     summary,
	}, nil
}

// minimalConflictCore 删除式最小不可满足子集。
// 前置条件：整体不可行。质量守恒行（公理）不参与删除。
func (s *Service) minimalConflictCore(compiled *Compiled) []string {
	m := len(compiled.Problem.A)
	// 行索引按 origin 分组：只有 origin != ""（用户约束）的行参与删除候选；
	// 同位素区间行没有 origin，也参与（它们是数据驱动约束）。
	removed := make([]bool, m)
	// 删除式扫描：对每一行尝试移除，若剩余仍不可行则该行不在核心（保留移除）。
	// 行顺序保持稳定。
	for i := 0; i < m; i++ {
		// 跳过已移除行。
		if removed[i] {
			continue
		}
		removed[i] = true
		sub := subProblem(compiled, removed)
		ok, res := lp.Feasible(sub)
		if res.Status == lp.StatusUnstable {
			// 数值不稳定：保守起见恢复该行（不轻率断定核心）。
			removed[i] = false
			continue
		}
		if !ok {
			// 移除后仍不可行 -> 该行不在核心。
			continue
		}
		// 移除后变可行 -> 该行在核心。
		removed[i] = false
	}

	seen := map[string]bool{}
	var core []string
	for i, r := range compiled.Rows {
		if !removed[i] {
			if r.Origin != "" && !seen[r.Origin] {
				seen[r.Origin] = true
				core = append(core, r.Origin)
			}
		}
	}
	sort.Strings(core)
	return core
}

// subProblem 根据 removed 标记构建子问题。
func subProblem(c *Compiled, removed []bool) *lp.Problem {
	p := c.Problem
	var A [][]float64
	var B []float64
	for i := 0; i < len(p.A); i++ {
		if removed[i] {
			continue
		}
		A = append(A, p.A[i])
		B = append(B, p.B[i])
	}
	return &lp.Problem{A: A, B: B, C: make([]float64, len(p.C))}
}

// Get 按 ID 查询求解。
func (s *Service) Get(id string) (*model.Solution, error) {
	return s.solutions.GetSolution(id)
}

// List 列出全部求解（按创建时间倒序）。
func (s *Service) List() ([]model.Solution, error) {
	all, err := s.solutions.ListSolutions()
	if err != nil {
		return nil, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })
	return all, nil
}

// endmemberConsistent 检查端元内部维度一致。
func endmemberConsistent(ems []model.Endmember) bool {
	if len(ems) == 0 {
		return true
	}
	base := keySet(ems[0].Components)
	for _, e := range ems[1:] {
		if !sameKeySet(base, keySet(e.Components)) {
			return false
		}
	}
	return true
}

// checkSampleDimension 检查端元与样品是否共享完全一致的同位素名集合。
// 精确比较同位素名而非仅比基数：返回缺失（样品有端元无）与多余（端元有样品无）列表，
// 任一非空即说明维度不一致，调用方据此拒绝求解。
func checkSampleDimension(ems []model.Endmember, sp model.Sample) (missing, extra []string) {
	endmembers := map[string]bool{}
	for _, e := range ems {
		for k := range e.Components {
			endmembers[k] = true
		}
	}
	for k := range sp.Measurements {
		if !endmembers[k] {
			missing = append(missing, k)
		}
	}
	for k := range endmembers {
		if _, ok := sp.Measurements[k]; !ok {
			extra = append(extra, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}

func keySet(m map[string]model.Range) map[string]bool {
	out := make(map[string]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}

func sameKeySet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// Summary 输出求解服务摘要（供自检/统计 API）。
func (s *Service) Summary() (map[string]any, error) {
	n, err := s.solutions.CountSolutions()
	if err != nil {
		return nil, err
	}
	return map[string]any{"solutions": n}, nil
}

var _ = fmt.Sprintf
