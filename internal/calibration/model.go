// Package calibration turns a solved interval into comparable confidence
// scores. It is deliberately deterministic: no random sampling or global
// state is involved, so a published explanation can be reproduced later.
package calibration

import "time"

// Point is one source proportion and its calibrated confidence.
type Point struct {
	ID         string  `json:"id"`
	Center     float64 `json:"center"`
	Width      float64 `json:"width"`
	Confidence float64 `json:"confidence"`
	Band       string  `json:"band"`
}

// Summary aggregates calibrated source proportions.
type Summary struct {
	Points         []Point   `json:"points"`
	Average        float64   `json:"average_confidence"`
	Lowest         float64   `json:"lowest_confidence"`
	Highest        float64   `json:"highest_confidence"`
	Spread         float64   `json:"confidence_spread"`
	Interpretation string    `json:"interpretation"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// CurveSample documents the breakpoints used by the confidence curve.
type CurveSample struct {
	RelativeWidth float64 `json:"relative_width"`
	Confidence    float64 `json:"confidence"`
	Band          string  `json:"band"`
}
