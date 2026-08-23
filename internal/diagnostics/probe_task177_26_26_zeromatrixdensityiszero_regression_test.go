package diagnostics

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug26_ZeroMatrixDensityIsZero(t *testing.T) {
	s := &model.Solution{ID: "s", State: model.SolutionState{MatrixSummary: model.MatrixSummary{}}}; if got := Analyze(s).Density; got != 0 { t.Fatalf("density=%v", got) }
}
