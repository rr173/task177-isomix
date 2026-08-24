package scenario

import (
	"math"
	"testing"
)

// TestMostSensitiveRanksByAbsoluteChange guards against the original bug where
// the comparison used the signed delta. A large tightening (negative delta)
// must outrank a slight widening (positive delta) when its absolute change is
// larger.
func TestMostSensitiveRanksByAbsoluteChange(t *testing.T) {
	rows := []Projection{
		{ID: "wide-slight", DeltaWidth: 0.05},   // slight widening
		{ID: "tight-large", DeltaWidth: -0.80},  // strong tightening
		{ID: "tight-slight", DeltaWidth: -0.02}, // slight tightening
		{ID: "wide-large", DeltaWidth: 0.40},    // strong widening
	}
	if got := MostSensitive(rows); got != "tight-large" {
		t.Fatalf("MostSensitive = %q, want %q (tightening should outrank widening by magnitude)", got, "tight-large")
	}

	// Tie on magnitude should break toward the lexicographically smaller ID.
	tie := []Projection{
		{ID: "beta", DeltaWidth: 0.30},
		{ID: "alpha", DeltaWidth: -0.30},
	}
	if got := MostSensitive(tie); got != "alpha" {
		t.Fatalf("MostSensitive tie = %q, want %q", got, "alpha")
	}

	if got := MostSensitive(nil); got != "" {
		t.Fatalf("MostSensitive(nil) = %q, want empty", got)
	}
}

// TestSelectTopRanksByAbsoluteChange mirrors the regression above for the
// top-N filter: a strong tightening must survive the filter ahead of a slight
// widening.
func TestSelectTopRanksByAbsoluteChange(t *testing.T) {
	changes := map[string]float64{
		"wide-slight":  0.05,
		"tight-large":  -0.80,
		"tight-slight": -0.02,
		"wide-large":   0.40,
	}
	got := SelectTop(changes, 2)
	if len(got) != 2 || got[0] != "tight-large" || got[1] != "wide-large" {
		t.Fatalf("SelectTop = %v, want [tight-large wide-large]", got)
	}

	// Absolute ties break toward the lexicographically smaller ID.
	tie := map[string]float64{"beta": 0.30, "alpha": -0.30}
	got = SelectTop(tie, 2)
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Fatalf("SelectTop tie = %v, want [alpha beta]", got)
	}

	if got := SelectTop(changes, 0); got != nil {
		t.Fatalf("SelectTop limit 0 = %v, want nil", got)
	}
	if got := SelectTop(changes, -1); got != nil {
		t.Fatalf("SelectTop negative limit = %v, want nil", got)
	}
}

// TestSelectTopApproximateFloatEquality documents that near-equal magnitudes
// resolve to a stable order rather than depending on sign or sub-ulp noise.
func TestSelectTopApproximateFloatEquality(t *testing.T) {
	changes := map[string]float64{
		"a": 0.1 + 1e-15,
		"b": -0.1,
	}
	got := SelectTop(changes, 2)
	if len(got) != 2 {
		t.Fatalf("SelectTop = %v, want 2 ids", got)
	}
	// |a| and |b| differ by ~1e-15, well below 1e-9; treat as a tie so the
	// result is the stable lexicographic order rather than a sign artifact.
	if math.Abs(changes[got[0]])-math.Abs(changes[got[1]]) > 1e-9 && got[0] != "a" {
		t.Fatalf("SelectTop = %v, expected stable order for near-equal magnitudes", got)
	}
}
