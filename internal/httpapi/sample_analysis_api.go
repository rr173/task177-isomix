package httpapi

import (
	"net/http"

	"task177-isomix/internal/uncertainty"
)

// sampleUncertainty GET /api/samples/{id}/uncertainty
func (a *App) sampleUncertainty(w http.ResponseWriter, r *http.Request) {
	sample, err := a.Measure.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, uncertainty.Assess(*sample))
}
