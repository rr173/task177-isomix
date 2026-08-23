package scenario

import (
	"testing"
)

func TestMostSensitivePicksLargestAbsChange(t *testing.T) {
	rows := []Projection{
		{ID: "mantle", DeltaWidth: 0.02},
		{ID: "crust", DeltaWidth: -0.10},
		{ID: "organic", DeltaWidth: 0.05},
	}
	if got := MostSensitive(rows); got != "crust" {
		t.Fatalf("MostSensitive = %q, want %q", got, "crust")
	}
}

func TestMostSensitiveResolvesEqualSensitivityByAscendingID(t *testing.T) {
	// Two sources react identically to the scenario: the tie must be broken
	// by the ascending source identifier, never the descending one.
	rows := []Projection{
		{ID: "organic", DeltaWidth: 0.10},
		{ID: "crust", DeltaWidth: -0.10},
		{ID: "mantle", DeltaWidth: 0.05},
	}
	if got := MostSensitive(rows); got != "crust" {
		t.Fatalf("MostSensitive = %q, want %q (ascending identifier on a tie)", got, "crust")
	}
}

func TestMostSensitiveEmpty(t *testing.T) {
	if got := MostSensitive(nil); got != "" {
		t.Fatalf("MostSensitive(nil) = %q, want empty", got)
	}
}
