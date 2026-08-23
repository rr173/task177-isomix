package provenance

import (
	"fmt"
	"time"

	"task177-isomix/internal/model"
)

// Event is a stable explanation of one versioned report input.
type Event struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Version int    `json:"version"`
	Label   string `json:"label"`
}

func reportEvents(report model.Report) []Event {
	events := make([]Event, 0, len(report.Endmembers)+len(report.Constraints)+1)
	events = append(events, Event{Kind: "sample", ID: report.SampleID, Version: 0, Label: "measurement used by solution"})
	for _, e := range report.Endmembers {
		events = append(events, Event{Kind: "endmember", ID: e.ID, Version: e.Version, Label: e.Name})
	}
	for _, c := range report.Constraints {
		events = append(events, Event{Kind: "constraint", ID: c.ID, Version: c.Version, Label: c.Name})
	}
	return events
}

func eventKey(event Event) string {
	return fmt.Sprintf("%s:%s:v%d", event.ID, event.Kind, event.Version)
}

// Snapshot records when the report explanation was generated.
type Snapshot struct {
	ReportID  string    `json:"report_id"`
	InputHash string    `json:"input_hash"`
	Status    string    `json:"status"`
	Captured  time.Time `json:"captured"`
	Events    []Event   `json:"events"`
}
