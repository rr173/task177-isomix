package export

import (
	"sort"

	"task177-isomix/internal/model"
)

// ColumnNames exposes the deterministic columns available to a client.
func ColumnNames() []string {
	columns := []string{"report_id", "title", "endmember_id", "name", "version", "low", "high"}
	sort.Strings(columns)
	return columns
}

// RowCount returns the number of source proportion rows in an export.
func RowCount(report model.Report) int { return len(report.State.EndmemberBounds) }
