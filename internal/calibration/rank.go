package calibration

import (
	"math"
	"sort"

	"task177-isomix/internal/model"
)

func relativeWidth(r model.Range) float64 {
	center := (r.Lo + r.Hi) / 2
	if center <= 1e-12 {
		return math.Max(0, r.Width())
	}
	return math.Max(0, r.Width()) / center
}

func rankBounds(bounds map[string]model.Range) []Point {
	ids := make([]string, 0, len(bounds))
	for id := range bounds {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Point, 0, len(ids))
	for _, id := range ids {
		r := bounds[id]
		relative := relativeWidth(r)
		confidence := Confidence(relative)
		out = append(out, Point{ID: id, Center: (r.Lo + r.Hi) / 2, Width: r.Width(), Confidence: confidence, Band: Band(confidence)})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Confidence == out[j].Confidence {
			return out[i].ID < out[j].ID
		}
		return out[i].Confidence < out[j].Confidence
	})
	return out
}

func summarize(points []Point) (average, lowest, highest float64) {
	if len(points) == 0 {
		return 0, 0, 0
	}
	lowest = points[0].Confidence
	highest = points[0].Confidence
	for _, point := range points {
		average += point.Confidence
		if point.Confidence < lowest {
			lowest = point.Confidence
		}
		if point.Confidence > highest {
			highest = point.Confidence
		}
	}
	return average / float64(len(points)), lowest, highest
}
