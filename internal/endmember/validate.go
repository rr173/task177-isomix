package endmember

import (
	"math"
	"sort"

	"task177-isomix/internal/model"
)

// ValidateComponents 校验组成区间：
//  1. 至少一个同位素维度；
//  2. 每个区间合法（Lo <= Hi）；
//  3. 拒绝 NaN / Inf。
//
// 注意：同位素 δ 值（千分偏差）可为负（如 d2H = -90‰），因此不做非负强制；
// “负质量”错误只用于比例类约束（见 constraint 包）。
//
// 返回模型错误，保证错误码稳定。
func ValidateComponents(components map[string]model.Range) error {
	if len(components) == 0 {
		return model.NewError("BAD_REQUEST", "at least one isotope component is required")
	}
	names := make([]string, 0, len(components))
	for k := range components {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, name := range names {
		r := components[name]
		if math.IsNaN(r.Lo) || math.IsNaN(r.Hi) || math.IsInf(r.Lo, 0) || math.IsInf(r.Hi, 0) {
			return model.NewError("BAD_REQUEST", "component %q has non-finite bound", name)
		}
		if !r.Valid() {
			return model.NewError("INVALID_RANGE", "component %q range [%v,%v] is empty or reversed", name, r.Lo, r.Hi)
		}
	}
	return nil
}

// CheckDimension 检查一组端元与样品是否共享同一同位素维度（同位素名集合一致）。
// 返回缺失/多余维度列表，便于错误信息定位。
func CheckDimension(ems []model.Endmember, sample model.Sample) (missing, extra []string) {
	dim := sampleDimension(sample)
	union := map[string]bool{}
	for _, e := range ems {
		for k := range e.Components {
			union[k] = true
		}
	}
	// 样品有但端元没有 -> 缺失（无法建模）；端元有但样品没有 -> 多余。
	for k := range dim {
		if !union[k] {
			missing = append(missing, k)
		}
	}
	for k := range union {
		if !dim[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}

func sampleDimension(s model.Sample) map[string]bool {
	out := make(map[string]bool, len(s.Measurements))
	for k := range s.Measurements {
		out[k] = true
	}
	return out
}

// Consistent 判断所有端元内部维度一致（同位素名集合相同）。
func Consistent(ems []model.Endmember) bool {
	if len(ems) == 0 {
		return true
	}
	base := keySet(ems[0].Components)
	for _, e := range ems[1:] {
		if !sameSet(base, keySet(e.Components)) {
			return false
		}
	}
	return true
}

func keySet(m map[string]model.Range) map[string]bool {
	out := make(map[string]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}

func sameSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}
