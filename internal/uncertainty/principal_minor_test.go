package uncertainty

import (
	"testing"

	"task177-isomix/internal/model"
)

// TestAssessRejectsNegativePrincipalMinor ensures that a covariance matrix with
// a negative 2x2 principal minor is kept on the rejected measurement boundary:
// PositiveDefinite must be false, quality "invalid-covariance", and a warning
// about the negative principal minor must be emitted. The 2x2 principal minor of
// [[1, 1.2], [1.2, 1]] is 1*1 - 1.2*1.2 = -0.44 < 0, so the matrix is not
// positive semidefinite. Previously the threshold was `-9`, which let -0.44 slip
// through and be reported as high-confidence.
func TestAssessRejectsNegativePrincipalMinor(t *testing.T) {
	sample := model.Sample{
		Dimension:    2,
		Measurements: map[string]model.Range{"d18O": {Lo: 1, Hi: 1.1}, "d2H": {Lo: -2, Hi: -1.9}},
		Covariance:   []float64{1, 1.2, 1},
	}
	report := Assess(sample)
	if report.PositiveDefinite {
		t.Fatalf("negative principal minor must remain rejected, got PositiveDefinite=true: %#v", report)
	}
	if report.Quality != "invalid-covariance" {
		t.Fatalf("quality = %q, want invalid-covariance", report.Quality)
	}
	if !contains(report.Warnings, "covariance matrix has a negative principal minor") {
		t.Fatalf("missing negative-principal-minor warning, got %v", report.Warnings)
	}
}

// TestAssessAcceptsPositiveSemidefinite confirms a clearly PSD covariance is
// still accepted after tightening the principal-minor threshold.
func TestAssessAcceptsPositiveSemidefinite(t *testing.T) {
	sample := model.Sample{
		Dimension:    2,
		Measurements: map[string]model.Range{"d18O": {Lo: 1, Hi: 1.1}, "d2H": {Lo: -2, Hi: -1.9}},
		Covariance:   []float64{1, 0.3, 1}, // minor = 1 - 0.09 = 0.91 > 0
	}
	report := Assess(sample)
	if !report.PositiveDefinite {
		t.Fatalf("PSD covariance wrongly rejected: %#v", report)
	}
	if report.Quality == "invalid-covariance" {
		t.Fatalf("PSD covariance classified invalid: %#v", report)
	}
}

func contains(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}
