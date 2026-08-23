// Package uncertainty turns interval and covariance data into a user-facing
// quality report without changing the persisted sample.
package uncertainty

import "task177-isomix/internal/model"

// Assess evaluates interval spread and covariance health.
func Assess(sample model.Sample) Report {
	mean, maximum, names := intervalStats(sample.Measurements)
	positive := PositiveSemidefinite(sample)
	report := Report{
		Dimension:        sample.Dimension,
		IntervalCount:    len(names),
		MeanWidth:        mean,
		MaximumRelative:  maximum,
		CovarianceTrace:  CovarianceTrace(sample),
		PositiveDefinite: positive,
		Quality:          classify(maximum, positive),
	}
	if !positive {
		report.Warnings = append(report.Warnings, "covariance matrix has a negative principal minor")
	}
	if maximum >= 0.5 {
		report.Warnings = append(report.Warnings, "at least one interval exceeds the service uncertainty threshold")
	}
	return report
}
