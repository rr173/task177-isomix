package scenario

import (
	"testing"
)

func TestBug10_NetWidthChangeKeepsDirection(t *testing.T) {
	rows := []Projection{{DeltaWidth: .2}, {DeltaWidth: -.1}}; if got := NetWidthChange(rows); got != .1 { t.Fatalf("change=%v", got) }
}
