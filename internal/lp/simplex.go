// Package lp 实现精简的两阶段单纯形法，求解线性规划：
//
//	min  c^T x
//	s.t. A x <= b
//	     x >= 0
//
// 等式约束由调用方拆分为两条不等式（如 Σx = 1 -> Σx <= 1 与 -Σx <= -1）。
// 数值策略：容差 1e-9；Bland 最小索引规则保证退化时仍终止；NaN/Inf 一律判为
// 数值不稳定并返回证据，绝不伪造精确解。
package lp

import (
	"fmt"
	"math"
)

// tol 数值容差。
const tol = 1e-9

// Status 求解状态。
type Status string

const (
	StatusOptimal    Status = "optimal"     // 找到最优解
	StatusInfeasible Status = "infeasible"  // 约束不可行
	StatusUnbounded  Status = "unbounded"   // 目标无界
	StatusUnstable   Status = "unstable"    // 数值不稳定，无法给可信解
)

// Result 求解结果。
type Result struct {
	Status    Status
	X         []float64 // 原始变量解（仅 optimal 时有效）
	Objective float64   // 目标值（仅 optimal 时有效）
	Evidence  string    // 非 optimal 状态下的说明
}

// Problem 线性规划输入。
type Problem struct {
	A [][]float64 // m x n 不等式系数
	B []float64   // m 右端
	C []float64   // n 目标系数（求最小）
}

