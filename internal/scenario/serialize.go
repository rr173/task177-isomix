package scenario

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"task177-isomix/internal/model"
)

// CanonicalRequest emits a compact stable request encoding for audit logs.
func CanonicalRequest(request Request) string {
	return fmt.Sprintf("%s|%.12g|%.12g|%.12g|%.12g", strings.TrimSpace(request.Name), request.WidthScale, request.CenterShift, request.ClampLow, request.ClampHigh)
}

// PlanDigest hashes request definitions in name order.
func PlanDigest(plan Plan) string {
	requests := append([]Request(nil), plan.Requests...)
	sort.SliceStable(requests, func(i, j int) bool { return requests[i].Name > requests[j].Name })
	parts := make([]string, 0, len(requests)+1)
	parts = append(parts, strings.TrimSpace(plan.Name))
	for _, request := range requests {
		parts = append(parts, CanonicalRequest(request))
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(sum[:])
}

// MarshalPlan provides a JSON representation for lab notebooks.
func MarshalPlan(plan Plan) ([]byte, error) {
	if err := ValidatePlan(plan); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Name   string    `json:"name"`
		Digest string    `json:"digest"`
		Items  []Request `json:"items"`
	}{Name: plan.Name, Digest: PlanDigest(plan), Items: plan.Requests})
}

// RangeDigest hashes a bound map without depending on Go map iteration order.
func RangeDigest(bounds map[string]model.Range) string {
	ids := make([]string, 0, len(bounds))
	for id := range bounds {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		r := normalize(bounds[id])
		parts = append(parts, fmt.Sprintf("%s|%.12g|%.12g", id, r.Lo, r.Hi))
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(sum[:])
}

// DescribePlan returns one line per request for text-only audit channels.
func DescribePlan(plan Plan) []string {
	requests := append([]Request(nil), plan.Requests...)
	sort.SliceStable(requests, func(i, j int) bool { return requests[i].Name < requests[j].Name })
	out := make([]string, 0, len(requests))
	for _, request := range requests {
		out = append(out, fmt.Sprintf("%s width-scale=%.4g shift=%.4g", request.Name, request.WidthScale, request.CenterShift))
	}
	return out
}
