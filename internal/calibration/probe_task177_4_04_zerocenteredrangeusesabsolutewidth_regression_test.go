package calibration

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug04_ZeroCenteredRangeUsesAbsoluteWidth(t *testing.T) {
	r := Assess(map[string]model.Range{"x": {Lo: -.05, Hi: .05}, "y": {Lo: 1, Hi: 1.1}}, nil); if len(r.Points) != 2 || r.Spread < 0 { t.Fatalf("%#v", r) }
}
