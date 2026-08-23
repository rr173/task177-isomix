package uncertainty

import (
	"math"
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

// TestAssessZeroCenteredIntervalUsesAbsoluteWidth 验证以零为中心的区间
// 不会因除零而把相对宽度吹成 +Inf/NaN：它必须退化为绝对宽度。
func TestAssessZeroCenteredIntervalUsesAbsoluteWidth(t *testing.T) {
	sample := model.Sample{
		Dimension:  1,
		Measurements: map[string]model.Range{"ratio": {Lo: -0.05, Hi: 0.05}},
		Covariance:  []float64{1},
	}
	report := Assess(sample)
	if math.IsInf(report.MaximumRelative, 0) || math.IsNaN(report.MaximumRelative) {
		t.Fatalf("zero-centered interval produced non-finite relative width: %v", report.MaximumRelative)
	}
	// 绝对宽度 = 0.1，且 <= 0.5，应落在 high-confidence 桶，而非因 +Inf 沦为 overspread。
	if report.MaximumRelative != 0.1 {
		t.Fatalf("MaximumRelative = %v, want 0.1 (absolute width fallback)", report.MaximumRelative)
	}
	if report.Quality != "high-confidence" {
		t.Fatalf("Quality = %q, want high-confidence", report.Quality)
	}
}
