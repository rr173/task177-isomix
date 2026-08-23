package scenario

import (
	"testing"
)

func TestBug17_CoverageWithEmptyRowsIsZero(t *testing.T) {
	if got := Coverage(nil, .1); got != 0 { t.Fatalf("coverage=%v", got) }
}
