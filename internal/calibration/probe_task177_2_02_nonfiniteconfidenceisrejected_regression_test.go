package calibration

import (
	"testing"
)

func TestBug02_NonFiniteConfidenceIsRejected(t *testing.T) {
	zero := 0.0; if got := Confidence(zero / zero); got != 0 { t.Fatalf("confidence=%v", got) }
}
