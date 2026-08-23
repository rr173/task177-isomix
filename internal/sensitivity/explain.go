package sensitivity

import (
	"fmt"
	"sort"
)

// Explain creates a concise human-readable ordering for a report. It is used
// by the HTTP layer and kept separate so JSON consumers do not depend on text.
func Explain(report *Report) []string {
	if report == nil || len(report.Bounds) == 0 {
		return nil
	}
	ordered := append([]BoundSummary(nil), report.Bounds...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Width == ordered[j].Width {
			return ordered[i].EndmemberID < ordered[j].EndmemberID
		}
		return ordered[i].Width > ordered[j].Width
	})
	lines := make([]string, 0, len(ordered))
	for _, b := range ordered {
		lines = append(lines, fmt.Sprintf("%s can vary by %.6g (relative %.6g)", b.EndmemberID, b.Width, b.Relative))
	}
	return lines
}
