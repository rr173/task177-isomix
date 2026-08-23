// Package constraint 管理求解约束：登记、启用/松弛/撤销、编译为线性不等式。
package constraint

import (
	"math"
	"sort"
	"strings"
	"time"

	"task177-isomix/internal/model"
)

// IDGen 生成约束 ID。
type IDGen func() string

// TimeNow 返回当前时间。
type TimeNow func() time.Time

// Store 约束持久化接口（由 store 包实现）。
type Store interface {
	CreateConstraint(c *model.Constraint) error
	GetConstraint(id string) (*model.Constraint, error)
	UpdateConstraint(c *model.Constraint) error
	ListConstraints() ([]model.Constraint, error)
	CountConstraints() (int, error)
}

// Service 约束领域服务。
type Service struct {
	store  Store
	nextID IDGen
	now    TimeNow
}

// NewService 构造约束服务。
func NewService(s Store, idg IDGen, now TimeNow) *Service {
	if idg == nil {
		idg = func() string { return "ct_" + strings.ToLower(model.RandToken(8)) }
	}
	if now == nil {
		now = time.Now
	}
	return &Service{store: s, nextID: idg, now: now}
}

// Create 登记约束。target 语义随类型：
//   - exclude：端元 ID；
//   - ratio：格式 "A:B"，A 为分子端元、B 为分母端元；
//   - bound：端元 ID，range 的 Hi 作为比例上限。
func (s *Service) Create(ct model.ConstraintType, name, target string, r model.Range, note string) (*model.Constraint, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, model.NewError("BAD_REQUEST", "constraint name is required")
	}
	if err := s.validatePayload(ct, target, r); err != nil {
		return nil, err
	}
	c := &model.Constraint{
		ID:        s.nextID(),
		Name:      name,
		Type:      ct,
		Target:    target,
		Range:     r,
		Status:    model.ConstraintEnabled,
		Note:      note,
		Version:   1,
		CreatedAt: s.now(),
		UpdatedAt: s.now(),
	}
	if err := s.store.CreateConstraint(c); err != nil {
		return nil, err
	}
	return c, nil
}

// validatePayload 校验约束载荷合法性。
func (s *Service) validatePayload(ct model.ConstraintType, target string, r model.Range) error {
	if target = strings.TrimSpace(target); target == "" {
		return model.NewError("BAD_REQUEST", "constraint target is required")
	}
	if math.IsNaN(r.Lo) || math.IsNaN(r.Hi) || math.IsInf(r.Lo, 0) || math.IsInf(r.Hi, 0) || !r.Valid() {
		return model.NewError("INVALID_RANGE", "constraint range must be finite and non-empty")
	}
	switch ct {
	case model.ConstraintExclude:
		if r.Lo != 0 || r.Hi != 0 {
			return model.NewError("BAD_REQUEST", "exclude constraint range must be [0,0]")
		}
	case model.ConstraintRatio:
		parts := strings.Split(target, ":")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return model.NewError("BAD_REQUEST", "ratio constraint target must be \"A:B\"")
		}
		if r.Lo < 0 {
			return model.NewError("NEGATIVE_MASS", "ratio lower bound must be non-negative")
		}
	case model.ConstraintBound:
		if r.Lo != 0 || r.Hi < 0 {
			return model.NewError("BAD_REQUEST", "bound constraint range must be [0, upper] with upper >= 0")
		}
	default:
		return model.NewError("BAD_REQUEST", "unknown constraint type %q", ct)
	}
	return nil
}

// Get 按 ID 查询约束。
func (s *Service) Get(id string) (*model.Constraint, error) {
	return s.store.GetConstraint(id)
}

// List 列出全部约束（按名称排序）。
func (s *Service) List() ([]model.Constraint, error) {
	all, err := s.store.ListConstraints()
	if err != nil {
		return nil, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	return all, nil
}

// ListActive 列出参与编译的约束（状态为启用）。
func (s *Service) ListActive() ([]model.Constraint, error) {
	all, err := s.store.ListConstraints()
	if err != nil {
		return nil, err
	}
	var out []model.Constraint
	for _, c := range all {
		if c.Status == model.ConstraintEnabled {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// transition 通用状态流转。
func (s *Service) transition(id string, to model.ConstraintStatus) (*model.Constraint, error) {
	c, err := s.store.GetConstraint(id)
	if err != nil {
		return nil, err
	}
	switch to {
	case model.ConstraintEnabled:
		if c.Status == model.ConstraintRevoked {
			return nil, model.NewError("INVALID_TRANSITION", "revoked constraint cannot be re-enabled")
		}
	case model.ConstraintRevoked:
		// 任意状态均可撤销。
	default:
		// relaxed / conflict 为中间标记，允许自由切换。
	}
	c.Status = to
	c.Version++
	c.UpdatedAt = s.now()
	if err := s.store.UpdateConstraint(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Enable 启用约束。
func (s *Service) Enable(id string) (*model.Constraint, error) {
	return s.transition(id, model.ConstraintEnabled)
}

// Relax 松弛约束：不参与编译但保留。
func (s *Service) Relax(id string) (*model.Constraint, error) {
	return s.transition(id, model.ConstraintRelaxed)
}

// MarkConflict 标记约束为冲突（由求解流程在识别到冲突核心时调用）。
func (s *Service) MarkConflict(id string) (*model.Constraint, error) {
	return s.transition(id, model.ConstraintConflict)
}

// Revoke 撤销约束：永久失效。
func (s *Service) Revoke(id string) (*model.Constraint, error) {
	return s.transition(id, model.ConstraintRevoked)
}
