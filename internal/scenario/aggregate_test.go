package scenario

import (
	"math"
	"testing"
)

func TestNormalizeRowsKeepsLargestAbsoluteMovement(t *testing.T) {
	// Two rows share ID "a": a large tightening move (negative DeltaWidth)
	// and a small widening move (positive DeltaWidth). The signed-value
	// implementation retained the small positive move; the spec requires
	// the largest *absolute* movement to win regardless of input order.
	largeTight := Projection{ID: "a", DeltaWidth: -0.10, Direction: "tighter"}
	smallWide := Projection{ID: "a", DeltaWidth: +0.02, Direction: "wider"}
	distinct := Projection{ID: "b", DeltaWidth: +0.05, Direction: "wider"}

	for i, order := range [][]Projection{
		{largeTight, smallWide, distinct},
		{smallWide, largeTight, distinct},
		{distinct, smallWide, largeTight},
	} {
		got := NormalizeRows(order)
		if len(got) != 2 {
			t.Fatalf("case %d: expected 2 rows, got %d", i, len(got))
		}
		var aRow Projection
		for _, row := range got {
			if row.ID == "a" {
				aRow = row
			}
		}
		if aRow.DeltaWidth != largeTight.DeltaWidth {
			t.Fatalf("case %d: expected largest absolute move (DeltaWidth=%v) to win, got %v",
				i, largeTight.DeltaWidth, aRow.DeltaWidth)
		}
	}
}

func TestNormalizeRowsBreaksTiesStably(t *testing.T) {
	// When two rows tie on absolute movement, a deterministic tie-break
	// keeps behavior reproducible. We assert the set is stable rather than
	// a particular winner, since magnitude is equal.
	left := Projection{ID: "a", DeltaWidth: +0.10, Direction: "wider"}
	right := Projection{ID: "a", DeltaWidth: -0.10, Direction: "tighter"}
	got1 := NormalizeRows([]Projection{left, right})
	got2 := NormalizeRows([]Projection{right, left})
	if len(got1) != 1 || len(got2) != 1 {
		t.Fatalf("expected deduplicated single row")
	}
	if math.Abs(got1[0].DeltaWidth) != math.Abs(got2[0].DeltaWidth) {
		t.Fatalf("tie-break should preserve magnitude across input orders")
	}
}