// Solve 求解 LP。返回值的 X 为原始变量解（长度 n）。
func Solve(p *Problem) Result {
	m, n := len(p.B), len(p.C)
	if m == 0 || n == 0 {
		return Result{Status: StatusUnstable, Evidence: "empty problem"}
	}
	for _, row := range p.A {
		if len(row) != n {
			return Result{Status: StatusUnstable, Evidence: "row length mismatch"}
		}
	}

	// 标准型：A x + slack = b。b<0 的行翻转符号并引入人工变量。
	// 先统计需要人工变量的行数，仅分配实际需要的列，避免出现未使用空列。
	artCount := 0
	for i := 0; i < m; i++ {
		if p.B[i] < 0 {
			artCount++
		}
	}
	// 列布局：[0..n) 原始变量 | [n..n+m) 松弛 | [n+m..n+m+artCount) 人工变量。
	total := n + m + artCount
	tab := make([][]float64, m+1) // 最后一行是成本行
	for i := range tab {
		tab[i] = make([]float64, total+1) // +1 列 RHS
	}

	artCols := make([]bool, total)
	basic := make([]int, m)
	artIdx := 0
	for i := 0; i < m; i++ {
		row := tab[i]
		if p.B[i] >= 0 {
			for j := 0; j < n; j++ {
				row[j] = p.A[i][j]
			}
			row[n+i] = 1
			row[total] = p.B[i]
			basic[i] = n + i
		} else {
			for j := 0; j < n; j++ {
				row[j] = -p.A[i][j]
			}
			row[n+i] = -1
			art := n + m + artIdx
			row[art] = 1
			row[total] = -p.B[i]
			basic[i] = art
			artCols[art] = true
			artIdx++
		}
	}

	// 阶段一：目标 min Σ 人工变量；阶段二：目标 min c^T x。
	phase1 := artIdx > 0

	// 成本向量（长度 total，不含 RHS 列）。
	costVec := make([]float64, total)
	if phase1 {
		for k := 0; k < artIdx; k++ {
			costVec[n+m+k] = 1
		}
	} else {
		copy(costVec[:n], p.C)
	}
	recomputeCosts(tab, basic, m, total, costVec)

	iter := 0
	const maxIter = 200000
	for {
		iter++
		if iter > maxIter {
			return Result{Status: StatusUnstable, Evidence: "iteration limit exceeded"}
		}
		if hasNaNOrInf(tab, m, total) {
			return Result{Status: StatusUnstable, Evidence: "tableau contains NaN/Inf during iteration"}
		}
		q := entering(tab, m, total)
		if q < 0 {
			break
		}
		p := leaving(tab, m, total, q)
		if p < 0 {
			if phase1 {
				return Result{Status: StatusUnstable, Evidence: "phase1 unbounded (numeric anomaly)"}
			}
			return Result{Status: StatusUnbounded, Evidence: "objective unbounded below on column " + fmt.Sprint(q)}
		}
		pivot(tab, m, total, p, q)
		basic[p] = q
		recomputeCosts(tab, basic, m, total, costVec)
	}

	// 阶段一收尾。
	if phase1 {
		// 成本行 RHS 存的是 -z（reduced-cost 行约定），故目标值 z = -tab[m][total]。
		obj := -tab[m][total]
		if obj > tol {
			return Result{Status: StatusInfeasible, Evidence: fmt.Sprintf("phase1 objective %.3e exceeds tolerance", obj)}
		}
		// 可行：剥离人工变量列，重建 tableau 进入阶段二。
		tab, basic, m, total = phase2Setup(tab, basic, m, total, artCols, n, p.C)
		if m == 0 {
			return Result{Status: StatusUnstable, Evidence: "no rows left after phase1"}

		}
		costVec = make([]float64, total)
		copy(costVec[:n], p.C)
		recomputeCosts(tab, basic, m, total, costVec)
	}

	// 阶段二迭代。
	iter = 0
	for {
		iter++
		if iter > maxIter {
			return Result{Status: StatusUnstable, Evidence: "phase2 iteration limit exceeded"}
		}
		if hasNaNOrInf(tab, m, total) {
			return Result{Status: StatusUnstable, Evidence: "tableau contains NaN/Inf in phase2"}
		}
		q := entering(tab, m, total)
		if q < 0 {
			break
		}
		p := leaving(tab, m, total, q)
		if p < 0 {
			return Result{Status: StatusUnbounded, Evidence: "objective unbounded below on column " + fmt.Sprint(q)}
		}
		pivot(tab, m, total, p, q)
		basic[p] = q
		recomputeCosts(tab, basic, m, total, costVec)
	}

	// 提取原始变量解。
	x := make([]float64, n)
	for i := 0; i < m; i++ {
		if basic[i] < n {
			x[basic[i]] = tab[i][total]
		}
	}
	obj := 0.0
	for j := 0; j < n; j++ {
		obj += p.C[j] * x[j]
	}
	return Result{Status: StatusOptimal, X: x, Objective: obj}
}

// phase2Setup 剥离人工变量列：对仍以人工变量为基的行做基替换或删除冗余行，
// 返回 (tab, basic, m, total)，成本行已复位待 recomputeCosts。
func phase2Setup(tab [][]float64, basic []int, m, total int, artCols []bool, n int, c []float64) ([][]float64, []int, int, int) {
	// 1) 删列：跳过 art 列。
	colMap := make([]int, 0)
	for j := 0; j < total; j++ {
		if !artCols[j] {
			colMap = append(colMap, j)
		}
	}
	newTotal := n + m
	out := make([][]float64, m+1)
	for i := range out {
		out[i] = make([]float64, newTotal+1)
	}
	for i := 0; i <= m; i++ {
		for k, src := range colMap {
			out[i][k] = tab[i][src]
		}
		out[i][newTotal] = tab[i][total]
	}
	// 2) 修正基索引：art 基 -> -1。
	newBasic := make([]int, m)
	for i := 0; i < m; i++ {
		b := basic[i]
		newBasic[i] = -1
		for k, src := range colMap {
			if src == b {
				newBasic[i] = k
				break
			}
		}
	}

	// 3) 迭代处理 art 基行：找非零非 art 列入基；全零则视为冗余行删除。
	for {
		artRow := -1
		for i := 0; i < len(out)-1; i++ {
			if newBasic[i] == -1 {
				artRow = i
				break
			}
		}
		if artRow < 0 {
			break
		}
		q := -1
		for j := 0; j < newTotal; j++ {
			if math.Abs(out[artRow][j]) > tol {
				q = j
				break
			}
		}
		if q < 0 {
			// 冗余行（全零）：删除该行（含成本行同步删除）。
			out = append(out[:artRow], out[artRow+1:]...)
			newBasic = append(newBasic[:artRow], newBasic[artRow+1:]...)
			continue
		}
		pivotFull(out, newTotal, artRow, q)
		newBasic[artRow] = q
	}

	m2 := len(out) - 1
	out = out[:m2+1]
	newBasic = newBasic[:m2]
	// 成本行清零（等 recomputeCosts 写入）。
	cost := out[m2]
	for j := 0; j <= newTotal; j++ {
		cost[j] = 0
	}
	_ = c
	return out, newBasic, m2, newTotal
}

