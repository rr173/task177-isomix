package httpapi

import (
	"net/http"

	"task177-isomix/internal/model"
)

// solutionSubmitRequest 求解提交请求体。
type solutionSubmitRequest struct {
	SampleID string `json:"sample_id"`
}

// submitSolution POST /api/solutions
func (a *App) submitSolution(w http.ResponseWriter, r *http.Request) {
	var req solutionSubmitRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.SampleID == "" {
		writeErr(w, model.NewError("BAD_REQUEST", "sample_id is required"))
		return
	}
	sol, err := a.Solve.Submit(req.SampleID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sol)
}

// listSolutions GET /api/solutions
func (a *App) listSolutions(w http.ResponseWriter, r *http.Request) {
	all, err := a.Solve.List()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, all)
}

// getSolution GET /api/solutions/{id}
func (a *App) getSolution(w http.ResponseWriter, r *http.Request) {
	sol, err := a.Solve.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sol)
}

// solutionFeasibleRegion GET /api/solutions/{id}/feasible-region
func (a *App) solutionFeasibleRegion(w http.ResponseWriter, r *http.Request) {
	sol, err := a.Solve.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if !sol.State.Feasible {
		writeJSON(w, http.StatusOK, map[string]any{
			"solution_id": sol.ID,
			"feasible":    false,
			"status":      sol.Status,
			"reason":      "no feasible region; inspect conflict core",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"solution_id":      sol.ID,
		"feasible":         true,
		"endmember_bounds": sol.State.EndmemberBounds,
		"active_constraints": sol.State.ActiveConstraints,
		"matrix":           sol.State.MatrixSummary,
	})
}

// solutionConflictCore GET /api/solutions/{id}/conflict-core
func (a *App) solutionConflictCore(w http.ResponseWriter, r *http.Request) {
	sol, err := a.Solve.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"solution_id":  sol.ID,
		"feasible":     sol.State.Feasible,
		"conflict_core": sol.State.ConflictCore,
	})
}
