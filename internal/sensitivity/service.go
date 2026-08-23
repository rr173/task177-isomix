// Package sensitivity explains how much each source proportion can move while
// preserving the feasible solution returned by the LP solver.
package sensitivity

import (
	"time"

	"task177-isomix/internal/calibration"
	"task177-isomix/internal/model"
	"task177-isomix/internal/scenario"
)

// SolutionReader is the small part of solve.Service needed by this package.
type SolutionReader interface {
	Get(string) (*model.Solution, error)
}

// Service computes read-only sensitivity reports from persisted solutions.
type Service struct {
	solutions SolutionReader
	now       func() time.Time
}

func NewService(solutions SolutionReader, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{solutions: solutions, now: now}
}

// Analyze returns a stable explanation and never mutates the solution record.
func (s *Service) Analyze(solutionID string) (*Report, error) {
	sol, err := s.solutions.Get(solutionID)
	if err != nil {
		return nil, err
	}
	bounds, total, relative := summarizeBounds(sol.State.EndmemberBounds)
	report := &Report{
		SolutionID:   sol.ID,
		Feasible:     sol.State.Feasible,
		TotalWidth:   total,
		MeanRelative: relative,
		Stability:    classify(total, relative, sol.State.Feasible),
		Bounds:       bounds,
		GeneratedAt:  s.now().UTC(),
	}
	report.Calibration = calibration.Assess(sol.State.EndmemberBounds, s.now)
	report.Scenarios = scenario.EvaluateStandard(sol.State.EndmemberBounds, s.now)
	rows := append(append([]scenario.Projection(nil), report.Scenarios.Conservative...), report.Scenarios.Expansive...)
	report.ScenarioDistribution = scenario.WidthDistribution(rows)
	report.ScenarioRMS = scenario.RootMeanSquare(rows)
	report.ScenarioDigest = scenario.PlanDigest(scenario.DefaultPlan())
	aggregate := scenario.SummarizeRows(scenario.NormalizeRows(rows))
	report.ScenarioAggregate = map[string]float64{"wider": float64(aggregate.Wider), "tighter": float64(aggregate.Tighter), "unchanged": float64(aggregate.Unchanged), "average_move": aggregate.AverageMove, "largest_move": aggregate.LargestMove}
	thresholds := scenario.CountThresholds(rows)
	report.ScenarioThresholds = map[string]int{"small": thresholds.Small, "moderate": thresholds.Moderate, "large": thresholds.Large, "extreme": thresholds.Extreme}
	if !sol.State.Feasible {
		report.Recommendations = append(report.Recommendations, "inspect the conflict core before comparing source proportions")
	} else if report.Stability == "diffuse" {
		report.Recommendations = append(report.Recommendations, "add an independent isotope or a conservative source bound")
	} else if report.Stability == "sensitive" {
		report.Recommendations = append(report.Recommendations, "record a narrower measurement interval before publishing")
	}
	return report, nil
}
