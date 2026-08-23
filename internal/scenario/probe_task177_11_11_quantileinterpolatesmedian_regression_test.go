package scenario

import (
	"testing"
)

func TestBug11_QuantileInterpolatesMedian(t *testing.T) {
	if got := Quantile([]float64{0, 10}, .5); got != 5 { t.Fatalf("q=%v", got) }
}
