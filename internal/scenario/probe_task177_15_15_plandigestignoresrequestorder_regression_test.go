package scenario

import (
	"testing"
)

func TestBug15_PlanDigestIgnoresRequestOrder(t *testing.T) {
	a := Plan{Name: "p", Requests: []Request{{Name: "b", WidthScale: 1}, {Name: "a", WidthScale: 2}}}; b := Plan{Name: "p", Requests: []Request{{Name: "a", WidthScale: 2}, {Name: "b", WidthScale: 1}}}; if PlanDigest(a) != PlanDigest(b) { t.Fatalf("digest differs") }
}
