// Package model 定义同位素混合来源约束求解服务的领域实体与领域错误。
package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// 实体状态常量。

// EndmemberStatus 端元生命周期状态。
type EndmemberStatus string

const (
	EndmemberDraft     EndmemberStatus = "draft"     // 草拟：组成区间尚在录入，不可参与求解
	EndmemberValidated EndmemberStatus = "validated" // 已校验：维度与区间均通过校验
	EndmemberAvailable EndmemberStatus = "available" // 可用：可参与混合求解
	EndmemberExcluded  EndmemberStatus = "excluded"  // 已排除：被研究员排除出候选集
)

// SampleStatus 样品生命周期状态。
type SampleStatus string

const (
	SamplePending    SampleStatus = "pending"    // 待测：测量未完成
	SampleSolvable   SampleStatus = "solvable"   // 可求解：测量完备且不确定度可接受
	SampleOverspread SampleStatus = "overspread" // 误差过大：不确定度超阈值，直接求解无意义
	SampleSolved     SampleStatus = "solved"     // 已求解：存在已接受的求解
	SampleSealed     SampleStatus = "sealed"     // 已封存：禁止再次求解或修订
)

// ConstraintStatus 约束生命周期状态。
type ConstraintStatus string

const (
	ConstraintEnabled  ConstraintStatus = "enabled"  // 启用：参与编译
	ConstraintRelaxed  ConstraintStatus = "relaxed"  // 松弛：暂不参与编译但保留
	ConstraintConflict ConstraintStatus = "conflict" // 冲突：曾导致不可行并被标记
	ConstraintRevoked  ConstraintStatus = "revoked"  // 已撤销：永久失效
)

// SolutionStatus 求解生命周期状态。
type SolutionStatus string

const (
	SolutionQueued    SolutionStatus = "queued"     // 排队中：等待执行
	SolutionRunning   SolutionStatus = "running"    // 求解中
	SolutionFeasible  SolutionStatus = "feasible"   // 可行：存在满足全部约束的比例
	SolutionInfeasible SolutionStatus = "infeasible" // 不可行：存在冲突约束
	SolutionUnstable  SolutionStatus = "numerically_unstable" // 数值不稳定：无法给出可信比例
)

// ReportStatus 报告生命周期状态。
type ReportStatus string

const (
	ReportDraft      ReportStatus = "draft"      // 草案：可编辑
	ReportPublished  ReportStatus = "published"  // 已发布：冻结快照，可被引用
	ReportSuperseded ReportStatus = "superseded" // 已替代：因端元/约束修订而废弃
)

// Range 表示一个闭区间 [Lo, Hi]。
type Range struct {
	Lo float64 `json:"lo"`
	Hi float64 `json:"hi"`
}

// Valid 判断区间是否合法（非空且两端有序）。
func (r Range) Valid() bool {
	return r.Lo <= r.Hi
}

// Width 返回区间宽度。
func (r Range) Width() float64 {
	return r.Hi - r.Lo
}

// Contains 判断值是否落在区间内（含边界，容差 1e-9）。
func (r Range) Contains(v float64) bool {
	const tol = 1e-9
	return v >= r.Lo-tol && v <= r.Hi+tol
}

// Endmember 同位素端元：一个候选来源，携带各同位素比率的组成区间。
type Endmember struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Status     EndmemberStatus        `json:"status"`
	Components map[string]Range       `json:"components"` // 同位素名 -> 组成区间
	Dimension  int                    `json:"dimension"`  // 同位素维度数
	Meta       map[string]string      `json:"meta,omitempty"`
	Version    int                    `json:"version"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// Snapshot 返回端元的版本化快照（供报告冻结）。
func (e Endmember) Snapshot() EndmemberSnapshot {
	return EndmemberSnapshot{
		ID:         e.ID,
		Name:       e.Name,
		Status:     e.Status,
		Components: e.Components,
		Dimension:  e.Dimension,
		Version:    e.Version,
	}
}

// EndmemberSnapshot 报告冻结时使用的端元不可变快照。
type EndmemberSnapshot struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Status     EndmemberStatus  `json:"status"`
	Components map[string]Range `json:"components"`
	Dimension  int              `json:"dimension"`
	Version    int              `json:"version"`
}

// Sample 样品：一组同位素测量值与不确定度，等待与候选端元混合解释。
type Sample struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Status      SampleStatus      `json:"status"`
	Measurements map[string]Range `json:"measurements"` // 同位素名 -> 测量区间
	Covariance  []float64         `json:"covariance,omitempty"` // 协方差矩阵上三角摘要（可选）
	Dimension   int               `json:"dimension"`
	Meta        map[string]string `json:"meta,omitempty"`
	Version     int               `json:"version"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CovarianceMatrix 把协方差摘要还原为方阵。
func (s Sample) CovarianceMatrix() [][]float64 {
	d := s.Dimension
	if len(s.Covariance) == 0 {
		return nil
	}
	m := make([][]float64, d)
	for i := range m {
		m[i] = make([]float64, d)
	}
	idx := 0
	for i := 0; i < d; i++ {
		for j := i; j < d; j++ {
			if idx < len(s.Covariance) {
				m[i][j] = s.Covariance[idx]
				m[j][i] = s.Covariance[idx]
				idx++
			}
		}
	}
	return m
}

// ConstraintType 约束类型。
type ConstraintType string

const (
	ConstraintExclude ConstraintType = "exclude" // 排除：指定端元比例强制为 0
	ConstraintRatio   ConstraintType = "ratio"   // 比例：两个端元比例之比落在区间内
	ConstraintBound   ConstraintType = "bound"   // 上限：指定端元比例不超过阈值
)

