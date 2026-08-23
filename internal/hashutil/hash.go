// Package hashutil 提供输入哈希：对求解输入的规范序列化结果计算 SHA-256。
// 相同输入哈希只返回已有结果；发布报告绑定输入哈希以保持可审计性。
package hashutil

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"task177-isomix/internal/model"
)

// canonicalEndmember 把端元降为规范字段（忽略时间、状态与元信息，
// 避免工作流元数据进入哈希；内容变化由 Version 自增体现）。
type canonicalEndmember struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Dimension  int                    `json:"dimension"`
	Components map[string]model.Range `json:"components"`
	Version    int                    `json:"version"`
}

// canonicalSample 把样品降为规范字段。
type canonicalSample struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Dimension   int                    `json:"dimension"`
	Measurements map[string]model.Range `json:"measurements"`
	Covariance  []float64              `json:"covariance,omitempty"`
	Version     int                    `json:"version"`
}

// canonicalConstraint 把约束降为规范字段。
type canonicalConstraint struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Target  string      `json:"target"`
	Range   model.Range `json:"range"`
	Version int         `json:"version"`
}

// canonicalInput 规范输入集合。
type canonicalInput struct {
	Sample      canonicalSample       `json:"sample"`
	Endmembers  []canonicalEndmember  `json:"endmembers"`
	Constraints []canonicalConstraint `json:"constraints"`
}

// SortEndmembers 按 ID 稳定排序端元，保证相同集合的哈希一致。
func SortEndmembers(ems []model.Endmember) {
	sort.Slice(ems, func(i, j int) bool { return ems[i].ID < ems[j].ID })
}

// SortConstraints 按 ID 稳定排序约束。
func SortConstraints(cs []model.Constraint) {
	sort.Slice(cs, func(i, j int) bool { return cs[i].ID < cs[j].ID })
}

func toCanonicalSample(s model.Sample) canonicalSample {
	return canonicalSample{
		ID:          s.ID,
		Name:        s.Name,
		Dimension:   s.Dimension,
		Measurements: s.Measurements,
		Covariance:  s.Covariance,
		Version:     s.Version,
	}
}

func toCanonicalEndmember(e model.Endmember) canonicalEndmember {
	return canonicalEndmember{
		ID:         e.ID,
		Name:       e.Name,
		Dimension:  e.Dimension,
		Components: e.Components,
		Version:    e.Version,
	}
}

func toCanonicalConstraint(c model.Constraint) canonicalConstraint {
	return canonicalConstraint{
		ID:      c.ID,
		Name:    c.Name,
		Type:    string(c.Type),
		Target:  c.Target,
		Range:   c.Range,
		Version: c.Version,
	}
}

// InputHash 计算规范输入摘要的 SHA-256 十六进制串。
func InputHash(s model.Sample, ems []model.Endmember, cs []model.Constraint) (string, error) {
	// 复制后排序，避免修改调用方切片。
	ems2 := make([]model.Endmember, len(ems))
	copy(ems2, ems)
	SortEndmembers(ems2)
	cs2 := make([]model.Constraint, len(cs))
	copy(cs2, cs)
	SortConstraints(cs2)

	in := canonicalInput{Sample: toCanonicalSample(s)}
	for _, e := range ems2 {
		in.Endmembers = append(in.Endmembers, toCanonicalEndmember(e))
	}
	for _, c := range cs2 {
		in.Constraints = append(in.Constraints, toCanonicalConstraint(c))
	}

	raw, err := json.Marshal(in)
	if err != nil {
		return "", fmt.Errorf("hash input marshal: %w", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// HashBytes 对任意字节流计算 SHA-256 十六进制串（供校验与自检复用）。
func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
