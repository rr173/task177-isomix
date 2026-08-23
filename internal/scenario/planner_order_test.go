package scenario

import (
	"strings"
	"testing"
)

func TestDefaultPlanOrderIsTightToWide(t *testing.T) {
	plan := DefaultPlan()
	got := strings.Join(plan.Order, ",")
	want := "tight-10,tight-25,wide-10,wide-25"
	if got != want {
		t.Fatalf("plan.Order = %q, want %q (tight-to-wide)", got, want)
	}
	if len(plan.Order) != len(plan.Requests) {
		t.Fatalf("Order len %d != Requests len %d", len(plan.Order), len(plan.Requests))
	}
	for i, r := range plan.Requests {
		if plan.Order[i] != r.Name {
			t.Fatalf("Order[%d]=%q != Requests[%d].Name=%q", i, plan.Order[i], i, r.Name)
		}
	}
}
