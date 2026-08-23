package sensitivity

import (
	"testing"
	"time"

	"task177-isomix/internal/model"
)

type solutionStore struct{ value *model.Solution }

func (s solutionStore) Get(string) (*model.Solution, error) { return s.value, nil }

func TestAnalyzeClassifiesStableBounds(t *testing.T) {
	solution := &model.Solution{
		ID:     "so1",
		Status: model.SolutionFeasible,
		State: model.SolutionState{
			Feasible: true,
			EndmemberBounds: map[string]model.Range{
				"a": {Lo: .4, Hi: .5},
				"b": {Lo: .5, Hi: .6},
			},
		},
	}
	service := NewService(solutionStore{value: solution}, func() time.Time { return time.Unix(0, 0) })
	report, err := service.Analyze("so1")
	if err != nil || report.Stability != "stable" || len(report.Bounds) != 2 {
		t.Fatalf("unexpected sensitivity report: %#v, %v", report, err)
	}
	if len(Explain(report)) != 2 {
		t.Fatalf("expected explanations for both bounds")
	}
}
