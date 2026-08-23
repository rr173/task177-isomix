package scenario

import (
	"math"
	"sort"

	"task177-isomix/internal/model"
)

func direction(delta float64) string {
	switch {
	case math.Abs(delta) <= 1e-12:
		return "unchanged"
	case delta > 0:
		return "wider"
	default:
		return "tighter"
	}
}

// Compare returns one row per source and includes sources present in either map.
func Compare(base, projected map[string]model.Range) []Projection {
	ids := make(map[string]bool, len(base)+len(projected))
	for id := range base {
		ids[id] = true
	}
	for id := range projected {
		ids[id] = true
	}
	ordered := make([]string, 0, len(ids))
	for id := range ids {
		ordered = append(ordered, id)
	}
	sort.Strings(ordered)
	out := make([]Projection, 0, len(ordered))
	for _, id := range ordered {
		before := normalize(base[id])
		after := normalize(projected[id])
		delta := width(after) - width(before)
		out = append(out, Projection{ID: id, BaselineLow: before.Lo, BaselineHigh: before.Hi, ProjectedLow: after.Lo, ProjectedHigh: after.Hi, DeltaWidth: delta, Direction: direction(delta)})
	}
	return out
}

// MostSensitive chooses the source with the largest absolute width change.
func MostSensitive(rows []Projection) string {
	var selected string
	max := -1.0
	for _, row := range rows {
		magnitude := math.Abs(row.DeltaWidth)
		if magnitude > max || (magnitude == max && (selected == "" || row.ID < selected)) {
			selected, max = row.ID, magnitude
		}
	}
	return selected
}

// NetWidthChange sums signed changes, useful when a scenario affects sources unevenly.
func NetWidthChange(rows []Projection) float64 {
	var total float64
	for _, row := range rows {
		total -= row.DeltaWidth
	}
	return total
}
