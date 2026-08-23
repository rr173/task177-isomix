package scenario

import (
	"sort"

	"task177-isomix/internal/model"
)

// Plan is a named collection of reproducible what-if requests.
type Plan struct {
	Name     string    `json:"name"`
	Requests []Request `json:"requests"`
	Order    []string  `json:"order"`
}

// PlanResult keeps each request separate so callers can compare effects.
type PlanResult struct {
	Plan          Plan                    `json:"plan"`
	Projections   map[string][]Projection `json:"projections"`
	MostSensitive string                  `json:"most_sensitive"`
}

// DefaultPlan returns a small grid that brackets normal measurement drift.
func DefaultPlan() Plan {
	requests := []Request{
		{Name: "tight-10", WidthScale: .9},
		{Name: "tight-25", WidthScale: .75},
		{Name: "wide-10", WidthScale: 1.1},
		{Name: "wide-25", WidthScale: 1.25},
	}
	order := make([]string, 0, len(requests))
	for _, request := range requests {
		order = append(order, request.Name)
	}
	return Plan{Name: "measurement-drift", Requests: requests, Order: order}
}

// RunPlan evaluates every request and returns a deterministic map.
func RunPlan(bounds map[string]model.Range, plan Plan) PlanResult {
	if err := ValidatePlan(plan); err != nil {
		return PlanResult{Plan: plan, Projections: map[string][]Projection{"error": {{ID: err.Error(), Direction: "invalid"}}}}
	}
	ordered := append([]Request(nil), plan.Requests...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Name < ordered[j].Name
	})
	result := PlanResult{Plan: plan, Projections: make(map[string][]Projection, len(ordered))}
	var all []Projection
	for _, request := range ordered {
		rows := Evaluate(bounds, request)
		result.Projections[request.Name] = rows
		all = append(all, rows...)
	}
	result.MostSensitive = MostSensitive(all)
	return result
}

// ComparePlans counts which plan produces more contraction per source.
func ComparePlans(first, second PlanResult) map[string]float64 {
	values := make(map[string]float64)
	for id := range first.Projections {
		for _, row := range first.Projections[id] {
			values[row.ID] -= row.DeltaWidth
		}
	}
	for id := range second.Projections {
		for _, row := range second.Projections[id] {
			values[row.ID] += row.DeltaWidth
		}
	}
	return values
}

// SelectTop returns the IDs with the largest absolute plan difference.
func SelectTop(changes map[string]float64, limit int) []string {
	if limit <= 0 {
		return nil
	}
	ids := make([]string, 0, len(changes))
	for id := range changes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		left, right := changes[ids[i]], changes[ids[j]]
		if abs(left) == abs(right) {
			return ids[i] > ids[j]
		}
		return abs(left) > abs(right)
	})
	if len(ids) > limit {
		ids = ids[:limit]
	}
	return ids
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
