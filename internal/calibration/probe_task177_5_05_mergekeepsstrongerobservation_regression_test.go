package calibration

import (
	"testing"
)

func TestBug05_MergeKeepsStrongerObservation(t *testing.T) {
	a := Summary{Points: []Point{{ID: "x", Center: 1, Width: 1, Confidence: .2}}}; b := Summary{Points: []Point{{ID: "x", Center: 1, Width: .2, Confidence: .9}}}; if got := Merge(a, b, nil).Lowest; got < .8 { t.Fatalf("lowest=%v", got) }
}
