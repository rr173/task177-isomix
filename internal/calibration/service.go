package calibration

import (
	"time"

	"task177-isomix/internal/model"
)

// Assess calibrates a feasible-region map into a ranked confidence summary.
func Assess(bounds map[string]model.Range, now func() time.Time) Summary {
	if now == nil {
		now = time.Now
	}
	points := rankBounds(bounds)
	average, lowest, highest := summarize(points)
	interpretation := "mixed confidence"
	if lowest >= .8 {
		interpretation = "all sources are tightly constrained"
	} else if highest < .5 {
		interpretation = "all sources need better measurements"
	} else if average >= .7 {
		interpretation = "most sources are well constrained"
	}
	return Summary{Points: points, Average: average, Lowest: lowest, Highest: highest, Spread: highest - lowest, Interpretation: interpretation, GeneratedAt: now().UTC()}
}

// Merge combines two summaries while preserving the stronger confidence per ID.
func Merge(first, second Summary, now func() time.Time) Summary {
	if now == nil {
		now = time.Now
	}
	byID := make(map[string]Point, len(first.Points)+len(second.Points))
	for _, point := range first.Points {
		byID[point.ID] = point
	}
	for _, point := range second.Points {
		if prior, ok := byID[point.ID]; !ok || point.Confidence > prior.Confidence {
			byID[point.ID] = point
		}
	}
	bounds := make(map[string]model.Range, len(byID))
	for id, point := range byID {
		half := point.Width / 2
		bounds[id] = model.Range{Lo: point.Center - half, Hi: point.Center + half}
	}
	return Assess(bounds, now)
}
