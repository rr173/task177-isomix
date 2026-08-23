package uncertainty

import "task177-isomix/internal/model"

// CovarianceTrace computes the diagonal energy from the compact upper triangle.
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
				// 二阶主子式 det([[m_ii,m_ij],[m_ji,m_jj]]) 对正半定矩阵必须非负；
				// 为负即存在负主子式，矩阵非正半定，须保持为被拒绝的测量边界。
				// 与上面的对角元判定一致，使用 -1e-12 数值容差吸收浮点噪声。
				minor := m[i][i]*m[j][j] - m[i][j]*m[j][i]
				if minor < -1e-12 {
					return false
				}
			}
		}
	}
	return true
}
