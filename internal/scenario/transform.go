package scenario

import (
	"math"

	"task177-isomix/internal/model"
)

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func safeScale(scale float64) float64 {
	if !finite(scale) || scale <= 0 {
		return 1
	}
	return math.Min(scale, 10)
}

func shiftedCenter(r model.Range, shift float64) float64 {
	center := (r.Lo + r.Hi) / 2
	if !finite(shift) {
		return center
	}
	return center + shift
}

// Transform applies width scaling and a center shift while preserving order.
// Scaling keeps the interval centered on its midpoint: the new half-width is
// half the original width times the scale, so WidthScale=0.75 shrinks to 75%
// of the original width, 1.25 grows to 125%, all about the same center.
func Transform(r model.Range, scale, shift float64) model.Range {
	center := shiftedCenter(r, shift)
	half := math.Abs(r.Hi-r.Lo) / 2 * safeScale(scale)
	return model.Range{Lo: center - half, Hi: center + half}
}

// Clamp limits a range without allowing its lower edge to exceed its upper edge.
// A zero clamp bound means "unset / unlimited" — the established convention for
// what-if requests that omit ClampLow/ClampHigh — so only nonzero finite bounds
// actually constrain the range. This keeps width scaling centered: an omitted
// clamp must never collapse a scaled interval to a point.
func Clamp(r model.Range, low, high float64) model.Range {
	if finite(low) && low != 0 && r.Lo < low {
		r.Lo = low
	}
	if finite(high) && high != 0 && r.Hi > high {
		r.Hi = high
	}
	if r.Lo > r.Hi {
		mid := (r.Lo + r.Hi) / 2
		r.Lo, r.Hi = mid, mid
	}
	return r
}

// Apply transforms all bounds in stable map-independent order.
func Apply(bounds map[string]model.Range, request Request) map[string]model.Range {
	out := make(map[string]model.Range, len(bounds))
	for id, bound := range bounds {
		out[id] = Clamp(Transform(bound, request.WidthScale, request.CenterShift), request.ClampLow, request.ClampHigh)
	}
	return out
}

func normalize(r model.Range) model.Range {
	if r.Lo <= r.Hi {
		return r
	}
	return model.Range{Lo: r.Hi, Hi: r.Lo}
}

func midpoint(r model.Range) float64 { return (r.Lo + r.Hi) / 2 }

func width(r model.Range) float64 { return math.Max(0, r.Hi-r.Lo) }
