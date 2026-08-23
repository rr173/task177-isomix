package calibration

import (
	"math"
	"testing"
	"time"

	"task177-isomix/internal/model"
)

// TestConfidenceRejectsNonFinite pins the contract that non-finite interval
// confidence must remain rejected: a non-finite relative width (NaN/Inf) is
// produced whenever a solved bound is NaN or ±Inf, and such a bound carries no
// information. Ranking and aggregation depend on this staying zero so an
// unstable endmember never tops the ranking nor inflates the aggregate summary.
func TestConfidenceRejectsNonFinite(t *testing.T) {
	for _, rw := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if c := Confidence(rw); c != 0 {
			t.Fatalf("Confidence(%v) = %v, want 0 (rejected)", rw, c)
		}
	}
	// Finite widths keep their normal curve; zero/negative widths stay maximal.
	if c := Confidence(0); c != 1 {
		t.Fatalf("Confidence(0) = %v, want 1", c)
	}
	if c := Confidence(1); c != 1/(1+1.0*1.0*4) {
		t.Fatalf("Confidence(1) = %v, want %v", c, 1/(1+1.0*1.0*4))
	}
}

// TestAssessRejectsNonFiniteBounds exercises the full ranking+aggregation path:
// a non-finite bound must land at the bottom of the ranking and pull the
// interpretation away from "tightly constrained" even when the other bounds are.
func TestAssessRejectsNonFiniteBounds(t *testing.T) {
	bounds := map[string]model.Range{
		"tight":   {Lo: 5.2, Hi: 5.8},                 // small finite width -> high confidence
		"unstable": {Lo: math.NaN(), Hi: math.Inf(1)}, // non-finite -> rejected (0)
	}
	summary := Assess(bounds, func() time.Time { return time.Time{} })

	// Ranking is ascending by confidence: the rejected bound must be first.
	if len(summary.Points) != 2 || summary.Points[0].ID != "unstable" {
		t.Fatalf("ranking did not place the non-finite bound first: %+v", summary.Points)
	}
	if summary.Points[0].Confidence != 0 {
		t.Fatalf("non-finite confidence = %v, want 0", summary.Points[0].Confidence)
	}
	if Band(summary.Points[0].Confidence) != "insufficient" {
		t.Fatalf("non-finite band = %q, want insufficient", Band(summary.Points[0].Confidence))
	}

	// Aggregation must reflect the rejection: lowest is 0, so it is not "tight".
	if summary.Lowest != 0 {
		t.Fatalf("lowest confidence = %v, want 0", summary.Lowest)
	}
	if summary.Interpretation == "all sources are tightly constrained" {
		t.Fatalf("interpretation must not claim tight constraints with a non-finite bound, got %q", summary.Interpretation)
	}
}
