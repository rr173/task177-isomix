// Package report 管理求解报告：冻结端元/约束版本快照、发布、替代与差异比较。
// 已发布报告不可修改；端元/约束修订只触发新求解与新报告，不回写旧报告。
package report

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"task177-isomix/internal/constraint"
	"task177-isomix/internal/endmember"
	"task177-isomix/internal/model"
	"task177-isomix/internal/solve"
)

// IDGen 生成报告 ID。
type IDGen func() string

// TimeNow 返回当前时间。
type TimeNow func() time.Time

// Store 报告持久化接口。
type Store interface {
	CreateReport(r *model.Report) error
	UpdateReport(r *model.Report) error
	GetReport(id string) (*model.Report, error)
	ListReports() ([]model.Report, error)
	CountReports() (int, error)
}

// Service 报告领域服务。
type Service struct {
	store   Store
	solv    *solve.Service
	endm    *endmember.Service
	cons    *constraint.Service
	nextID  IDGen
	now     TimeNow
}

// NewService 构造报告服务。
func NewService(st Store, sv *solve.Service, em *endmember.Service, cs *constraint.Service, idg IDGen, now TimeNow) *Service {
	if idg == nil {
		idg = func() string { return "rp_" + strings.ToLower(model.RandToken(8)) }
	}
	if now == nil {
		now = time.Now
	}
	return &Service{store: st, solv: sv, endm: em, cons: cs, nextID: idg, now: now}
}

// Create 基于已接受（可行）求解创建报告草案，冻结全部端元与约束版本快照。
func (s *Service) Create(solutionID, title string) (*model.Report, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, model.NewError("BAD_REQUEST", "report title is required")
	}
	sol, err := s.solv.Get(solutionID)
	if err != nil {
		return nil, err
	}
	if sol.Status != model.SolutionFeasible {
		return nil, model.NewError("BAD_REQUEST", "only feasible solutions can be published as reports, current=%s", sol.Status)
	}

	r := &model.Report{
		ID:         s.nextID(),
		Title:      title,
		Status:     model.ReportDraft,
		SolutionID: sol.ID,
		SampleID:   sol.SampleID,
		InputHash:  sol.InputHash,
		State:      sol.State,
		CreatedAt:  s.now(),
	}
	// 冻结端元快照。
	for _, id := range sol.EndmemberIDs {
		e, err := s.endm.Get(id)
		if err != nil {
			return nil, err
		}
		r.Endmembers = append(r.Endmembers, e.Snapshot())
	}
	// 冻结约束快照。
	for _, id := range sol.ConstraintIDs {
		c, err := s.cons.Get(id)
		if err != nil {
			return nil, err
		}
		r.Constraints = append(r.Constraints, c.Snapshot())
	}
	sort.Slice(r.Endmembers, func(i, j int) bool { return r.Endmembers[i].ID < r.Endmembers[j].ID })
	sort.Slice(r.Constraints, func(i, j int) bool { return r.Constraints[i].ID < r.Constraints[j].ID })

	if err := s.store.CreateReport(r); err != nil {
		return nil, err
	}
	return r, nil
}

// Get 按 ID 查询报告。
func (s *Service) Get(id string) (*model.Report, error) {
	return s.store.GetReport(id)
}

// List 列出全部报告（按创建时间倒序）。
func (s *Service) List() ([]model.Report, error) {
	all, err := s.store.ListReports()
	if err != nil {
		return nil, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })
	return all, nil
}

// Publish 发布报告：草案 -> 已发布，冻结时间戳。
func (s *Service) Publish(id string) (*model.Report, error) {
	r, err := s.store.GetReport(id)
	if err != nil {
		return nil, err
	}
	if r.Status == model.ReportPublished {
		return r, nil
	}
	if r.Status != model.ReportDraft {
		return nil, model.NewError("INVALID_TRANSITION", "only draft reports can be published, current=%s", r.Status)
	}
	r.Status = model.ReportPublished
	now := s.now()
	r.PublishedAt = &now
	if err := s.store.UpdateReport(r); err != nil {
		return nil, err
	}
	return r, nil
}

