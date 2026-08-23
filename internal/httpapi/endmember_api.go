package httpapi

import (
	"net/http"

	"task177-isomix/internal/model"
)

// endmemberRequest 端元创建/修订请求体。
type endmemberRequest struct {
	Name       string             `json:"name"`
	Components map[string]model.Range `json:"components"`
	Meta       map[string]string  `json:"meta,omitempty"`
}

// createEndmember POST /api/endmembers
func (a *App) createEndmember(w http.ResponseWriter, r *http.Request) {
	var req endmemberRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	e, err := a.Endmembers.Create(req.Name, req.Components, req.Meta)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

// listEndmembers GET /api/endmembers?status=available
func (a *App) listEndmembers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("status") == "available" {
		ems, err := a.Endmembers.ListAvailable()
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, ems)
		return
	}
	ems, err := a.Endmembers.List()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ems)
}

// getEndmember GET /api/endmembers/{id}
func (a *App) getEndmember(w http.ResponseWriter, r *http.Request) {
	e, err := a.Endmembers.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// reviseEndmember PUT /api/endmembers/{id}
func (a *App) reviseEndmember(w http.ResponseWriter, r *http.Request) {
	var req endmemberRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	e, err := a.Endmembers.Revise(r.PathValue("id"), req.Components)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// validateEndmember POST /api/endmembers/{id}/validate
func (a *App) validateEndmember(w http.ResponseWriter, r *http.Request) {
	e, err := a.Endmembers.Validate(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// makeEndmemberAvailable POST /api/endmembers/{id}/available
func (a *App) makeEndmemberAvailable(w http.ResponseWriter, r *http.Request) {
	e, err := a.Endmembers.MakeAvailable(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// excludeEndmember POST /api/endmembers/{id}/exclude
func (a *App) excludeEndmember(w http.ResponseWriter, r *http.Request) {
	e, err := a.Endmembers.Exclude(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}
