package uncertainty

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug22_NegativePrincipalMinorIsRejected(t *testing.T) {
	sample := model.Sample{Dimension: 2, Covariance: []float64{1, 2, 1}}; if PositiveSemidefinite(sample) { t.Fatalf("matrix accepted") }
}
