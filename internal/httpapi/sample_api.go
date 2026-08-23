package httpapi

import (
	"net/http"

	"task177-isomix/internal/model"
)

// sampleRequest 样品创建/修订请求体。
type sampleRequest struct {
	Name         string             `json:"name"`
	Measurements map[string]model.Range `json:"measurements"`
	Covariance   []float64          `json:"covariance,omitempty"`
	Meta         map[string]string  `json:"meta,omitempty"`
}

// createSample POST /api/samples
func (a *App) createSample(w http.ResponseWriter, r *http.Request) {
	var req sampleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	sp, err := a.Measure.Create(req.Name, req.Measurements, req.Covariance, req.Meta)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sp)
}

// listSamples GET /api/samples
func (a *App) listSamples(w http.ResponseWriter, r *http.Request) {
	all, err := a.Measure.List()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, all)
}

// getSample GET /api/samples/{id}
func (a *App) getSample(w http.ResponseWriter, r *http.Request) {
	sp, err := a.Measure.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sp)
}

// updateSample PUT /api/samples/{id}
func (a *App) updateSample(w http.ResponseWriter, r *http.Request) {
	var req sampleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	sp, err := a.Measure.UpdateMeasurement(r.PathValue("id"), req.Measurements, req.Covariance)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sp)
}

// checkSample POST /api/samples/{id}/check
func (a *App) checkSample(w http.ResponseWriter, r *http.Request) {
	sp, err := a.Measure.Check(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sp)
}

// sealSample POST /api/samples/{id}/seal
func (a *App) sealSample(w http.ResponseWriter, r *http.Request) {
	sp, err := a.Measure.Seal(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sp)
}
