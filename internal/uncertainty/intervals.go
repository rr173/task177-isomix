package uncertainty

import (
	"math"
	"sort"

	"task177-isomix/internal/model"
)

func intervalStats(values map[string]model.Range) (mean, maxRelative float64, names []string) {
	names = make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		r := values[name]
		mean += math.Max(0, r.Width())
		center := math.Abs((r.Lo + r.Hi) / 2)
		rel := r.Width()
		if center > 1e-12 {
			rel = r.Width() / center
		}
		if rel > maxRelative {
			maxRelative = rel
		}
	}
	if len(names) > 0 {
		mean /= float64(len(names))
	}
	return mean, maxRelative, names
}

func classify(maxRelative float64, positive bool) string {
	if !positive {
		return "invalid-covariance"
	}
	if maxRelative <= 0.1 {
		return "high-confidence"
	}
	if maxRelative <= 0.5 {
		return "usable"
	}
	return "overspread"
}
