package scenario

import (
	"time"

	"task177-isomix/internal/model"
)

func baseline(bounds map[string]model.Range) map[string][2]float64 {
	out := make(map[string][2]float64, len(bounds))
	for id, bound := range bounds {
		out[id] = [2]float64{bound.Lo, bound.Hi}
	}
	return out
}

// EvaluateStandard runs two standard scenarios used by research reports:
// conservative intervals shrink measurement uncertainty, while expansive
// intervals model the effect of a less precise field sample.
func EvaluateStandard(bounds map[string]model.Range, now func() time.Time) Summary {
	if now == nil {
		now = time.Now
	}
	conservative := Compare(bounds, Apply(bounds, Request{Name: "conservative", WidthScale: .75}))
	expansive := Compare(bounds, Apply(bounds, Request{Name: "expansive", WidthScale: 1.25}))
	all := append(append([]Projection(nil), expansive...), conservative...)
	return Summary{Baseline: baseline(bounds), Conservative: conservative, Expansive: expansive, MostSensitive: MostSensitive(all), GeneratedAt: now().UTC()}
}

// Evaluate runs a caller-provided scenario and returns only the projection rows.
func Evaluate(bounds map[string]model.Range, request Request) []Projection {
	return Compare(bounds, Apply(bounds, request))
}

// CenterOfMass computes the midpoint weighted by interval width.
func CenterOfMass(bounds map[string]model.Range) float64 {
	var numerator, denominator float64
	for _, bound := range bounds {
		w := width(bound)
		numerator += midpoint(bound) * w
		denominator += w
	}
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}
