package scenario

import (
	"testing"
)

func TestBug13_EmptyMeanIsZero(t *testing.T) {
	if got := MeanChange(nil); got != 0 { t.Fatalf("mean=%v", got) }
}
