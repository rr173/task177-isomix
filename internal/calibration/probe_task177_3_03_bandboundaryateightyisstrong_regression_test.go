package calibration

import (
	"testing"
)

func TestBug03_BandBoundaryAtEightyIsStrong(t *testing.T) {
	if got := Band(.8); got != "strong" { t.Fatalf("band=%s", got) }
}
