package solve

import (
	"task177-isomix/internal/constraint"
	"task177-isomix/internal/lp"
	"task177-isomix/internal/model"
)

// rowMeta 记录 LP 中每一行对应的约束来源，供活跃约束与冲突核心映射回领域对象。
type rowMeta struct {
	Label  string // 行标签
	Origin string // 关联约束 ID（空 = 内置质量守恒/同位素区间）
}

// Compiled 一次求解的编译产物。
type Compiled struct {
	Problem *lp.Problem
	Rows    []rowMeta
	// Builder 保存端元/样品上下文，供报告层复用。
	Builder *constraint.Builder
}

// BuildProblem 把端元、样品与启用约束编译为 LP 问题。
// 行顺序：质量守恒、同位素区间、用户约束（每个约束可能展开为多行）。
func BuildProblem(b *constraint.Builder, userConstraints []model.Constraint) (*Compiled, error) {
	rows, err := b.Compile(userConstraints)
	if err != nil {
		return nil, err
	}
	n := b.NumVars()
	m := 0
	for _, r := range rows {
		if r.Sense == 0 {
			m += 2
		} else {
			m += 1
		}
	}
	p := &lp.Problem{
		A: make([][]float64, 0, m),
		B: make([]float64, 0, m),
		C: make([]float64, n),
	}
	meta := make([]rowMeta, 0, m)
	for _, r := range rows {
		switch r.Sense {
		case 0: // 等式：拆两条不等式
			p.A = append(p.A, clone(r.Coefficients))
			p.B = append(p.B, r.B)
			meta = append(meta, rowMeta{Label: r.Label + "_eq_le", Origin: r.Origin})
			neg := make([]float64, n)
			for i, v := range r.Coefficients {
				neg[i] = -v
			}
			p.A = append(p.A, neg)
			p.B = append(p.B, -r.B)
			meta = append(meta, rowMeta{Label: r.Label + "_eq_ge", Origin: r.Origin})
		case 1: // a x <= b
			p.A = append(p.A, clone(r.Coefficients))
			p.B = append(p.B, r.B)
			meta = append(meta, rowMeta{Label: r.Label, Origin: r.Origin})
		case -1: // a x >= b -> -a x <= -b
			neg := make([]float64, n)
			for i, v := range r.Coefficients {
				neg[i] = -v
			}
			p.A = append(p.A, neg)
			p.B = append(p.B, -r.B)
			meta = append(meta, rowMeta{Label: r.Label, Origin: r.Origin})
		}
	}
	return &Compiled{Problem: p, Rows: meta, Builder: b}, nil
}

func clone(r []float64) []float64 {
	out := make([]float64, len(r))
	copy(out, r)
	return out
}

// NonZeros 统计矩阵非零元数量（供 MatrixSummary）。
func NonZeros(p *lp.Problem) int {
	cnt := 0
	for _, row := range p.A {
		for _, v := range row {
			if v != 0 {
				cnt++
			}
		}
	}
	return cnt
}
