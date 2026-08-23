// Package measure 管理样品测量：测量区间登记、不确定度校验、误差过大判定、封存。
package measure

import (
	"math"
	"sort"
	"strings"
	"time"

	"task177-isomix/internal/model"
)

// IDGen 生成样品 ID。
type IDGen func() string

// TimeNow 返回当前时间。
type TimeNow func() time.Time

// Store 样品持久化接口（由 store 包实现）。
type Store interface {
	CreateSample(s *model.Sample) error
	GetSample(id string) (*model.Sample, error)
	UpdateSample(s *model.Sample) error
	ListSamples() ([]model.Sample, error)
	CountSamples() (int, error)
}

// Service 测量领域服务。
type Service struct {
	store  Store
	nextID IDGen
	now    TimeNow
	// MaxRelativeWidth 误差过大阈值：某同位素测量区间相对宽度超过该值判为误差过大。
	MaxRelativeWidth float64
}

// NewService 构造测量服务。
func NewService(s Store, idg IDGen, now TimeNow) *Service {
	if idg == nil {
		idg = func() string { return "sp_" + strings.ToLower(model.RandToken(8)) }
	}
	if now == nil {
		now = time.Now
	}
	return &Service{store: s, nextID: idg, now: now, MaxRelativeWidth: 0.5}
}

// Create 登记样品测量：校验测量区间、协方差，状态为待测。
func (s *Service) Create(name string, measurements map[string]model.Range, covariance []float64, meta map[string]string) (*model.Sample, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, model.NewError("BAD_REQUEST", "sample name is required")
	}
	if err := ValidateMeasurements(measurements); err != nil {
		return nil, err
	}
	d := len(measurements)
	if len(covariance) > 0 {
		if err := CheckCovariance(covariance, d); err != nil {
			return nil, err
		}
	}
	sp := &model.Sample{
		ID:           s.nextID(),
		Name:         name,
		Status:       model.SamplePending,
		Measurements: cloneRanges(measurements),
		Covariance:   append([]float64(nil), covariance...),
		Dimension:    d,
		Meta:         meta,
		Version:      1,
		CreatedAt:    s.now(),
		UpdatedAt:    s.now(),
	}
	if err := s.store.CreateSample(sp); err != nil {
		return nil, err
	}
	return sp, nil
}

// Get 按 ID 查询样品。
func (s *Service) Get(id string) (*model.Sample, error) {
	return s.store.GetSample(id)
}

// List 列出全部样品（按名称排序）。
func (s *Service) List() ([]model.Sample, error) {
	all, err := s.store.ListSamples()
	if err != nil {
		return nil, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	return all, nil
}

// UpdateMeasurement 修订测量区间：版本自增；已封存样品禁止修订。
func (s *Service) UpdateMeasurement(id string, measurements map[string]model.Range, covariance []float64) (*model.Sample, error) {
	sp, err := s.store.GetSample(id)
	if err != nil {
		return nil, err
	}
	if sp.Status == model.SampleSealed {
		return nil, model.ErrSealedSample
	}
	if err := ValidateMeasurements(measurements); err != nil {
		return nil, err
	}
	if len(covariance) > 0 {
		if err := CheckCovariance(covariance, len(measurements)); err != nil {
			return nil, err
		}
	}
	sp.Measurements = cloneRanges(measurements)
	sp.Covariance = append([]float64(nil), covariance...)
	sp.Dimension = len(measurements)
	sp.Version++
	sp.Status = model.SamplePending
	sp.UpdatedAt = s.now()
	if err := s.store.UpdateSample(sp); err != nil {
		return nil, err
	}
	return sp, nil
}

// Check 评估样品测量质量：不确定度可接受则置为可求解，否则置为误差过大。
func (s *Service) Check(id string) (*model.Sample, error) {
	sp, err := s.store.GetSample(id)
	if err != nil {
		return nil, err
	}
	if sp.Status == model.SampleSealed {
		return nil, model.ErrSealedSample
	}
	if s.overspread(sp) {
		sp.Status = model.SampleOverspread
	} else {
		sp.Status = model.SampleSolvable
	}
	sp.UpdatedAt = s.now()
	if err := s.store.UpdateSample(sp); err != nil {
		return nil, err
	}
	return sp, nil
}

// Seal 封存样品：禁止后续求解与修订。
func (s *Service) Seal(id string) (*model.Sample, error) {
	sp, err := s.store.GetSample(id)
	if err != nil {
		return nil, err
	}
	if sp.Status == model.SampleSealed {
		return sp, nil
	}
	sp.Status = model.SampleSealed
	sp.UpdatedAt = s.now()
	if err := s.store.UpdateSample(sp); err != nil {
		return nil, err
	}
	return sp, nil
}

// MarkSolved 由求解流程在产生已接受结果后调用。
func (s *Service) MarkSolved(id string) error {
	sp, err := s.store.GetSample(id)
	if err != nil {
		return err
	}
	if sp.Status == model.SampleSealed {
		return nil
	}
	sp.Status = model.SampleSolved
	sp.UpdatedAt = s.now()
	return s.store.UpdateSample(sp)
}

// overspread 判断是否存在相对宽度超阈值的同位素测量区间。
func (s *Service) overspread(sp *model.Sample) bool {
	for _, r := range sp.Measurements {
		center := (r.Lo + r.Hi) / 2
		if math.Abs(center) < 1e-12 {
			// 中心接近零的比率（如比值接近 0）无法定义相对宽度，视为可接受。
			continue
		}
		rel := r.Width() / math.Abs(center)
		if rel > s.MaxRelativeWidth {
			return true
		}
	}
	return false
}

func cloneRanges(m map[string]model.Range) map[string]model.Range {
	out := make(map[string]model.Range, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
