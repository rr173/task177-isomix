package lp

// Feasible 判断约束系统 A x <= b, x >= 0 是否存在可行解。
// 通过求解目标为 0 的 LP 完成：可行 iff 状态为 optimal。
func Feasible(p *Problem) (bool, Result) {
	c := make([]float64, len(p.C))
	for i := range c {
		c[i] = 0
	}
	pp := &Problem{A: p.A, B: p.B, C: c}
	res := Solve(pp)
	if res.Status == StatusOptimal {
		return true, res
	}
	return false, res
}

// Solution 可行解（非严格），用于活跃约束判定与快照展示。
func Solution(p *Problem) Result {
	c := make([]float64, len(p.C))
	for i := range c {
		c[i] = 0
	}
	return Solve(&Problem{A: p.A, B: p.B, C: c})
}

// ActiveRows 给定可行解，返回每条约束的松弛量及是否活跃（|a·x - b| <= tol）。
// rows 与 A/B 等长一一对应；解向量来自 Solve 的 X。
func ActiveRows(p *Problem, x []float64) (active []bool, slacks []float64) {
	m := len(p.B)
	active = make([]bool, m)
	slacks = make([]float64, m)
	for i := 0; i < m; i++ {
		dot := 0.0
		for j := 0; j < len(x); j++ {
			dot += p.A[i][j] * x[j]
		}
		slack := p.B[i] - dot
		slacks[i] = slack
		active[i] = slack >= -tol && slack <= tol
	}
	return active, slacks
}
