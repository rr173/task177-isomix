package solve

import (
	"errors"
	"testing"

	"task177-isomix/internal/constraint"
	"task177-isomix/internal/endmember"
	"task177-isomix/internal/measure"
	"task177-isomix/internal/model"
	"task177-isomix/internal/store"
)

// newServices 构造一组基于内存 SQLite 的真实领域服务，供 solve.Submit 端到端验证。
func newServices(t *testing.T) (*store.Store, *endmember.Service, *measure.Service, *constraint.Service, *Service) {
	t.Helper()
	st, err := store.Open("")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	em := endmember.NewService(st.EndmemberStore, func() string { return "em_" + model.RandToken(4) }, nil)
	ms := measure.NewService(st.SampleStore, func() string { return "sp_" + model.RandToken(4) }, nil)
	cs := constraint.NewService(st.ConstraintStore, func() string { return "cs_" + model.RandToken(4) }, nil)
	sv := NewService(st.SolutionStore, em, ms, cs, func() string { return "so_" + model.RandToken(6) }, nil)
	return st, em, ms, cs, sv
}

func mustCreateEM(t *testing.T, em *endmember.Service, name string, comps map[string]model.Range) *model.Endmember {
	t.Helper()
	e, err := em.Create(name, comps, nil)
	if err != nil {
		t.Fatalf("create endmember %s: %v", name, err)
	}
	if _, err := em.Validate(e.ID); err != nil {
		t.Fatalf("validate %s: %v", name, err)
	}
	if _, err := em.MakeAvailable(e.ID); err != nil {
		t.Fatalf("make available %s: %v", name, err)
	}
	return e
}

func mustCreateSample(t *testing.T, ms *measure.Service, name string, meas map[string]model.Range) *model.Sample {
	t.Helper()
	sp, err := ms.Create(name, meas, nil, nil)
	if err != nil {
		t.Fatalf("create sample %s: %v", name, err)
	}
	if _, err := ms.Check(sp.ID); err != nil {
		t.Fatalf("check sample %s: %v", name, err)
	}
	return sp
}

func isDimMismatch(err error) bool {
	var de *model.DomainError
	return errors.As(err, &de) && de.Code == "DIMENSION_MISMATCH"
}

// 基数相同但同位素名不同的端元集合必须被拒绝。
func TestSubmitRejectsSameCardinalityDifferentNames(t *testing.T) {
	_, em, ms, _, sv := newServices(t)
	mustCreateEM(t, em, "mantle", map[string]model.Range{
		"d18O": {Lo: 5.2, Hi: 5.8}, "d2H": {Lo: -90, Hi: -70},
	})
	mustCreateEM(t, em, "crust", map[string]model.Range{
		"d13C": {Lo: -5, Hi: -1}, "d34S": {Lo: 5, Hi: 15}, // 同基数 2，名字完全不同
	})
	sp := mustCreateSample(t, ms, "river", map[string]model.Range{
		"d18O": {Lo: 6.0, Hi: 8.0}, "d2H": {Lo: -80, Hi: -55},
	})
	if _, err := sv.Submit(sp.ID); err == nil || !isDimMismatch(err) {
		t.Fatalf("Submit must reject endmembers with same cardinality but different isotope names, got: %v", err)
	}
}

// 样品同位素名与端元不同（基数相同）必须被拒绝。
func TestSubmitRejectsEndmemberSampleNameMismatch(t *testing.T) {
	_, em, ms, _, sv := newServices(t)
	mustCreateEM(t, em, "mantle", map[string]model.Range{
		"d18O": {Lo: 5.2, Hi: 5.8}, "d2H": {Lo: -90, Hi: -70},
	})
	mustCreateEM(t, em, "crust", map[string]model.Range{
		"d18O": {Lo: 8.0, Hi: 9.5}, "d2H": {Lo: -60, Hi: -40},
	})
	sp := mustCreateSample(t, ms, "river", map[string]model.Range{
		"d13C": {Lo: -3, Hi: 0}, "d34S": {Lo: 6, Hi: 12}, // 基数同为 2，名字与端元不同
	})
	if _, err := sv.Submit(sp.ID); err == nil || !isDimMismatch(err) {
		t.Fatalf("Submit must reject when endmember and sample isotope names differ, got: %v", err)
	}
}

// 端元有的同位素样品没有 -> 必须拒绝（这是核心修复点：之前只单向检查）。
func TestSubmitRejectsEndmemberExtraIsotope(t *testing.T) {
	_, em, ms, _, sv := newServices(t)
	mustCreateEM(t, em, "mantle", map[string]model.Range{
		"d18O": {Lo: 5.2, Hi: 5.8}, "d2H": {Lo: -90, Hi: -70},
	})
	mustCreateEM(t, em, "crust", map[string]model.Range{
		"d18O": {Lo: 8.0, Hi: 9.5}, "d2H": {Lo: -60, Hi: -40},
	})
	// 样品只测了 d18O：端元的 d2H 没有对应样品测量 -> 必须拒绝。
	sp := mustCreateSample(t, ms, "river", map[string]model.Range{
		"d18O": {Lo: 6.0, Hi: 8.0},
	})
	if _, err := sv.Submit(sp.ID); err == nil || !isDimMismatch(err) {
		t.Fatalf("Submit must reject when endmembers have isotopes absent from sample, got: %v", err)
	}
}

// 维度名完全一致时正常求解（回归基线，确认修复未误伤合法求解）。
func TestSubmitAcceptsMatchingDimensions(t *testing.T) {
	_, em, ms, _, sv := newServices(t)
	mustCreateEM(t, em, "mantle", map[string]model.Range{
		"d18O": {Lo: 5.2, Hi: 5.8}, "d2H": {Lo: -90, Hi: -70},
	})
	mustCreateEM(t, em, "crust", map[string]model.Range{
		"d18O": {Lo: 8.0, Hi: 9.5}, "d2H": {Lo: -60, Hi: -40},
	})
	mustCreateEM(t, em, "organic", map[string]model.Range{
		"d18O": {Lo: 14.0, Hi: 18.0}, "d2H": {Lo: -160, Hi: -110},
	})
	sp := mustCreateSample(t, ms, "river", map[string]model.Range{
		"d18O": {Lo: 6.0, Hi: 8.0}, "d2H": {Lo: -80, Hi: -55},
	})
	sol, err := sv.Submit(sp.ID)
	if err != nil {
		t.Fatalf("Submit should succeed when dimensions match, got: %v", err)
	}
	if sol.Status != model.SolutionFeasible {
		t.Fatalf("expected feasible, got %s", sol.Status)
	}
}
