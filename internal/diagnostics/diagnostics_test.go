package diagnostics

import (
	"strings"
	"testing"

	"task177-isomix/internal/model"
)

func TestAnalyzeHighlightsConflictAndHumanizes(t *testing.T) {
	solution := &model.Solution{ID: "so1", Status: model.SolutionInfeasible, State: model.SolutionState{MatrixSummary: model.MatrixSummary{Rows: 4, Endmembers: 2, NonZeros: 4}, ConflictCore: []string{"c1"}}}
	report := Analyze(solution)
	if report.Risk != "constraint-risk" || !strings.Contains(Humanize(report), "constraint-risk") {
		t.Fatalf("unexpected diagnostic: %#v", report)
	}
}
