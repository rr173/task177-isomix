package validation

import (
	"fmt"
	"sort"

	"task177-isomix/internal/model"
)

// MissingDimensions formats a deterministic message for API clients.
func MissingDimensions(expected, actual map[string]model.Range) []string {
	missing := make([]string, 0)
	for name := range expected {
		if _, ok := actual[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(missing)))
	return missing
}

// DescribeRange returns a short diagnostic used by validation endpoints.
func DescribeRange(name string, r model.Range) string {
	if !FiniteRange(r) {
		return fmt.Sprintf("%s is not finite", name)
	}
	if !r.Valid() {
		return fmt.Sprintf("%s is reversed", name)
	}
	return fmt.Sprintf("%s spans %.6g", name, r.Width())
}
