package uncertainty

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug25_WideIntervalsProduceWarning(t *testing.T) {
	sample := model.Sample{Dimension: 1, Measurements: map[string]model.Range{"x": {Lo: 0, Hi: 2}}}; if len(Assess(sample).Warnings) == 0 { t.Fatalf("warning missing") }
}
