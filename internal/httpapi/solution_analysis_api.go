package httpapi

import (
	"net/http"

	"task177-isomix/internal/diagnostics"
	"task177-isomix/internal/sensitivity"
)

// solutionSensitivity GET /api/solutions/{id}/sensitivity
func (a *App) solutionSensitivity(w http.ResponseWriter, r *http.Request) {
	report, err := a.Sensitivity.Analyze(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"report":      report,
		"explanation": sensitivity.Explain(report),
	})
}

// solutionDiagnostics GET /api/solutions/{id}/diagnostics
func (a *App) solutionDiagnostics(w http.ResponseWriter, r *http.Request) {
	solution, err := a.Solve.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, diagnostics.Analyze(solution))
}
