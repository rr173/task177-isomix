package export

import (
	"encoding/json"

	"task177-isomix/internal/model"
)

// JSON returns a stable, self-contained report export.
func JSON(report model.Report) ([]byte, error) {
	payload := struct {
		Summary map[string]any      `json:"summary"`
		Bounds  []BoundRow          `json:"bounds"`
		State   model.SolutionState `json:"state"`
	}{Summary: statusRow(report), Bounds: rows(report), State: report.State}
	return json.MarshalIndent(payload, "", "  ")
}
