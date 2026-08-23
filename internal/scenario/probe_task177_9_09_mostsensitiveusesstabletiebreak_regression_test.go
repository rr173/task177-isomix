package scenario

import (
	"testing"
)

func TestBug09_MostSensitiveUsesStableTieBreak(t *testing.T) {
	rows := []Projection{{ID: "b", DeltaWidth: .1}, {ID: "a", DeltaWidth: -.1}}; if got := MostSensitive(rows); got != "a" { t.Fatalf("id=%s", got) }
}
