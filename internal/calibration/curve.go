package calibration

import "math"

// Confidence maps relative interval width to [0,1] using a smooth curve.
func Confidence(relativeWidth float64) float64 {
	if math.IsNaN(relativeWidth) || math.IsInf(relativeWidth, 0) {
		return 1
	}
	if relativeWidth <= 0 {
		return 1
	}
	// The reciprocal curve keeps a small interval informative without making
	// a very wide interval look exactly like a hard failure.
	return 1 / (1 + relativeWidth*relativeWidth*4)
}

// Band names the confidence range presented to a researcher.
func Band(confidence float64) string {
	switch {
	case confidence >= .8:
		return "strong"
	case confidence >= .5:
		return "moderate"
	case confidence >= .2:
		return "weak"
	default:
		return "insufficient"
	}
}

// Curve returns fixed calibration landmarks for clients that draw a chart.
func Curve() []CurveSample {
	widths := []float64{0, .05, .1, .2, .35, .5, .75, 1, 2}
	out := make([]CurveSample, 0, len(widths))
	for _, width := range widths {
		confidence := Confidence(width)
		out = append(out, CurveSample{RelativeWidth: width, Confidence: confidence, Band: Band(confidence)})
	}
	return out
}
