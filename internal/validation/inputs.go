package validation

import (
	"strings"

	"task177-isomix/internal/model"
)

// CleanName trims user labels while preserving the original value for errors.
func CleanName(name string) (string, bool) {
	clean := strings.Join(strings.Fields(name), " ")
	return clean, clean != ""
}

// ConstraintTarget checks the textual shape before a domain service resolves IDs.
func ConstraintTarget(c model.Constraint) bool {
	target := strings.TrimSpace(c.Target)
	if target == "" {
		return false
	}
	if c.Type == model.ConstraintRatio {
		parts := strings.Split(target, ":")
		return len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != ""
	}
	return !strings.ContainsAny(target, " \t\n")
}
