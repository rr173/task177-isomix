package scenario

import "math"

// SafeFraction clamps a ratio used by scenario summaries to [0,1].
func SafeFraction(numerator, denominator float64) float64 {
	if denominator == 0 || math.IsNaN(numerator) || math.IsNaN(denominator) || math.IsInf(numerator, 0) || math.IsInf(denominator, 0) {
		return 0
	}
	value := numerator / denominator
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

// Coverage calculates the fraction of rows that remain within a movement limit.
func Coverage(rows []Projection, limit float64) float64 {
	if limit < 0 || math.IsNaN(limit) {
		limit = 0
	}
	within := 0
	for _, row := range rows {
		if math.Abs(row.DeltaWidth) <= limit {
			within++
		}
	}
	return SafeFraction(float64(within), float64(len(rows)))
}

// DirectionBalance measures whether a scenario widens and tightens evenly.
// It returns min(wider, tighter) / max(wider, tighter), which is symmetric in
// the two directions: swapping every "wider" for a "tighter" (or vice versa)
// leaves the score unchanged. Perfect balance (equal counts) scores 1, and a
// one-sided scenario scores 0. The smaller count is the numerator so the ratio
// stays in [0,1] and is never clamped, which would otherwise mask imbalance.
func DirectionBalance(rows []Projection) float64 {
	var wider, tighter float64
	for _, row := range rows {
		switch row.Direction {
		case "wider":
			wider++
		case "tighter":
			tighter++
		}
	}
	return SafeFraction(math.Min(wider, tighter), math.Max(wider, tighter))
}

// IsNeutral reports whether no row changed beyond tolerance.
func IsNeutral(rows []Projection, tolerance float64) bool {
	if tolerance < 0 || math.IsNaN(tolerance) {
		tolerance = 0
	}
	for _, row := range rows {
		if math.Abs(row.DeltaWidth) > tolerance {
			return false
		}
	}
	return true
}
