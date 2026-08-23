package scenario

import (
	"testing"
)

func TestBug14_DefaultPlanPreservesRequestOrder(t *testing.T) {
	plan := DefaultPlan(); if len(plan.Order) != 4 || plan.Order[0] != "tight-10" { t.Fatalf("%#v", plan) }
}
