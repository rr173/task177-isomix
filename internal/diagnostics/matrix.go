package diagnostics

import "task177-isomix/internal/model"

func density(summary model.MatrixSummary) float64 {
	denominator := summary.Rows * summary.Endmembers
	if denominator <= 0 {
		return 0
	}
	return float64(summary.NonZeros) / float64(denominator)
}

func risk(feasible bool, residual, density float64, status model.SolutionStatus) string {
	if status == model.SolutionUnstable {
		return "numerical-risk"
	}
	if !feasible {
		return "constraint-risk"
	}
	if residual > 1e-7 {
		return "residual-risk"
	}
	if density < 0.2 {
		return "sparse-system"
	}
	return "normal"
}
