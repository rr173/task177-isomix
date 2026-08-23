package export

import (
	"strings"
	"testing"
	"task177-isomix/internal/model"
)

func TestBug29_CSVSortsEndmembers(t *testing.T) {
	report := model.Report{ID: "r", Endmembers: []model.EndmemberSnapshot{{ID: "b"}, {ID: "a"}}, State: model.SolutionState{EndmemberBounds: map[string]model.Range{"b": {Hi: 1}, "a": {Hi: 1}}}}; csv, _ := CSV(report); if strings.Index(csv, "a") > strings.Index(csv, "b") { t.Fatalf("csv=%s", csv) }
}
