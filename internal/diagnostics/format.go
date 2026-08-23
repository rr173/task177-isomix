package diagnostics

import (
	"fmt"
	"strings"
)

// Humanize turns a report into stable text suitable for lab notes.
func Humanize(report Report) string {
	lines := []string{
		fmt.Sprintf("solution %s: %s", report.SolutionID, strings.ToUpper(report.Risk)),
		fmt.Sprintf("matrix %d x %d, density %.4f", report.ConstraintRows, report.Variables, report.Density),
	}
	lines = append(lines, report.Observations...)
	return strings.Join(lines, "; ")
}
