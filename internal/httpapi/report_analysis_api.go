package httpapi

import (
	"net/http"

	"task177-isomix/internal/export"
	"task177-isomix/internal/model"
	"task177-isomix/internal/provenance"
)

// reportProvenance GET /api/reports/{id}/provenance
func (a *App) reportProvenance(w http.ResponseWriter, r *http.Request) {
	report, err := a.Reports.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	snapshot := provenance.BuildSnapshot(*report, report.CreatedAt)
	writeJSON(w, http.StatusOK, map[string]any{
		"snapshot": snapshot,
		"digest":   provenance.Digest(snapshot.Events),
	})
}

// reportExport GET /api/reports/{id}/export?format=json|csv
func (a *App) reportExport(w http.ResponseWriter, r *http.Request) {
	report, err := a.Reports.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	format := r.URL.Query().Get("format")
	switch format {
	case "csv":
		body, err := export.CSV(*report)
		if err != nil {
			writeErr(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	case "", "json":
		body, err := export.JSON(*report)
		if err != nil {
			writeErr(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	default:
		writeErr(w, model.NewError("BAD_REQUEST", "unsupported export format %q", format))
	}
}