// pivotFull 在含成本行的 tableau 上做高斯消去（遍历全部行）。
func pivotFull(tab [][]float64, total, p, q int) {
	piv := tab[p][q]
	if math.Abs(piv) <= tol {
		return
	}
	for j := 0; j <= total; j++ {
		tab[p][j] /= piv
	}
	for i := 0; i < len(tab); i++ {
		if i == p {
			continue
		}
		f := tab[i][q]
		if math.Abs(f) <= tol {
			continue
		}
		for j := 0; j <= total; j++ {
			tab[i][j] -= f * tab[p][j]
		}
	}
}

// entering 按 Bland 规则选择进入列：最左侧 reduced cost < -tol 的列。
func entering(tab [][]float64, m, total int) int {
	cost := tab[m]
	for j := 0; j < total; j++ {
		if cost[j] < -tol {
			return j
		}
	}
	return -1
}

// leaving 选择离开行：ratio 最小（Bland 取最小行号）。
func leaving(tab [][]float64, m, total, q int) int {
	best := -1
	bestRatio := math.Inf(1)
	for i := 0; i < m; i++ {
		a := tab[i][q]
		if a > tol {
			r := tab[i][total] / a
			if r < bestRatio-tol || (math.Abs(r-bestRatio) <= tol && best == -1) {
				bestRatio = r
				best = i
			}
		}
	}
	return best
}

// pivot 以 (p, q) 为主元做高斯消去。
func pivot(tab [][]float64, m, total, p, q int) {
	piv := tab[p][q]
	if math.Abs(piv) <= tol {
		return
	}
	for j := 0; j <= total; j++ {
		tab[p][j] /= piv
	}
	for i := 0; i <= m; i++ {
		if i == p {
			continue
		}
		f := tab[i][q]
		if math.Abs(f) <= tol {
			continue
		}
		for j := 0; j <= total; j++ {
			tab[i][j] -= f * tab[p][j]
		}
	}
}

// recomputeCosts 计算 reduced cost：cost[j] = costVec[j] - Σ_i costVec[basic[i]] * tab[i][j]。
// tab[m] 为成本行；costVec 长度 total，对应各列目标成本。
func recomputeCosts(tab [][]float64, basic []int, m, total int, costVec []float64) {
	cost := tab[m]
	for j := 0; j <= total; j++ {
		base := 0.0
		if j < total {
			base = costVec[j]
		}
		cost[j] = base
	}
	for j := 0; j <= total; j++ {
		for i := 0; i < m; i++ {
			b := basic[i]
			if b >= 0 && b < total {
				cost[j] -= costVec[b] * tab[i][j]
			}
		}
	}
}

// hasNaNOrInf 检测 tableau 是否出现非有限值。
func hasNaNOrInf(tab [][]float64, m, total int) bool {
	for i := 0; i <= m; i++ {
		for j := 0; j <= total; j++ {
			v := tab[i][j]
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return true
			}
		}
	}
	return false
}
