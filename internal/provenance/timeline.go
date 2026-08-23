package provenance

import (
	"time"

	"task177-isomix/internal/model"
)

// BuildSnapshot captures a report's immutable input chain at a caller-provided
// time, which makes replay and audit output deterministic in tests.
func BuildSnapshot(report model.Report, captured time.Time) Snapshot {
	events := Chain(report)
	return Snapshot{ReportID: report.ID, InputHash: report.InputHash, Status: string(report.Status), Captured: captured.UTC(), Events: events}
}
