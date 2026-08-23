package httpapi

import (
	"net/http"

	"task177-isomix/internal/model"
)

// reportRequest 报告创建请求体。
type reportRequest struct {
	SolutionID string `json:"solution_id"`
	Title      string `json:"title"`
}

// createReport POST /api/reports
func (a *App) createReport(w http.ResponseWriter, r *http.Request) {
	var req reportRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	rp, err := a.Reports.Create(req.SolutionID, req.Title)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rp)
}

// listReports GET /api/reports
func (a *App) listReports(w http.ResponseWriter, r *http.Request) {
	all, err := a.Reports.List()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, all)
}

// getReport GET /api/reports/{id}
func (a *App) getReport(w http.ResponseWriter, r *http.Request) {
	rp, err := a.Reports.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rp)
}

// publishReport POST /api/reports/{id}/publish
func (a *App) publishReport(w http.ResponseWriter, r *http.Request) {
	rp, err := a.Reports.Publish(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rp)
}

// supersedeCheck GET /api/reports/{id}/supersede-check
func (a *App) supersedeCheck(w http.ResponseWriter, r *http.Request) {
	rp, err := a.Reports.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	stale, changes := a.Reports.CheckSuperseded(rp)
	writeJSON(w, http.StatusOK, map[string]any{
		"report_id": rp.ID,
		"status":    rp.Status,
		"stale":     stale,
		"changes":   changes,
	})
}

// reportDiff GET /api/reports/{id}/diff?against={otherID}
func (a *App) reportDiff(w http.ResponseWriter, r *http.Request) {
	other := r.URL.Query().Get("against")
	if other == "" {
		writeErr(w, model.NewError("BAD_REQUEST", "query param 'against' is required"))
		return
	}
	aR, err := a.Reports.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	bR, err := a.Reports.Get(other)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a.Reports.Diff(aR, bR))
}
