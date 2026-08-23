package scenario

import (
	"math"
	"testing"
)

func TestCoverageBoundedFraction(t *testing.T) {
	rows := []Projection{
		{ID: "a", DeltaWidth: 0.02},
		{ID: "b", DeltaWidth: -0.10},
		{ID: "c", DeltaWidth: 0.30},
	}
	// Two of three rows fall within the 0.1 movement limit.
	if got := Coverage(rows, 0.10); got != 2.0/3.0 {
		t.Fatalf("Coverage = %v, want %v", got, 2.0/3.0)
	}
	// Within-limit rows must yield full coverage, clamped to a bounded 1.
	if got := Coverage(rows, 1.0); got != 1 {
		t.Fatalf("Coverage = %v, want 1", got)
	}
	// A negative or NaN limit collapses to 0 and covers nothing.
	if got := Coverage(rows, -1); got != 0 {
		t.Fatalf("Coverage(neg limit) = %v, want 0", got)
	}
	if got := Coverage(rows, math.NaN()); got != 0 {
		t.Fatalf("Coverage(NaN limit) = %v, want 0", got)
	}
}

// Empty observations must not look fully covered: coverage stays a bounded
// fraction (0), never NaN and never the misleading 1 of "everything in limit".
func TestCoverageEmptyObservationsNotFullyCovered(t *testing.T) {
	if got := Coverage(nil, 0.10); got != 0 {
		t.Fatalf("Coverage(nil) = %v, want 0", got)
	}
	if got := Coverage([]Projection{}, 0.10); got != 0 {
		t.Fatalf("Coverage(empty) = %v, want 0", got)
	}
	if got := Coverage([]Projection{}, 0.10); math.IsNaN(got) {
		t.Fatalf("Coverage(empty) = NaN, want bounded 0")
	}
}

func TestSafeFractionBounded(t *testing.T) {
	cases := []struct {
		num, denom, want float64
	}{
		{0, 0, 0},       // empty -> not fully covered, bounded
		{0, -1, 0},      // negative denominator
		{3, 10, 0.3},
		{7, 7, 1},       // exactly full coverage (non-empty)
		{8, 7, 1},       // over-full clamps down
		{-2, 7, 0},      // negative ratio
		{math.NaN(), 5, 0},
		{5, math.NaN(), 0},
		{math.Inf(1), 5, 0},
	}
	for _, c := range cases {
		if got := SafeFraction(c.num, c.denom); got != c.want {
			t.Fatalf("SafeFraction(%v, %v) = %v, want %v", c.num, c.denom, got, c.want)
		}
	}
}

func TestDirectionBalanceBounded(t *testing.T) {
	// No rows -> bounded 0, not NaN.
	if got := DirectionBalance(nil); got != 0 {
		t.Fatalf("DirectionBalance(nil) = %v, want 0", got)
	}
	// One-sided movement -> no balance.
	rows := []Projection{
		{ID: "a", Direction: "wider"},
		{ID: "b", Direction: "wider"},
	}
	if got := DirectionBalance(rows); got != 0 {
		t.Fatalf("DirectionBalance = %v, want 0", got)
	}
	// Equal widening/tightening -> fully balanced.
	rows = []Projection{
		{ID: "a", Direction: "wider"},
		{ID: "b", Direction: "tighter"},
	}
	if got := DirectionBalance(rows); got != 1 {
		t.Fatalf("DirectionBalance = %v, want 1", got)
	}
}
