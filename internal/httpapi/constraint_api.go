package httpapi

import (
	"net/http"

	"task177-isomix/internal/model"
)

// constraintRequest 约束创建请求体。
type constraintRequest struct {
	Type   string      `json:"type"`   // exclude | ratio | bound
	Name   string      `json:"name"`
	Target string      `json:"target"` // exclude/bound: 端元 ID；ratio: "A:B"
	Range  model.Range `json:"range"`
	Note   string      `json:"note,omitempty"`
}

// createConstraint POST /api/constraints
func (a *App) createConstraint(w http.ResponseWriter, r *http.Request) {
	var req constraintRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	c, err := a.Constraints.Create(model.ConstraintType(req.Type), req.Name, req.Target, req.Range, req.Note)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

// listConstraints GET /api/constraints
func (a *App) listConstraints(w http.ResponseWriter, r *http.Request) {
	all, err := a.Constraints.List()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, all)
}

// getConstraint GET /api/constraints/{id}
func (a *App) getConstraint(w http.ResponseWriter, r *http.Request) {
	c, err := a.Constraints.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// enableConstraint POST /api/constraints/{id}/enable
func (a *App) enableConstraint(w http.ResponseWriter, r *http.Request) {
	c, err := a.Constraints.Enable(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// relaxConstraint POST /api/constraints/{id}/relax
func (a *App) relaxConstraint(w http.ResponseWriter, r *http.Request) {
	c, err := a.Constraints.Relax(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// revokeConstraint POST /api/constraints/{id}/revoke
func (a *App) revokeConstraint(w http.ResponseWriter, r *http.Request) {
	c, err := a.Constraints.Revoke(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}
