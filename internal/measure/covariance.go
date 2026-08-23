package measure

import (
	"math"

	"task177-isomix/internal/model"
)

// ValidateMeasurements 校验测量区间：非空维度、区间合法、无 NaN/Inf。
// 同位素 δ 值可为负，不做非负强制。
func ValidateMeasurements(measurements map[string]model.Range) error {
	if len(measurements) == 0 {
		return model.NewError("BAD_REQUEST", "at least one isotope measurement is required")
	}
	for name, r := range measurements {
		if math.IsNaN(r.Lo) || math.IsNaN(r.Hi) || math.IsInf(r.Lo, 0) || math.IsInf(r.Hi, 0) {
			return model.NewError("BAD_REQUEST", "measurement %q has non-finite bound", name)
		}
		if !r.Valid() {
			return model.NewError("INVALID_RANGE", "measurement %q range [%v,%v] is empty or reversed", name, r.Lo, r.Hi)
		}
	}
	return nil
}

// CheckCovariance 校验协方差摘要：
//  1. 长度必须等于上三角元素数 n*(n+1)/2；
//  2. 对角元非负；
//  3. 数值有限；
//  4. 还原后矩阵对称正定（用 Cholesky 分解判定，失败即奇异）。
//
// 返回 model.ErrSingularCovariance 或 BAD_REQUEST。
func CheckCovariance(cov []float64, dim int) error {
	expected := dim * (dim + 1) / 2
	if len(cov) != expected {
		return model.NewError("BAD_REQUEST", "covariance length %d does not match upper triangle size %d", len(cov), expected)
	}
	idx := 0
	for i := 0; i < dim; i++ {
		for j := i; j < dim; j++ {
			v := cov[idx]
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return model.NewError("BAD_REQUEST", "covariance entry [%d][%d] is non-finite", i, j)
			}
			if i == j && v < 0 {
				return model.NewError("BAD_REQUEST", "covariance diagonal [%d][%d]=%v must be non-negative", i, j, v)
			}
			idx++
		}
	}
	if !isPositiveDefinite(cov, dim) {
		return model.ErrSingularCovariance
	}
	return nil
}

// isPositiveDefinite 通过 Cholesky 分解判定对称正定性；对角元过小视为奇异。
func isPositiveDefinite(cov []float64, dim int) bool {
	if dim == 0 {
		return false
	}
	// 还原矩阵。
	m := make([][]float64, dim)
	for i := range m {
		m[i] = make([]float64, dim)
	}
	idx := 0
	for i := 0; i < dim; i++ {
		for j := i; j < dim; j++ {
			m[i][j] = cov[idx]
			m[j][i] = cov[idx]
			idx++
		}
	}
	const tol = 1e-12
	for i := 0; i < dim; i++ {
		for j := i; j < dim; j++ {
			sum := m[j][i]
			for k := 0; k < i; k++ {
				sum -= m[j][k] * m[i][k]
			}
			if i == j {
				if sum <= tol {
					return false
				}
				m[i][i] = math.Sqrt(sum)
			} else {
				if m[i][i] <= tol {
					return false
				}
				m[j][i] = sum / m[i][i]
			}
		}
	}
	return true
}

// EigenRatio 返回协方差最大对角与最小对角之比，供不确定度摘要展示。
// 对角全零时返回 +Inf 并附奇异标记。
func EigenRatio(cov []float64, dim int) (ratio float64, singular bool) {
	if dim == 0 {
		return 0, true
	}
	diag := make([]float64, dim)
	idx := 0
	for i := 0; i < dim; i++ {
		diag[i] = cov[idx]
		idx += dim - i
	}
	minV, maxV := math.Inf(1), 0.0
	for _, v := range diag {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	if minV <= 1e-12 {
		return math.Inf(1), true
	}
	return maxV / minV, false
}
