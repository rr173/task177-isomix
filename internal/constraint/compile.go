package constraint

import (
	"fmt"
	"sort"
	"strings"

	"task177-isomix/internal/model"
)

// Compiled 编译产物：由约束生成的线性不等式（稠密行），变量索引对应端元排序。
type Compiled struct {
	// Coefficients 行向量 a，满足 a·x <= b（或 a·x >= b，见 Sense）。
	Coefficients []float64
	B            float64
	// Sense: -1 表示 a·x >= b；+1 表示 a·x <= b；0 表示等式 a·x == b。
	Sense int
	// Label 供冲突核心报告引用的约束描述。
	Label string
	// Origin 关联约束 ID（空表示质量守恒/区间等内置约束）。
	Origin string
}

// Builder 把端元、样品与约束编译为线性规划的行集合。
type Builder struct {
	// IsotopeNames 同位素顺序（按字典序稳定）。
	IsotopeNames []string
	// EndmemberIDs 端元 ID（按字典序稳定，变量顺序与之对应）。
	EndmemberIDs []string
	// endmembers 端元 ID -> 端元，供区间查询。
	endmembers map[string]model.Endmember
	sample     model.Sample
}

// NewBuilder 构造编译器。要求端元内部维度一致且与样品维度一致（调用方保证）。
func NewBuilder(ems []model.Endmember, sample model.Sample) (*Builder, error) {
	isoSet := map[string]bool{}
	for _, e := range ems {
		for k := range e.Components {
			isoSet[k] = true
		}
	}
	for k := range sample.Measurements {
		isoSet[k] = true
	}
	names := make([]string, 0, len(isoSet))
	for k := range isoSet {
		names = append(names, k)
	}
	sort.Strings(names)

	ids := make([]string, 0, len(ems))
	byID := make(map[string]model.Endmember, len(ems))
	for _, e := range ems {
		ids = append(ids, e.ID)
		byID[e.ID] = e
	}
	sort.Strings(ids)

	return &Builder{IsotopeNames: names, EndmemberIDs: ids, endmembers: byID, sample: sample}, nil
}

// NumVars 返回变量数（端元数）。
func (b *Builder) NumVars() int { return len(b.EndmemberIDs) }

// endmemberIndex 返回端元 ID 对应的变量索引；不存在返回 -1。
func (b *Builder) endmemberIndex(id string) int {
	for i, eid := range b.EndmemberIDs {
		if eid == id {
			return i
		}
	}
	return -1
}

// MassConservation 质量守恒：Σ x_i = 1（拆为 >= 1 与 <= 1 两条）。
func (b *Builder) MassConservation() []Compiled {
	n := b.NumVars()
	row := make([]float64, n)
	for i := range row {
		row[i] = 1
	}
	return []Compiled{
		{Coefficients: cloneRow(row), B: 1, Sense: 0, Label: "mass_conservation"},
	}
}

// IsotopeInterval 同位素区间约束：对每个同位素 k，
// Σ_i x_i * E_i(k).Lo <= S(k).Hi（保证混合体不超出样品上界）
// 且 Σ_i x_i * E_i(k).Hi >= S(k).Lo（混合体不低于样品下界）。
// 用端元组成区间的保守端点构建，保证不误判可行域。
func (b *Builder) IsotopeInterval() []Compiled {
	var rows []Compiled
	for _, k := range b.IsotopeNames {
		sr, ok := b.sample.Measurements[k]
		if !ok {
			continue
		}
		rowUp := make([]float64, b.NumVars())
		rowLo := make([]float64, b.NumVars())
		for i, eid := range b.EndmemberIDs {
			er, ok := b.endmembers[eid].Components[k]
			if !ok {
				// 端元缺该同位素 -> 无法建模，调用方应在维度校验中拦截。
				continue
			}
			rowUp[i] = er.Hi
			rowLo[i] = er.Lo
		}
		rows = append(rows,
			Compiled{Coefficients: rowUp, B: sr.Hi, Sense: 1, Label: fmt.Sprintf("isotope_%s_upper", k)},
			Compiled{Coefficients: rowLo, B: sr.Lo, Sense: -1, Label: fmt.Sprintf("isotope_%s_lower", k)},
		)
	}
	return rows
}

// NonNegativity 非负约束：x_i >= 0（LP 标准型自带，返回空行集）。
func (b *Builder) NonNegativity() []Compiled { return nil }

// CompileUserConstraint 把一条用户约束编译为一行（或两行）不等式。
func (b *Builder) CompileUserConstraint(c model.Constraint) ([]Compiled, error) {
	switch c.Type {
	case model.ConstraintExclude:
		idx := b.endmemberIndex(c.Target)
		if idx < 0 {
			return nil, model.NewError("BAD_REQUEST", "exclude target endmember %q not in candidate set", c.Target)
		}
		row := make([]float64, b.NumVars())
		row[idx] = 1
		return []Compiled{{Coefficients: row, B: 0, Sense: 0, Label: c.Name, Origin: c.ID}}, nil

	case model.ConstraintBound:
		idx := b.endmemberIndex(c.Target)
		if idx < 0 {
			return nil, model.NewError("BAD_REQUEST", "bound target endmember %q not in candidate set", c.Target)
		}
		row := make([]float64, b.NumVars())
		row[idx] = 1
		return []Compiled{{Coefficients: row, B: c.Range.Hi, Sense: 1, Label: c.Name, Origin: c.ID}}, nil

	case model.ConstraintRatio:
		parts := strings.Split(c.Target, ":")
		a, bb := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		ia := b.endmemberIndex(a)
		ib := b.endmemberIndex(bb)
		if ia < 0 || ib < 0 {
			return nil, model.NewError("BAD_REQUEST", "ratio target references unknown endmember %q", c.Target)
		}
		// x_a / x_b <= hi  =>  x_a - hi*x_b <= 0
		rowUp := make([]float64, b.NumVars())
		rowUp[ia] = 1
		rowUp[ib] = -c.Range.Hi
		// x_a / x_b >= lo  =>  lo*x_b - x_a <= 0
		rowLo := make([]float64, b.NumVars())
		rowLo[ib] = c.Range.Lo
		rowLo[ia] = -1
		return []Compiled{
			{Coefficients: rowUp, B: 0, Sense: 1, Label: c.Name + "_upper", Origin: c.ID},
			{Coefficients: rowLo, B: 0, Sense: 1, Label: c.Name + "_lower", Origin: c.ID},
		}, nil
	}
	return nil, model.NewError("BAD_REQUEST", "unknown constraint type %q", c.Type)
}

// Compile 编译全部输入：质量守恒 + 同位素区间 + 用户启用约束。
func (b *Builder) Compile(userConstraints []model.Constraint) ([]Compiled, error) {
	rows := b.MassConservation()
	rows = append(rows, b.IsotopeInterval()...)
	for _, c := range userConstraints {
		cr, err := b.CompileUserConstraint(c)
		if err != nil {
			return nil, err
		}
		rows = append(rows, cr...)
	}
	return rows, nil
}

func cloneRow(r []float64) []float64 {
	out := make([]float64, len(r))
	copy(out, r)
	return out
}
