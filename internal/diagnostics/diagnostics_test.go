package diagnostics

import (
	"strings"
	"testing"

	"task177-isomix/internal/model"
)

func TestAnalyzeHighlightsConflictAndHumanizes(t *testing.T) {
	solution := &model.Solution{ID: "so1", Status: model.SolutionInfeasible, State: model.SolutionState{MatrixSummary: model.MatrixSummary{Rows: 4, Endmembers: 2, NonZeros: 4}, ConflictCore: []string{"c1"}}}
	report := Analyze(solution)
	if report.Risk != "constraint-risk" || !strings.Contains(Humanize(report), "CONSTRAINT-RISK") {
		t.Fatalf("unexpected diagnostic: %#v", report)
	}
}

func TestAnalyzeReportsZeroDensityForEmptyMatrix(t *testing.T) {
	// 空矩阵（无行或无变量）必须报告零密度，并落入稳定的低信息量诊断。
	for _, summary := range []model.MatrixSummary{
		{Rows: 0, Endmembers: 3, NonZeros: 0}, // 无约束行
		{Rows: 3, Endmembers: 0, NonZeros: 0}, // 无变量
		{Rows: 0, Endmembers: 0, NonZeros: 0}, // 完全空
	} {
		solution := &model.Solution{ID: "so2", Status: model.SolutionFeasible, State: model.SolutionState{Feasible: true, MatrixSummary: summary}}
		report := Analyze(solution)
		if report.Density != 0 {
			t.Fatalf("empty matrix %#v density = %v, want 0", summary, report.Density)
		}
		if report.Risk != "sparse-system" {
			t.Fatalf("empty matrix %#v risk = %q, want sparse-system", summary, report.Risk)
		}
		if !strings.Contains(Humanize(report), "density 0.0000") {
			t.Fatalf("empty matrix humanized = %q, want density 0.0000", Humanize(report))
		}
	}
}
