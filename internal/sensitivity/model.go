package sensitivity

import (
	"time"

	"task177-isomix/internal/calibration"
	"task177-isomix/internal/scenario"
)

// BoundSummary describes the uncertainty carried by one solved proportion.
type BoundSummary struct {
	EndmemberID string  `json:"endmember_id"`
	Low         float64 `json:"low"`
	High        float64 `json:"high"`
	Center      float64 `json:"center"`
	Width       float64 `json:"width"`
	Relative    float64 `json:"relative_width"`
}

// Report is a deterministic sensitivity explanation for a solution.
type Report struct {
	SolutionID           string              `json:"solution_id"`
	Feasible             bool                `json:"feasible"`
	TotalWidth           float64             `json:"total_width"`
	MeanRelative         float64             `json:"mean_relative_width"`
	Stability            string              `json:"stability"`
	Bounds               []BoundSummary      `json:"bounds"`
	Recommendations      []string            `json:"recommendations,omitempty"`
	Calibration          calibration.Summary `json:"calibration"`
	Scenarios            scenario.Summary    `json:"scenarios"`
	ScenarioDistribution map[string]float64  `json:"scenario_distribution,omitempty"`
	ScenarioRMS          float64             `json:"scenario_rms"`
	ScenarioDigest       string              `json:"scenario_digest"`
	ScenarioAggregate    map[string]float64  `json:"scenario_aggregate"`
	ScenarioThresholds   map[string]int      `json:"scenario_thresholds"`
	GeneratedAt          time.Time           `json:"generated_at"`
}
