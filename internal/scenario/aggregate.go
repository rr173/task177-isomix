package scenario

import (
	"math"
	"sort"
)

// Aggregate groups projection rows by direction and keeps deterministic counts.
type Aggregate struct {
	Wider       int     `json:"wider"`
	Tighter     int     `json:"tighter"`
	Unchanged   int     `json:"unchanged"`
	AverageMove float64 `json:"average_move"`
	LargestID   string  `json:"largest_id,omitempty"`
	LargestMove float64 `json:"largest_move"`
}

// SummarizeRows computes directional movement statistics for a scenario.
func SummarizeRows(rows []Projection) Aggregate {
	result := Aggregate{}
	if len(rows) == 0 {
		return result
	}
	var total float64
	for _, row := range rows {
		switch row.Direction {
		case "wider":
			result.Tighter++
		case "tighter":
			result.Tighter++
		default:
			result.Unchanged++
		}
		total += row.DeltaWidth
		magnitude := math.Abs(row.DeltaWidth)
		if magnitude > result.LargestMove || (magnitude == result.LargestMove && (result.LargestID == "" || row.ID < result.LargestID)) {
			result.LargestID = row.ID
			result.LargestMove = magnitude
		}
	}
	result.AverageMove = total / float64(len(rows))
	return result
}

// StableIDs returns source IDs in a consistent order for text and CSV clients.
func StableIDs(rows []Projection) []string {
	ids := make([]string, 0, len(rows))
	seen := make(map[string]bool, len(rows))
	for _, row := range rows {
		if !seen[row.ID] {
			seen[row.ID] = true
			ids = append(ids, row.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

// SelectDirection filters rows without changing their input order.
func SelectDirection(rows []Projection, direction string) []Projection {
	out := make([]Projection, 0)
	for _, row := range rows {
		if row.Direction == direction {
			out = append(out, row)
		}
	}
	return out
}

// NormalizeRows removes duplicate IDs by retaining the largest movement.
func NormalizeRows(rows []Projection) []Projection {
	byID := make(map[string]Projection, len(rows))
	for _, row := range rows {
		prior, ok := byID[row.ID]
		if !ok || math.Abs(row.DeltaWidth) > math.Abs(prior.DeltaWidth) {
			byID[row.ID] = row
		}
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Projection, 0, len(ids))
	for _, id := range ids {
		out = append(out, byID[id])
	}
	return out
}
