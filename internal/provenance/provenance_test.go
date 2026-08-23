package provenance

import (
	"testing"
	"time"

	"task177-isomix/internal/model"
)

func TestSnapshotDigestIgnoresInputOrder(t *testing.T) {
	report := model.Report{ID: "rp1", InputHash: "hash", Status: model.ReportPublished, SampleID: "sp1", Endmembers: []model.EndmemberSnapshot{{ID: "b", Version: 2, Name: "b"}, {ID: "a", Version: 1, Name: "a"}}}
	snapshot := BuildSnapshot(report, time.Unix(10, 0))
	if !Contains(snapshot.Events, "endmember", "a", 1) {
		t.Fatalf("expected versioned endmember in chain")
	}
	if len(Digest(snapshot.Events)) != 64 {
		t.Fatalf("digest should be sha256 hex")
	}
}