// Constraint 求解约束：编译为线性不等式参与可行域求解。
type Constraint struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        ConstraintType `json:"type"`
	Target      string         `json:"target"`  // exclude/bound: 端元 ID；ratio: "A:B"
	Range       Range          `json:"range"`   // ratio: 比值区间；bound: [0, 上限]
	Status      ConstraintStatus `json:"status"`
	Note        string         `json:"note,omitempty"`
	Version     int            `json:"version"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ConstraintSnapshot 报告冻结时使用的约束不可变快照。
type ConstraintSnapshot struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Type    ConstraintType `json:"type"`
	Target  string         `json:"target"`
	Range   Range          `json:"range"`
	Status  ConstraintStatus `json:"status"`
	Version int            `json:"version"`
}

// Snapshot 返回约束的版本化快照。
func (c Constraint) Snapshot() ConstraintSnapshot {
	return ConstraintSnapshot{
		ID:      c.ID,
		Name:    c.Name,
		Type:    c.Type,
		Target:  c.Target,
		Range:   c.Range,
		Status:  c.Status,
		Version: c.Version,
	}
}

// SolutionState 求解输出：可行域区间、活跃约束或最小不可满足子集。
type SolutionState struct {
	Feasible         bool              `json:"feasible"`
	EndmemberBounds  map[string]Range  `json:"endmember_bounds,omitempty"`  // 各端元比例可行区间
	Residual         float64           `json:"residual,omitempty"`          // 最优残差（可行时为 0）
	ActiveConstraints []string         `json:"active_constraints,omitempty"` // 边界活跃约束 ID
	ConflictCore     []string          `json:"conflict_core,omitempty"`     // 最小不可满足约束 ID 集
	UnstableEvidence string            `json:"unstable_evidence,omitempty"` // 数值不稳定证据
	MatrixSummary    MatrixSummary     `json:"matrix_summary"`
}

// MatrixSummary 求解矩阵摘要：约束数、变量数、非零元。
type MatrixSummary struct {
	Endmembers int `json:"endmembers"`
	Isotopes   int `json:"isotopes"`
	Rows       int `json:"rows"`
	NonZeros   int `json:"non_zeros"`
}

// Solution 一次混合求解记录。
type Solution struct {
	ID          string        `json:"id"`
	SampleID    string        `json:"sample_id"`
	Status      SolutionStatus `json:"status"`
	InputHash   string        `json:"input_hash"`
	State       SolutionState `json:"state"`
	EndmemberIDs []string     `json:"endmember_ids"`
	ConstraintIDs []string    `json:"constraint_ids"`
	Error       string        `json:"error,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	FinishedAt  *time.Time    `json:"finished_at,omitempty"`
}

// Report 可引用的求解报告：绑定全部端元与约束版本快照。
type Report struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Status      ReportStatus       `json:"status"`
	SolutionID  string             `json:"solution_id"`
	SampleID    string             `json:"sample_id"`
	InputHash   string             `json:"input_hash"`
	Endmembers  []EndmemberSnapshot  `json:"endmembers"`
	Constraints []ConstraintSnapshot `json:"constraints"`
	State       SolutionState      `json:"state"`
	SupersededBy string            `json:"superseded_by,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	PublishedAt *time.Time         `json:"published_at,omitempty"`
}

// 领域错误：用哨兵错误 + 包装，HTTP 层按类型映射状态码。

// DomainError 领域错误基类，携带面向用户的稳定代码。
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewError 构造领域错误。
func NewError(code, format string, args ...any) *DomainError {
	return &DomainError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// 标准领域错误哨兵（Code 稳定，供错误映射与测试断言）。
var (
	ErrNotFound            = &DomainError{Code: "NOT_FOUND", Message: "resource not found"}
	ErrDuplicateEndmember  = &DomainError{Code: "DUPLICATE_ENDMEMBER", Message: "endmember id already exists"}
	ErrDimensionMismatch   = &DomainError{Code: "DIMENSION_MISMATCH", Message: "isotope dimensions inconsistent"}
	ErrInvalidRange        = &DomainError{Code: "INVALID_RANGE", Message: "range is empty or reversed"}
	ErrNegativeMass        = &DomainError{Code: "NEGATIVE_MASS", Message: "mass or concentration must be non-negative"}
	ErrSingularCovariance  = &DomainError{Code: "SINGULAR_COVARIANCE", Message: "covariance matrix is singular"}
	ErrSealedSample        = &DomainError{Code: "SEALED_SAMPLE", Message: "sample is sealed and cannot be re-solved"}
	ErrProportionsNotSumOne = &DomainError{Code: "PROPORTIONS_NOT_SUM_ONE", Message: "proportions must sum to one"}
	ErrNoAvailableEndmember = &DomainError{Code: "NO_AVAILABLE_ENDMEMBER", Message: "no available endmember for solving"}
	ErrInactiveEndmember   = &DomainError{Code: "INACTIVE_ENDMEMBER", Message: "endmember is not available"}
	ErrAlreadySolved       = &DomainError{Code: "ALREADY_SOLVED", Message: "sample already solved; duplicate results are returned by input hash"}
	ErrReportFrozen        = &DomainError{Code: "REPORT_FROZEN", Message: "published report cannot be modified"}
	ErrInvalidTransition   = &DomainError{Code: "INVALID_TRANSITION", Message: "illegal status transition"}
	ErrNumericallyUnstable = &DomainError{Code: "NUMERICALLY_UNSTABLE", Message: "solver detected numeric instability"}
)

// MarshalJSON 让 DomainError 输出稳定 JSON 形状（供 API 响应复用）。
func (e *DomainError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{Code: e.Code, Message: e.Message})
}
