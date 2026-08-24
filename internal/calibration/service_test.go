package calibration

import (
	"testing"
	"time"

	"task177-isomix/internal/model"
)

func fixedClock() func() time.Time { return func() time.Time { return time.Unix(0, 0) } }

// confidenceForWidth builds a Point for one source at a given interval width.
// Narrower widths map to higher (stronger) confidence.
func confidenceForWidth(id string, width float64) Point {
	r := model.Range{Lo: 1 - width/2, Hi: 1 + width/2}
	rel := relativeWidth(r)
	conf := Confidence(rel)
	return Point{ID: id, Center: (r.Lo + r.Hi) / 2, Width: r.Width(), Confidence: conf, Band: Band(conf)}
}

func TestMergeRetainsStrongerConfidence(t *testing.T) {
	strong := confidenceForWidth("mantle", 0.05) // narrow → high confidence
	weak := confidenceForWidth("mantle", 1.00)   // wide   → low confidence
	if weak.Confidence >= strong.Confidence {
		t.Fatalf("test fixture wrong: weak=%v should be below strong=%v", weak.Confidence, strong.Confidence)
	}

	first := Summary{Points: []Point{strong}, Average: strong.Confidence, Lowest: strong.Confidence, Highest: strong.Confidence}
	second := Summary{Points: []Point{weak}, Average: weak.Confidence, Lowest: weak.Confidence, Highest: weak.Confidence}

	// second observation is weaker → must NOT replace the stronger first one.
	merged := Merge(first, second, fixedClock())
	got := merged.Points[0]
	if got.Confidence != strong.Confidence {
		t.Fatalf("Merge dropped the stronger confidence: got %v, want %v", got.Confidence, strong.Confidence)
	}

	// Reverse order: stronger second observation must replace the weaker first.
	merged2 := Merge(second, first, fixedClock())
	got2 := merged2.Points[0]
	if got2.Confidence != strong.Confidence {
		t.Fatalf("Merge failed to upgrade to stronger confidence: got %v, want %v", got2.Confidence, strong.Confidence)
	}
}

func TestMergeKeepsDeterministicOrderOnTie(t *testing.T) {
	a := confidenceForWidth("mantle", 0.10)
	b := confidenceForWidth("mantle", 0.10) // equal confidence, different width-bearing range is fine
	if a.Confidence != b.Confidence {
		t.Fatalf("expected equal confidence: %v vs %v", a.Confidence, b.Confidence)
	}
	// On a tie the first observation must be retained (deterministic ordering
	// unchanged). The merged point should reflect the first interval.
	first := Summary{Points: []Point{a}}
	second := Summary{Points: []Point{b}}
	merged := Merge(first, second, fixedClock())
	got := merged.Points[0]
	if got.Width != a.Width {
		t.Fatalf("Merge changed the tied observation: got width %v, want first %v", got.Width, a.Width)
	}
}
