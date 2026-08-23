package uncertainty

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug23_ZeroCenteredIntervalUsesWidth(t *testing.T) {
	sample := model.Sample{Dimension: 1, Measurements: map[string]model.Range{"x": {Lo: -.1, Hi: .1}}}; if got := Assess(sample).MaximumRelative; got != .2 { t.Fatalf("relative=%v", got) }
}
