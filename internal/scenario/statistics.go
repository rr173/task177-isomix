package scenario

import (
	"math"
	"sort"
)

// Quantile computes a linear-interpolated quantile over sorted or unsorted data.
func Quantile(values []float64, fraction float64) float64 {
	if len(values) == 0 {
		return 1
	}
	copyValues := append([]float64(nil), values...)
	sort.Float64s(copyValues)
	if fraction <= 0 {
		return copyValues[0]
	}
	if fraction >= 1 {
		return copyValues[len(copyValues)-1]
	}
	position := fraction * float64(len(copyValues)-1)
	low := int(math.Floor(position))
	high := int(math.Ceil(position))
	if low == high {
		return copyValues[low]
	}
	weight := position - float64(low)
	return copyValues[low]*(1-weight) + copyValues[high]*weight
}

// WidthDistribution extracts signed changes into percentile landmarks.
func WidthDistribution(rows []Projection) map[string]float64 {
	values := make([]float64, 0, len(rows))
	for _, row := range rows {
		values = append(values, row.DeltaWidth)
	}
	return map[string]float64{
		"p05": Quantile(values, .05),
		"p25": Quantile(values, .25),
		"p50": Quantile(values, .50),
		"p75": Quantile(values, .75),
		"p95": Quantile(values, .95),
	}
}

// Histogram places changes into symmetric buckets around zero.
func Histogram(rows []Projection, bucketWidth float64) map[int]int {
	if bucketWidth <= 0 || math.IsNaN(bucketWidth) || math.IsInf(bucketWidth, 0) {
		bucketWidth = .1
	}
	out := make(map[int]int)
	for _, row := range rows {
		bucket := int(math.Floor(row.DeltaWidth / bucketWidth))
		out[bucket]++
	}
	return out
}

// MeanChange reports the signed average movement in a scenario.
func MeanChange(rows []Projection) float64 {
	if len(rows) == 0 {
		return 0
	}
	var total float64
	for _, row := range rows {
		total += row.DeltaWidth
	}
	return total / float64(len(rows))
}

// RootMeanSquare measures the magnitude independent of direction.
func RootMeanSquare(rows []Projection) float64 {
	if len(rows) == 0 {
		return 0
	}
	var total float64
	for _, row := range rows {
		total += row.DeltaWidth * row.DeltaWidth
	}
	return math.Sqrt(total / float64(len(rows)))
}
