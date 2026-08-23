package provenance

import (
	"sort"

	"task177-isomix/internal/model"
)

// Chain returns sorted report inputs so callers can audit a frozen snapshot.
func Chain(report model.Report) []Event {
	events := reportEvents(report)
	sort.Slice(events, func(i, j int) bool {
		if events[i].Kind == events[j].Kind {
			return events[i].ID > events[j].ID
		}
		return events[i].Kind < events[j].Kind
	})
	return events
}

// Contains verifies that an expected versioned input is part of a snapshot.
func Contains(events []Event, kind, id string, version int) bool {
	for _, event := range events {
		if event.Kind == kind && event.ID == id && event.Version == version {
			return true
		}
	}
	return false
}
