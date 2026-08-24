package export

import (
	"strings"
	"testing"

	"task177-isomix/internal/model"
)

func TestReportExportsContainStableBounds(t *testing.T) {
	report := model.Report{ID: "rp1", Title: "demo", Endmembers: []model.EndmemberSnapshot{{ID: "a", Name: "A", Version: 1}}, State: model.SolutionState{Feasible: true, EndmemberBounds: map[string]model.Range{"a": {Lo: .2, Hi: .8}}}}
	csv, err := CSV(report)
	if err != nil || !strings.Contains(csv, "endmember_id") || !strings.Contains(csv, "a") {
		t.Fatalf("bad csv: %q, %v", csv, err)
	}
	json, err := JSON(report)
	if err != nil || !strings.Contains(string(json), "report_id") {
		t.Fatalf("bad json: %s, %v", json, err)
	}
}

// TestReportExportStableOrder 锁定导出的两个稳定性要求：
//  1. CSV 表头完整且 report_id 列名正确（旧实现误写为重复的 "title"）；
//  2. 端元行按 endmember_id 升序输出，跨调用可逐行比对。
func TestReportExportStableOrder(t *testing.T) {
	report := model.Report{
		ID:    "rp9",
		Title: "stable-order demo",
		Endmembers: []model.EndmemberSnapshot{
			{ID: "em_b", Name: "B", Version: 1},
			{ID: "em_a", Name: "A", Version: 2},
			{ID: "em_c", Name: "C", Version: 1},
		},
		State: model.SolutionState{Feasible: true, EndmemberBounds: map[string]model.Range{
			"em_a": {Lo: .1, Hi: .2},
			"em_b": {Lo: .3, Hi: .4},
			"em_c": {Lo: .5, Hi: .6},
		}},
	}

	// 表头列：与 ColumnNames 一致，report_id 列必须存在且不再被误写为 title。
	wantHeader := strings.Join(ColumnNames(), ",")
	csvOut, err := CSV(report)
	if err != nil {
		t.Fatalf("csv: %v", err)
	}
	lines := strings.Split(strings.TrimRight(csvOut, "\n"), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected header + 3 rows, got %d lines: %q", len(lines), csvOut)
	}
	if lines[0] != wantHeader {
		t.Fatalf("header mismatch:\n got: %s\nwant: %s", lines[0], wantHeader)
	}
	if !strings.HasPrefix(lines[0], "report_id") {
		t.Fatalf("header must start with report_id, got %q", lines[0])
	}

	// 端元行按 ID 升序排列。
	want := []string{"em_a", "em_b", "em_c"}
	for i, id := range want {
		row := lines[1+i]
		fields := strings.Split(row, ",")
		if len(fields) != 7 || fields[2] != id {
			t.Fatalf("row %d: want endmember_id=%s, got %q", i, id, row)
		}
		if fields[0] != report.ID {
			t.Fatalf("row %d: report_id=%s, want %s", i, fields[0], report.ID)
		}
		if fields[1] != report.Title {
			t.Fatalf("row %d: title=%s, want %s", i, fields[1], report.Title)
		}
	}

	// JSON 导出共享 rows()，端元顺序应与 CSV 一致。
	jout, err := JSON(report)
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	jstr := string(jout)
	if ia, ib, ic := strings.Index(jstr, "\"em_a\""), strings.Index(jstr, "\"em_b\""), strings.Index(jstr, "\"em_c\""); !(ia < ib && ib < ic) {
		t.Fatalf("json bounds not in ascending id order: a=%d b=%d c=%d", ia, ib, ic)
	}
}
