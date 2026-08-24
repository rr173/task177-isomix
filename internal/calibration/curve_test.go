package calibration

import (
	"math"
	"testing"
	"time"

	"task177-isomix/internal/model"
)

// TestConfidenceTightIntervalsPreserveCalibration locks in the fix for the bug
// where tight (zero-width) intervals lost their calibrated confidence across
// the curve, the ranking, and the summary classification.
func TestConfidenceTightIntervalsPreserveCalibration(t *testing.T) {
	// A zero-width interval is fully determined: it must carry the maximum
	// confidence, not be treated as a hard failure.
	if c := Confidence(0); c != 1 {
		t.Fatalf("Confidence(0) = %v, want 1 (tight interval is fully determined)", c)
	}
	if b := Band(Confidence(0)); b != "strong" {
		t.Fatalf("Band at width 0 = %q, want strong", b)
	}

	// The curve must be monotonic non-increasing: confidence drops as the
	// relative width grows. The earlier bug produced a non-monotonic dip at 0.
	curve := Curve()
	for i := 1; i < len(curve); i++ {
		if curve[i].Confidence > curve[i-1].Confidence+1e-12 {
			t.Fatalf("curve is non-monotonic: width %.3g (%.4f) < width %.3g (%.4f)",
				curve[i-1].RelativeWidth, curve[i-1].Confidence,
				curve[i].RelativeWidth, curve[i].Confidence)
		}
	}
	if curve[0].Confidence != 1 || curve[0].Band != "strong" {
		t.Fatalf("curve landmark at width 0 = {%.4f, %q}, want {1, strong}",
			curve[0].Confidence, curve[0].Band)
	}

	// A collapsed endmember interval [0,0] (e.g. an excluded source) now
	// carries the top confidence rather than the bottom. The ranking lists
	// sources from least to most confident, so the fully-determined source
	// must be last and read as strong.
	summary := Assess(map[string]model.Range{
		"free":   {Lo: 0.3, Hi: 0.5},
		"locked": {Lo: 0, Hi: 0},
	}, func() time.Time { return time.Time{} })
	top := summary.Points[len(summary.Points)-1]
	if top.ID != "locked" {
		t.Fatalf("ranking should end with the fully-determined source, got %s", top.ID)
	}
	if top.Confidence != 1 || top.Band != "strong" {
		t.Fatalf("collapsed interval calibrated to {%.4f, %q}, want {1, strong}",
			top.Confidence, top.Band)
	}
	// The collapsed source no longer drags the summary down to "insufficient";
	// every source is strong, so the summary reads as tightly constrained.
	if summary.Interpretation != "all sources are tightly constrained" {
		t.Fatalf("interpretation = %q, want all sources are tightly constrained", summary.Interpretation)
	}

	// Non-finite input remains a hard failure (guard retained).
	if c := Confidence(math.NaN()); c != 0 {
		t.Fatalf("Confidence(NaN) = %v, want 0", c)
	}
	if c := Confidence(math.Inf(1)); c != 0 {
		t.Fatalf("Confidence(+Inf) = %v, want 0", c)
	}
}
