package diagnostics

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug27_ResidualRiskWinsOverSparse(t *testing.T) {
	s := &model.Solution{ID: "s", Status: model.SolutionFeasible, State: model.SolutionState{Feasible: true, Residual: 1, MatrixSummary: model.MatrixSummary{Rows: 10, Endmembers: 10, NonZeros: 1}}}; if Analyze(s).Risk != "residual-risk" { t.Fatalf("risk=%s", Analyze(s).Risk) }
}
