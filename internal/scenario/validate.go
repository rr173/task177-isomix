package scenario

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"task177-isomix/internal/model"
)

// ValidateRequest checks user-provided what-if parameters before execution.
func ValidateRequest(request Request) error {
	if strings.TrimSpace(request.Name) == "" {
		return fmt.Errorf("scenario name is required")
	}
	if math.IsNaN(request.WidthScale) || math.IsInf(request.WidthScale, 0) || request.WidthScale <= 0 {
		return fmt.Errorf("scenario width scale must be positive and finite")
	}
	if request.WidthScale > 10 {
		return fmt.Errorf("scenario width scale exceeds 10")
	}
	if math.IsNaN(request.CenterShift) || math.IsInf(request.CenterShift, 0) {
		return fmt.Errorf("scenario center shift must be finite")
	}
	if math.IsNaN(request.ClampLow) || math.IsInf(request.ClampLow, 0) || math.IsNaN(request.ClampHigh) || math.IsInf(request.ClampHigh, 0) {
		return fmt.Errorf("scenario clamps must be finite")
	}
	if request.ClampLow > request.ClampHigh {
		return fmt.Errorf("scenario lower clamp exceeds upper clamp")
	}
	return nil
}

// ValidatePlan validates all requests and rejects duplicate names.
func ValidatePlan(plan Plan) error {
	if strings.TrimSpace(plan.Name) == "" {
		return fmt.Errorf("scenario plan name is required")
	}
	seen := make(map[string]bool, len(plan.Requests))
	for _, request := range plan.Requests {
		if err := ValidateRequest(request); err != nil {
			return fmt.Errorf("%s: %w", request.Name, err)
		}
		if seen[request.Name] {
			return fmt.Errorf("duplicate scenario name %q", request.Name)
		}
		seen[request.Name] = true
	}
	return nil
}

// CanonicalizeBounds returns a deep copy with ordered interval endpoints.
func CanonicalizeBounds(bounds map[string]model.Range) map[string]model.Range {
	ids := make([]string, 0, len(bounds))
	for id := range bounds {
		ids = append(ids, id)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	out := make(map[string]model.Range, len(ids))
	for _, id := range ids {
		out[id] = normalize(bounds[id])
	}
	return out
}

// MissingIDs reports IDs present in one projection but absent in another.
func MissingIDs(first, second map[string]model.Range) []string {
	missing := make([]string, 0)
	for id := range first {
		if _, ok := second[id]; !ok {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	return missing
}

// Equivalent compares interval maps with a numeric tolerance.
func Equivalent(first, second map[string]model.Range, tolerance float64) bool {
	if len(first) != len(second) {
		return false
	}
	if tolerance < 0 || math.IsNaN(tolerance) {
		tolerance = 0
	}
	for id, left := range first {
		right, ok := second[id]
		if !ok || math.Abs(left.Lo-right.Lo) > tolerance || math.Abs(left.Hi-right.Hi) > tolerance {
			return false
		}
	}
	return true
}
