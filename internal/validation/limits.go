package validation

import (
	"math"

	"task177-isomix/internal/model"
)

// FiniteRange rejects NaN and infinities before they enter a solver matrix.
func FiniteRange(r model.Range) bool {
	return !math.IsNaN(r.Lo) && !math.IsNaN(r.Hi) && !math.IsInf(r.Lo, 0) && !math.IsInf(r.Hi, 0)
}

// IntervalMass returns a normalized midpoint mass for a source interval.
func IntervalMass(r model.Range) (float64, bool) {
	if !FiniteRange(r) || !r.Valid() || r.Lo < 0 {
		return 0, false
	}
	return (r.Lo + r.Hi) / 2, true
}
