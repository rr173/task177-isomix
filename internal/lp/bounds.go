package lp

import "fmt"

// Bounds 对每个变量分别求最小与最大值（各一次 LP），得到可行域在各轴上的投影区间。
// 任一 LP 非 optimal 时返回该状态与证据。
func Bounds(p *Problem) (lows, highs []float64, unstable bool, evidence string) {
	n := len(p.C)
	lows = make([]float64, n)
	highs = make([]float64, n)
	for j := 0; j < n; j++ {
		// min x_j
		c := make([]float64, n)
		c[j] = 1
		r := Solve(&Problem{A: p.A, B: p.B, C: c})
		if r.Status != StatusOptimal {
			return nil, nil, true, fmt.Sprintf("min bound variable %d: %s (%s)", j, r.Status, r.Evidence)
		}
		lows[j] = r.X[j]

		// max x_j == min -x_j
		c = make([]float64, n)
		c[j] = -1
		r = Solve(&Problem{A: p.A, B: p.B, C: c})
		if r.Status != StatusOptimal {
			return nil, nil, true, fmt.Sprintf("max bound variable %d: %s (%s)", j, r.Status, r.Evidence)
		}
		highs[j] = -r.Objective // min(-x_j) = -max(x_j)
	}
	return lows, highs, false, ""
}
