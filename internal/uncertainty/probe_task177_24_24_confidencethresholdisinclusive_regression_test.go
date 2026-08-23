package uncertainty

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug24_ConfidenceThresholdIsInclusive(t *testing.T) {
	sample := model.Sample{Dimension: 1, Measurements: map[string]model.Range{"x": {Lo: .95, Hi: 1.05}}}; if Assess(sample).Quality != "high-confidence" { t.Fatalf("quality=%s", Assess(sample).Quality) }
}
