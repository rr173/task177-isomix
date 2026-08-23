// Package scenario provides reproducible what-if projections over a solved
// feasible region. It is an analysis layer: projections never overwrite the
// persisted solution or report snapshot.
package scenario

import "time"

// Request describes an interval perturbation scenario.
type Request struct {
	Name        string  `json:"name"`
	WidthScale  float64 `json:"width_scale"`
	CenterShift float64 `json:"center_shift"`
	ClampLow    float64 `json:"clamp_low"`
	ClampHigh   float64 `json:"clamp_high"`
}

// Projection is the difference for one source under a scenario.
type Projection struct {
	ID            string  `json:"id"`
	BaselineLow   float64 `json:"baseline_low"`
	BaselineHigh  float64 `json:"baseline_high"`
	ProjectedLow  float64 `json:"projected_low"`
	ProjectedHigh float64 `json:"projected_high"`
	DeltaWidth    float64 `json:"delta_width"`
	Direction     string  `json:"direction"`
}

// Summary records a pair of standard what-if scenarios.
type Summary struct {
	Baseline      map[string][2]float64 `json:"baseline"`
	Conservative  []Projection          `json:"conservative"`
	Expansive     []Projection          `json:"expansive"`
	MostSensitive string                `json:"most_sensitive,omitempty"`
	GeneratedAt   time.Time             `json:"generated_at"`
}
