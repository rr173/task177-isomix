package calibration

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug01_TightIntervalKeepsFullConfidence(t *testing.T) {
	r := Assess(map[string]model.Range{"x": {Lo: 1, Hi: 1}}, nil); if r.Lowest != 1 || r.Interpretation != "all sources are tightly constrained" { t.Fatalf("%#v", r) }
}
