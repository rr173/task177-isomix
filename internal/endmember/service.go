// Package endmember 管理端元：组成区间登记、维度校验、状态流转（草拟→已校验→可用/已排除）。
package endmember

import (
	"sort"
	"strings"
	"time"

	"task177-isomix/internal/model"
)

// IDGen 生成端元 ID（注入式，便于测试与自检）。
type IDGen func() string

// TimeNow 返回当前时间（注入式）。
type TimeNow func() time.Time

// Store 端元持久化接口（由 store 包实现，避免循环依赖）。
type Store interface {
	CreateEndmember(e *model.Endmember) error
	GetEndmember(id string) (*model.Endmember, error)
	UpdateEndmember(e *model.Endmember) error
	ListEndmembers() ([]model.Endmember, error)
	CountEndmembers() (int, error)
}

// Service 端元领域服务。
type Service struct {
	store   Store
	nextID  IDGen
	now     TimeNow
}

// NewService 构造端元服务。
func NewService(s Store, idg IDGen, now TimeNow) *Service {
	if idg == nil {
		idg = defaultIDGen
	}
	if now == nil {
		now = time.Now
	}
	return &Service{store: s, nextID: idg, now: now}
}

func defaultIDGen() string {
	return "em_" + strings.ToLower(model.RandToken(8))
}

// Create 登记新端元：校验组成区间与维度，初始状态为草拟。
func (s *Service) Create(name string, components map[string]model.Range, meta map[string]string) (*model.Endmember, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, model.NewError("BAD_REQUEST", "endmember name is required")
	}
	if err := ValidateComponents(components); err != nil {
		return nil, err
	}
	e := &model.Endmember{
		ID:         s.nextID(),
		Name:       name,
		Status:     model.EndmemberDraft,
		Components: cloneRanges(components),
		Dimension:  len(components),
		Meta:       meta,
		Version:    1,
		CreatedAt:  s.now(),
		UpdatedAt:  s.now(),
	}
	if err := s.store.CreateEndmember(e); err != nil {
		return nil, err
	}
	return e, nil
}

// Get 按 ID 查询端元。
func (s *Service) Get(id string) (*model.Endmember, error) {
	return s.store.GetEndmember(id)
}

// List 列出全部端元（按名称排序，稳定输出）。
func (s *Service) List() ([]model.Endmember, error) {
	all, err := s.store.ListEndmembers()
	if err != nil {
		return nil, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	return all, nil
}

// ListAvailable 列出可参与求解的端元（状态为可用）。
func (s *Service) ListAvailable() ([]model.Endmember, error) {
	all, err := s.store.ListEndmembers()
	if err != nil {
		return nil, err
	}
	var out []model.Endmember
	for _, e := range all {
		if e.Status == model.EndmemberAvailable {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Revise 修订端元组成区间：版本自增，状态回退到草拟（修订只触发新求解，不影响已发布报告）。
func (s *Service) Revise(id string, components map[string]model.Range) (*model.Endmember, error) {
	e, err := s.store.GetEndmember(id)
	if err != nil {
		return nil, err
	}
	if err := ValidateComponents(components); err != nil {
		return nil, err
	}
	if e.Status == model.EndmemberExcluded {
		return nil, model.NewError("INVALID_TRANSITION", "excluded endmember cannot be revised; recreate it")
	}
	e.Components = cloneRanges(components)
	e.Dimension = len(components)
	e.Version++
	e.Status = model.EndmemberDraft
	e.UpdatedAt = s.now()
	if err := s.store.UpdateEndmember(e); err != nil {
		return nil, err
	}
	return e, nil
}

// Validate 把端元从草拟推进到已校验。
func (s *Service) Validate(id string) (*model.Endmember, error) {
	e, err := s.store.GetEndmember(id)
	if err != nil {
		return nil, err
	}
	if e.Status != model.EndmemberDraft && e.Status != model.EndmemberValidated {
		return nil, model.NewError("INVALID_TRANSITION", "only draft/validated endmember can be validated, current=%s", e.Status)
	}
	e.Status = model.EndmemberValidated
	e.UpdatedAt = s.now()
	if err := s.store.UpdateEndmember(e); err != nil {
		return nil, err
	}
	return e, nil
}

// MakeAvailable 把端元推进到可用（要求先完成校验）。
func (s *Service) MakeAvailable(id string) (*model.Endmember, error) {
	e, err := s.store.GetEndmember(id)
	if err != nil {
		return nil, err
	}
	if e.Status != model.EndmemberValidated && e.Status != model.EndmemberAvailable {
		return nil, model.NewError("INVALID_TRANSITION", "must validate before making available, current=%s", e.Status)
	}
	e.Status = model.EndmemberAvailable
	e.UpdatedAt = s.now()
	if err := s.store.UpdateEndmember(e); err != nil {
		return nil, err
	}
	return e, nil
}

// Exclude 把端元排除出候选集（已发布报告不受影响，因快照已冻结）。
func (s *Service) Exclude(id string) (*model.Endmember, error) {
	e, err := s.store.GetEndmember(id)
	if err != nil {
		return nil, err
	}
	e.Status = model.EndmemberExcluded
	e.UpdatedAt = s.now()
	if err := s.store.UpdateEndmember(e); err != nil {
		return nil, err
	}
	return e, nil
}

// cloneRanges 深拷贝组成区间映射。
func cloneRanges(m map[string]model.Range) map[string]model.Range {
	out := make(map[string]model.Range, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
