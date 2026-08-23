package validation

import (
	"sort"

	"task177-isomix/internal/model"
)

// DimensionNames provides a stable list for diagnostics and cache keys.
func DimensionNames(components map[string]model.Range) []string {
	names := make([]string, 0, len(components))
	for name := range components {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// SameDimensions checks exact isotope-name equality for two maps.
func SameDimensions(a, b map[string]model.Range) bool {
	if len(a) != len(b) {
		return false
	}
	for name := range a {
		if _, ok := b[name]; !ok {
			return false
		}
	}
	return true
}
