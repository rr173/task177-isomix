package uncertainty

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug21_CovarianceTraceUsesDiagonal(t *testing.T) {
	sample := model.Sample{Dimension: 2, Covariance: []float64{1, 4, 2}}; if got := CovarianceTrace(sample); got != 3 { t.Fatalf("trace=%v", got) }
}
