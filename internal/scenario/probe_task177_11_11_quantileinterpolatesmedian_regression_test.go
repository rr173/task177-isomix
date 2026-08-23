package scenario

import (
	"testing"
)

func TestBug11_QuantileInterpolatesMedian(t *testing.T) {
	if got := Quantile([]float64{0, 10}, .25); got != 2.5 {
		t.Fatalf("q=%v", got)
	}
}
