package scenario

import (
	"testing"
)

func TestBug18_DirectionBalanceIsSymmetric(t *testing.T) {
	rows := []Projection{{Direction: "wider"}, {Direction: "wider"}, {Direction: "tighter"}}; if got := DirectionBalance(rows); got != .5 { t.Fatalf("balance=%v", got) }
}
