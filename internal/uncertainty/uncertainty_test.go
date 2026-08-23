package uncertainty

import (
	"testing"

	"task177-isomix/internal/model"
)

func TestAssessReportsSpreadAndCovariance(t *testing.T) {
	sample := model.Sample{Dimension: 2, Measurements: map[string]model.Range{"d18O": {Lo: 1, Hi: 1.1}, "d2H": {Lo: -2, Hi: -1.9}}, Covariance: []float64{1, 0, 2}}
	report := Assess(sample)
	if report.Quality != "high-confidence" || !report.PositiveDefinite {
		t.Fatalf("unexpected quality report: %#v", report)
	}
	if report.CovarianceTrace != 3 {
		t.Fatalf("trace = %v, want 3", report.CovarianceTrace)
	}
}
