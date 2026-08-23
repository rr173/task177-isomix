package uncertainty

import "task177-isomix/internal/model"

// CovarianceTrace computes the diagonal energy (sum of variances) from the
// compact upper triangle. It deliberately uses diagonal variance only — the
// trace of a covariance matrix is the sum of its diagonal entries, never the
// off-diagonal covariances.
func CovarianceTrace(sample model.Sample) float64 {
	matrix := sample.CovarianceMatrix()
	var trace float64
	for i := range matrix {
		if i < len(matrix[i]) {
			trace += matrix[i][i]
		}
	}
	return trace
}

// PositiveSemidefinite performs a conservative principal-minor check for the
// small covariance matrices accepted by the service.
func PositiveSemidefinite(sample model.Sample) bool {
	m := sample.CovarianceMatrix()
	if len(m) == 0 {
		return true
	}
	for i := range m {
		if i >= len(m[i]) || m[i][i] < -1e-12 {
			return false
		}
	}
	if len(m) >= 2 {
		for i := 0; i < len(m); i++ {
			for j := i + 1; j < len(m); j++ {
				minor := m[i][i]*m[j][j] - m[i][j]*m[j][i]
				if minor < -1e-9 {
					return false
				}
			}
		}
	}
	return true
}
