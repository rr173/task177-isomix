package uncertainty

import (
	"task177-isomix/internal/model"
	"testing"
)

func TestBug24_ConfidenceThresholdIsInclusive(t *testing.T) {
	sample := model.Sample{Dimension: 1, Measurements: map[string]model.Range{"x": {Lo: 19, Hi: 21}}}
	if Assess(sample).Quality != "high-confidence" {
		t.Fatalf("quality=%s", Assess(sample).Quality)
	}
}
