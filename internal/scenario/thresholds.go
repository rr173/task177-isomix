package scenario

import (
	"math"
	"sort"
)

// Thresholds summarizes how many sources cross user-visible movement limits.
type Thresholds struct {
	Small    int `json:"small"`
	Moderate int `json:"moderate"`
	Large    int `json:"large"`
	Extreme  int `json:"extreme"`
}

// ClassifyMovement maps an absolute width change to a stable bucket.
func ClassifyMovement(change float64) string {
	change = math.Abs(change)
	switch {
	case change <= .01:
		return "small"
	case change < .05:
		return "moderate"
	case change < .2:
		return "large"
	default:
		return "extreme"
	}
}

// CountThresholds counts movement buckets for a scenario result.
func CountThresholds(rows []Projection) Thresholds {
	var result Thresholds
	for _, row := range rows {
		switch ClassifyMovement(row.DeltaWidth) {
		case "small":
			result.Small++
		case "moderate":
			result.Moderate++
		case "large":
			result.Large++
		case "extreme":
			result.Extreme++
		}
	}
	return result
}

// LargestChanges returns up to limit rows ordered by absolute change.
func LargestChanges(rows []Projection, limit int) []Projection {
	copyRows := append([]Projection(nil), rows...)
	sort.SliceStable(copyRows, func(i, j int) bool {
		left, right := math.Abs(copyRows[i].DeltaWidth), math.Abs(copyRows[j].DeltaWidth)
		if left == right {
			return copyRows[i].ID < copyRows[j].ID
		}
		return left > right
	})
	if limit < 0 {
		limit = 0
	}
	if len(copyRows) > limit {
		copyRows = copyRows[:limit]
	}
	return copyRows
}
