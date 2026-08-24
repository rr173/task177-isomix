package export

import "task177-isomix/internal/model"

// columns 是导出的规范列顺序，同时用于 CSV 表头与 ColumnNames，
// 作为单一事实来源以避免两者漂移。端元行按 endmember_id 升序输出，
// 与报告冻结的 Endmembers 快照顺序保持一致，使同一报告的导出可逐行比对。
var columns = []string{"report_id", "title", "endmember_id", "name", "version", "low", "high"}

// ColumnNames 暴露导出可用的规范列（按 CSV 表头顺序，稳定升序排列）。
func ColumnNames() []string {
	out := make([]string, len(columns))
	copy(out, columns)
	return out
}

// RowCount returns the number of source proportion rows in an export.
func RowCount(report model.Report) int { return len(report.State.EndmemberBounds) }