// CheckSuperseded 检查报告快照是否已被输入修订淘汰：任一端元或约束当前版本
// 与快照版本不一致即需替代。返回 (需替代, 变更列表)。
func (s *Service) CheckSuperseded(r *model.Report) (bool, []string) {
	var changes []string
	for _, snap := range r.Endmembers {
		cur, err := s.endm.Get(snap.ID)
		if err == nil && cur.Version != snap.Version {
			changes = append(changes, "endmember "+snap.ID+" v"+strconv.Itoa(snap.Version)+"->v"+strconv.Itoa(cur.Version))
		}
	}
	for _, snap := range r.Constraints {
		cur, err := s.cons.Get(snap.ID)
		if err == nil && cur.Version != snap.Version {
			changes = append(changes, "constraint "+snap.ID+" v"+strconv.Itoa(snap.Version)+"->v"+strconv.Itoa(cur.Version))
		}
	}
	sort.Strings(changes)
	return len(changes) > 0, changes
}

// SupersedeStale 把已发布但输入已修订的报告标记为已替代（幂等，用于收尾维护）。
func (s *Service) SupersedeStale() (superseded int, err error) {
	all, err := s.store.ListReports()
	if err != nil {
		return 0, err
	}
	for i := range all {
		r := &all[i]
		if r.Status != model.ReportPublished {
			continue
		}
		stale, _ := s.CheckSuperseded(r)
		if stale {
			r.Status = model.ReportSuperseded
			if err := s.store.UpdateReport(r); err != nil {
				return superseded, err
			}
			superseded++
		}
	}
	return superseded, nil
}

// Diff 比较两份报告快照差异：端元/约束版本、可行域端点。
func (s *Service) Diff(a, b *model.Report) map[string]any {
	out := map[string]any{}
	var emChanges []string
	am := map[string]model.EndmemberSnapshot{}
	bm := map[string]model.EndmemberSnapshot{}
	for _, e := range a.Endmembers {
		am[e.ID] = e
	}
	for _, e := range b.Endmembers {
		bm[e.ID] = e
	}
	for id, ea := range am {
		if eb, ok := bm[id]; ok {
			if ea.Version != eb.Version {
				emChanges = append(emChanges, id+" v"+strconv.Itoa(ea.Version)+"->v"+strconv.Itoa(eb.Version))
			}
		} else {
			emChanges = append(emChanges, id+" removed")
		}
	}
	for id := range bm {
		if _, ok := am[id]; !ok {
			emChanges = append(emChanges, id+" added")
		}
	}
	sort.Strings(emChanges)
	out["endmember_changes"] = emChanges

	var ctChanges []string
	ac := map[string]model.ConstraintSnapshot{}
	bc := map[string]model.ConstraintSnapshot{}
	for _, c := range a.Constraints {
		ac[c.ID] = c
	}
	for _, c := range b.Constraints {
		bc[c.ID] = c
	}
	for id, ca := range ac {
		if cb, ok := bc[id]; ok {
			if ca.Version != cb.Version {
				ctChanges = append(ctChanges, id+" v"+strconv.Itoa(ca.Version)+"->v"+strconv.Itoa(cb.Version))
			}
		} else {
			ctChanges = append(ctChanges, id+" removed")
		}
	}
	for id := range bc {
		if _, ok := ac[id]; !ok {
			ctChanges = append(ctChanges, id+" added")
		}
	}
	sort.Strings(ctChanges)
	out["constraint_changes"] = ctChanges

	out["feasible"] = map[string]bool{"a": a.State.Feasible, "b": b.State.Feasible}
	if a.State.Feasible && b.State.Feasible {
		out["bounds_changed"] = boundsChanged(a.State.EndmemberBounds, b.State.EndmemberBounds)
	}
	return out
}

// boundsChanged 比较两报告各端元比例区间是否不同。
func boundsChanged(a, b map[string]model.Range) bool {
	if len(a) != len(b) {
		return true
	}
	for k, ra := range a {
		rb, ok := b[k]
		if !ok {
			return true
		}
		if ra.Lo != rb.Lo || ra.Hi != rb.Hi {
			return true
		}
	}
	return false
}

// Summary 输出报告服务摘要（供自检/统计 API）。
func (s *Service) Summary() (map[string]any, error) {
	n, err := s.store.CountReports()
	if err != nil {
		return nil, err
	}
	return map[string]any{"reports": n}, nil
}

