package sensitivity

import (
	"math"
	"sort"

	"task177-isomix/internal/model"
)

func summarizeBounds(bounds map[string]model.Range) ([]BoundSummary, float64, float64) {
	ids := make([]string, 0, len(bounds))
	for id := range bounds {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]BoundSummary, 0, len(ids))
	var total, relative float64
	for _, id := range ids {
		r := bounds[id]
		center := (r.Lo + r.Hi) / 2
		width := math.Max(0, r.Hi-r.Lo)
		rel := width
		if math.Abs(center) > 1e-12 {
			rel = width / math.Abs(center)
		}
		out = append(out, BoundSummary{EndmemberID: id, Low: r.Lo, High: r.Hi, Center: center, Width: width, Relative: rel})
		total += width
		relative += rel
	}
	if len(out) > 0 {
		relative /= float64(len(out))
	}
	return out, total, relative
}

func classify(total, relative float64, feasible bool) string {
	if !feasible {
		return "infeasible"
	}
	if total <= 1e-9 && relative <= 1e-9 {
		return "tight"
	}
	if total <= 0.25 && relative <= 0.5 {
		return "stable"
	}
	if total <= 0.75 {
		return "sensitive"
	}
	return "diffuse"
}
