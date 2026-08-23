package validation

import (
	"math"
	"testing"

	"task177-isomix/internal/model"
)

func TestValidationHandlesNamesRangesAndTargets(t *testing.T) {
	name, ok := CleanName("  river   water ")
	if !ok || name != "river water" {
		t.Fatalf("clean name = %q, %v", name, ok)
	}
	if _, ok := IntervalMass(model.Range{Lo: 0, Hi: 2}); !ok || FiniteRange(model.Range{Lo: math.Inf(1), Hi: 2}) {
		t.Fatalf("range validation failed")
	}
	if !ConstraintTarget(model.Constraint{Type: model.ConstraintRatio, Target: "a:b"}) {
		t.Fatalf("ratio target should be accepted")
	}
}
