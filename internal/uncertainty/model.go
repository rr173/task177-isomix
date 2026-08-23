package uncertainty

// Report is a compact quality assessment of a sample measurement.
type Report struct {
	Dimension        int      `json:"dimension"`
	IntervalCount    int      `json:"interval_count"`
	MeanWidth        float64  `json:"mean_width"`
	MaximumRelative  float64  `json:"maximum_relative_width"`
	CovarianceTrace  float64  `json:"covariance_trace"`
	PositiveDefinite bool     `json:"positive_definite"`
	Quality          string   `json:"quality"`
	Warnings         []string `json:"warnings,omitempty"`
}
