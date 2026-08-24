package export

import (
	"fmt"
	"sort"

	"task177-isomix/internal/model"
)

// BoundRow is the tabular form used by JSON and CSV exporters.
type BoundRow struct {
	EndmemberID string  `json:"endmember_id"`
	Name        string  `json:"name"`
	Version     int     `json:"version"`
	Low         float64 `json:"low"`
	High        float64 `json:"high"`
}

// rows 把报告端元可行域区间转为行序列，供 JSON 与 CSV 导出复用。
// 端元 ID 按稳定升序输出，与 report.Service 冻结 Endmembers 快照时的排序一致，
// 使同一报告的多次导出可逐行比对（否则 map 迭代顺序会破坏稳定性）。
func rows(report model.Report) []BoundRow {
	names := make(map[string]string, len(report.Endmembers))
	versions := make(map[string]int, len(report.Endmembers))
	for _, endmember := range report.Endmembers {
		names[endmember.ID] = endmember.Name
		versions[endmember.ID] = endmember.Version
	}
	ids := make([]string, 0, len(report.State.EndmemberBounds))
	for id := range report.State.EndmemberBounds {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]BoundRow, 0, len(ids))
	for _, id := range ids {
		bound := report.State.EndmemberBounds[id]
		out = append(out, BoundRow{EndmemberID: id, Name: names[id], Version: versions[id], Low: bound.Lo, High: bound.Hi})
	}
	return out
}

func statusRow(report model.Report) map[string]any {
	return map[string]any{
		"report_id":   report.ID,
		"title":       report.Title,
		"status":      report.Status,
		"solution_id": report.SolutionID,
		"sample_id":   report.SampleID,
		"input_hash":  report.InputHash,
		"feasible":    report.State.Feasible,
		"rows":        len(report.State.EndmemberBounds),
	}
}

func title(report model.Report) string {
	if report.Title == "" {
		return fmt.Sprintf("report-%s", report.ID)
	}
	return report.Title
}
