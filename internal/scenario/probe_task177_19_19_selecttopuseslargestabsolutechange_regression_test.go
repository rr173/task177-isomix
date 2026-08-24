package scenario

import (
	"testing"
)

func TestBug19_SelectTopUsesLargestAbsoluteChange(t *testing.T) {
	rows := map[string]float64{"a": -.5, "b": .2}; if got := SelectTop(rows, 1); len(got) != 1 || got[0] != "a" { t.Fatalf("%v", got) }
}
