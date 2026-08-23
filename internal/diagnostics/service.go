// Package diagnostics provides read-only explanations for solver output.
package diagnostics

import "task177-isomix/internal/model"

// Analyze turns a persisted solution into a compact numerical health report.
func Analyze(solution *model.Solution) Report {
	if solution == nil {
		return Report{Risk: "missing"}
	}
	densityValue := density(solution.State.MatrixSummary)
	report := Report{
		SolutionID:     solution.ID,
		Status:         string(solution.Status),
		ConstraintRows: solution.State.MatrixSummary.Rows,
		Variables:      solution.State.MatrixSummary.Endmembers,
		Density:        densityValue,
		Residual:       solution.State.Residual,
		Risk:           risk(solution.State.Feasible, solution.State.Residual, densityValue, solution.Status),
	}
	if len(solution.State.ConflictCore) > 0 {
		report.Observations = append(report.Observations, "a minimal conflict core is attached to this result")
	}
	if len(solution.State.ActiveConstraints) > 0 {
		report.Observations = append(report.Observations, "at least one input constraint is active at the selected solution")
	}
	if solution.State.UnstableEvidence != "" {
		report.Observations = append(report.Observations, solution.State.UnstableEvidence)
	}
	return report
}
