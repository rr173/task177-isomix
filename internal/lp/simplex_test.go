package lp

import (
	"math"
	"testing"
)

func mustOptimal(t *testing.T, p *Problem) Result {
	t.Helper()
	r := Solve(p)
	if r.Status != StatusOptimal {
		t.Fatalf("expected optimal, got %s (%s)", r.Status, r.Evidence)
	}
	return r
}

func almost(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// 基础：min x0 s.t. x0+x1+x2 = 1, x>=0  -> min = 0，最优解 x0=0。
func TestSimplexMassConservation(t *testing.T) {
	p := &Problem{
		A: [][]float64{
			{1, 1, 1},
			{-1, -1, -1},
		},
		B: []float64{1, -1},
		C: []float64{1, 0, 0},
	}
	r := mustOptimal(t, p)
	if !almost(r.Objective, 0) {
		t.Fatalf("objective = %v, want 0", r.Objective)
	}
	sum := r.X[0] + r.X[1] + r.X[2]
	if !almost(sum, 1) {
		t.Fatalf("sum = %v, want 1", sum)
	}
}

// 有界：min x0 s.t. x0+x1+x2=1, x0>=0.4  -> min = 0.4。
func TestSimplexLowerBound(t *testing.T) {
	p := &Problem{
		A: [][]float64{
			{1, 1, 1},
			{-1, -1, -1},
			{-1, 0, 0},
		},
		B: []float64{1, -1, -0.4},
		C: []float64{1, 0, 0},
	}
	r := mustOptimal(t, p)
	if !almost(r.Objective, 0.4) {
		t.Fatalf("objective = %v, want 0.4", r.Objective)
	}
	if !almost(r.X[0], 0.4) {
		t.Fatalf("x0 = %v, want 0.4", r.X[0])
	}
}

// 最大：max x0 == min -x0, s.t. x0+x1+x2=1 -> max = 1。
func TestSimplexMax(t *testing.T) {
	p := &Problem{
		A: [][]float64{
			{1, 1, 1},
			{-1, -1, -1},
		},
		B: []float64{1, -1},
		C: []float64{-1, 0, 0},
	}
	r := mustOptimal(t, p)
	if !almost(r.Objective, -1) {
		t.Fatalf("objective = %v, want -1", r.Objective)
	}
	if !almost(r.X[0], 1) {
		t.Fatalf("x0 = %v, want 1", r.X[0])
	}
}

// 排除：x2 = 0 强制，min x0 s.t. x0+x1+x2=1, x2=0 -> min x0 = 0。
func TestSimplexExclude(t *testing.T) {
	p := &Problem{
		A: [][]float64{
			{1, 1, 1},
			{-1, -1, -1},
			{0, 0, 1},
			{0, 0, -1},
		},
		B: []float64{1, -1, 0, 0},
		C: []float64{1, 0, 0},
	}
	r := mustOptimal(t, p)
	if !almost(r.Objective, 0) {
		t.Fatalf("objective = %v, want 0", r.Objective)
	}
	if !almost(r.X[2], 0) {
		t.Fatalf("x2 = %v, want 0", r.X[2])
	}
}

// 不可行：x0+x1=1 且 x0>=0.6, x1>=0.6 -> infeasible。
func TestSimplexInfeasible(t *testing.T) {
	p := &Problem{
		A: [][]float64{
			{1, 1},
			{-1, -1},
			{-1, 0},
			{0, -1},
		},
		B: []float64{1, -1, -0.6, -0.6},
		C: []float64{0, 0},
	}
	r := Solve(p)
	if r.Status != StatusInfeasible {
		t.Fatalf("expected infeasible, got %s (%s)", r.Status, r.Evidence)
	}
}

// 同位素场景：min x_organic，三端元 + 样品区间约束（与 smoke 场景一致）。
func TestSimplexIsotopeScenario(t *testing.T) {
	// 端元顺序 mantle, crust, organic；同位素 d18O, d2H。
	// 样品：d18O [6.0, 8.0], d2H [-85, -50]（落在混合可行域内部）。
	// 质量守恒 x0+x1+x2=1
	// d18O: x0*5.8 + x1*9.5 + x2*18 <= 8.0 ; x0*5.2 + x1*8 + x2*14 >= 6.0
	// d2H:  x0*(-70) + x1*(-40) + x2*(-110) <= -50 ; x0*(-90)+x1*(-60)+x2*(-160) >= -85
	//   -> -Σ x_i*Lo_d2H <= 85，即 Σ x_i*{90,60,160} <= 85
	p := &Problem{
		A: [][]float64{
			{1, 1, 1},
			{-1, -1, -1},
			{5.8, 9.5, 18},
			{-5.2, -8, -14},
			{-70, -40, -110},
			{90, 60, 160},
		},
		B: []float64{1, -1, 8.0, -6.0, -50, 85},
		C: []float64{1, 0, 0}, // min x_organic(变量2)
	}
	r := mustOptimal(t, p)
	if r.X[2] < -1e-6 {
		t.Fatalf("x_organic min = %v, want >= 0", r.X[2])
	}
	// 检查解满足约束。
	x := r.X
	if !almost(x[0]+x[1]+x[2], 1) {
		t.Fatalf("mass conservation violated: %v", x)
	}
	if x[0] < -1e-6 || x[1] < -1e-6 || x[2] < -1e-6 {
		t.Fatalf("non-negativity violated: %v", x)
	}
}

// 排除约束：x2 = 0 强制后 min x_organic 仍为 0 且 bounds 坍缩。
func TestSimplexExcludeShrinksBounds(t *testing.T) {
	p := &Problem{
		A: [][]float64{
			{1, 1, 1},
			{-1, -1, -1},
			{5.8, 9.5, 18},
			{-5.2, -8, -14},
			{-70, -40, -110},
			{90, 60, 160},
			{0, 0, 1},
			{0, 0, -1},
		},
		B: []float64{1, -1, 8.0, -6.0, -50, 85, 0, 0},
		C: []float64{1, 0, 0},
	}
	r := mustOptimal(t, p)
	if !almost(r.X[2], 0) {
		t.Fatalf("excluded endmember x2 = %v, want 0", r.X[2])
	}
	// max x2 也应为 0。
	p2 := &Problem{A: p.A, B: p.B, C: []float64{-1, 0, 0}}
	r2 := mustOptimal(t, p2)
	if !almost(r2.X[2], 0) {
		t.Fatalf("excluded endmember max x2 = %v, want 0", r2.X[2])
	}
}
