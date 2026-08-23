// Package httpapi 提供 REST JSON API（统一前缀 /api）。
// 错误映射：领域错误按 Code 映射 HTTP 状态码，其余为 500。
package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"task177-isomix/internal/constraint"
	"task177-isomix/internal/endmember"
	"task177-isomix/internal/measure"
	"task177-isomix/internal/model"
	"task177-isomix/internal/report"
	"task177-isomix/internal/sensitivity"
	"task177-isomix/internal/solve"
)

// App 聚合各领域服务，供 handler 使用。
type App struct {
	Endmembers  *endmember.Service
	Measure     *measure.Service
	Constraints *constraint.Service
	Solve       *solve.Service
	Reports     *report.Service
	Sensitivity *sensitivity.Service
}

// NewApp 构造 App。
func NewApp(em *endmember.Service, ms *measure.Service, cs *constraint.Service, sv *solve.Service, rp *report.Service, sens *sensitivity.Service) *App {
	return &App{Endmembers: em, Measure: ms, Constraints: cs, Solve: sv, Reports: rp, Sensitivity: sens}
}

// Routes 注册全部路由，返回根 mux。
func (a *App) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	// 端元
	mux.HandleFunc("POST /api/endmembers", a.createEndmember)
	mux.HandleFunc("GET /api/endmembers", a.listEndmembers)
	mux.HandleFunc("GET /api/endmembers/{id}", a.getEndmember)
	mux.HandleFunc("PUT /api/endmembers/{id}", a.reviseEndmember)
	mux.HandleFunc("POST /api/endmembers/{id}/validate", a.validateEndmember)
	mux.HandleFunc("POST /api/endmembers/{id}/available", a.makeEndmemberAvailable)
	mux.HandleFunc("POST /api/endmembers/{id}/exclude", a.excludeEndmember)

	// 样品
	mux.HandleFunc("POST /api/samples", a.createSample)
	mux.HandleFunc("GET /api/samples", a.listSamples)
	mux.HandleFunc("GET /api/samples/{id}", a.getSample)
	mux.HandleFunc("PUT /api/samples/{id}", a.updateSample)
	mux.HandleFunc("POST /api/samples/{id}/check", a.checkSample)
	mux.HandleFunc("POST /api/samples/{id}/seal", a.sealSample)

	// 约束
	mux.HandleFunc("POST /api/constraints", a.createConstraint)
	mux.HandleFunc("GET /api/constraints", a.listConstraints)
	mux.HandleFunc("GET /api/constraints/{id}", a.getConstraint)
	mux.HandleFunc("POST /api/constraints/{id}/enable", a.enableConstraint)
	mux.HandleFunc("POST /api/constraints/{id}/relax", a.relaxConstraint)
	mux.HandleFunc("POST /api/constraints/{id}/revoke", a.revokeConstraint)

	// 求解
	mux.HandleFunc("POST /api/solutions", a.submitSolution)
	mux.HandleFunc("GET /api/solutions", a.listSolutions)
	mux.HandleFunc("GET /api/solutions/{id}", a.getSolution)
	mux.HandleFunc("GET /api/solutions/{id}/feasible-region", a.solutionFeasibleRegion)
	mux.HandleFunc("GET /api/solutions/{id}/conflict-core", a.solutionConflictCore)
	mux.HandleFunc("GET /api/solutions/{id}/sensitivity", a.solutionSensitivity)
	mux.HandleFunc("GET /api/solutions/{id}/diagnostics", a.solutionDiagnostics)

	// 报告
	mux.HandleFunc("POST /api/reports", a.createReport)
	mux.HandleFunc("GET /api/reports", a.listReports)
	mux.HandleFunc("GET /api/reports/{id}", a.getReport)
	mux.HandleFunc("POST /api/reports/{id}/publish", a.publishReport)
	mux.HandleFunc("GET /api/reports/{id}/supersede-check", a.supersedeCheck)
	mux.HandleFunc("GET /api/reports/{id}/diff", a.reportDiff)
	mux.HandleFunc("GET /api/reports/{id}/provenance", a.reportProvenance)
	mux.HandleFunc("GET /api/reports/{id}/export", a.reportExport)

	// 自检与统计
	mux.HandleFunc("GET /api/health", a.health)
	mux.HandleFunc("GET /api/health/selfcheck", a.selfCheck)
	mux.HandleFunc("GET /api/stats", a.stats)
	mux.HandleFunc("GET /api/samples/{id}/uncertainty", a.sampleUncertainty)

	return mux
}

// writeJSON 输出 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

// writeErr 输出统一错误响应：{code, message}。
func writeErr(w http.ResponseWriter, err error) {
	var de *model.DomainError
	if errors.As(err, &de) {
		writeJSON(w, statusForCode(de.Code), de)
		return
	}
	writeJSON(w, http.StatusInternalServerError, model.NewError("INTERNAL", "%v", err))
}

// statusForCode 领域错误码 -> HTTP 状态码。
func statusForCode(code string) int {
	switch code {
	case "NOT_FOUND":
		return http.StatusNotFound
	case "BAD_REQUEST", "INVALID_RANGE", "DIMENSION_MISMATCH", "NEGATIVE_MASS",
		"DUPLICATE_ENDMEMBER", "SEALED_SAMPLE", "PROPORTIONS_NOT_SUM_ONE",
		"NO_AVAILABLE_ENDMEMBER", "INACTIVE_ENDMEMBER", "ALREADY_SOLVED":
		return http.StatusBadRequest
	case "SINGULAR_COVARIANCE", "NUMERICALLY_UNSTABLE":
		return http.StatusUnprocessableEntity
	case "INVALID_TRANSITION", "REPORT_FROZEN":
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// decodeJSON 解析请求体到 v。
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeErr(w, model.NewError("BAD_REQUEST", "invalid json body: %v", err))
		return false
	}
	return true
}
