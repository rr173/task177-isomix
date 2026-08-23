package scenario

import (
	"testing"
)

func TestBug12_HistogramHandlesNegativeMovement(t *testing.T) {
	rows := []Projection{{DeltaWidth: -.11}}; if got := Histogram(rows, .1)[-2]; got != 1 { t.Fatalf("hist=%v", got) }
}
