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
