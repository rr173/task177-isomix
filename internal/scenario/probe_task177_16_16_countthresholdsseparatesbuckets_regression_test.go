package scenario

import (
	"testing"
)

func TestBug16_CountThresholdsSeparatesBuckets(t *testing.T) {
	rows := []Projection{{DeltaWidth: .01}, {DeltaWidth: .1}, {DeltaWidth: .3}}; got := CountThresholds(rows); if got.Moderate != 1 || got.Large != 1 || got.Extreme != 1 { t.Fatalf("%#v", got) }
}
