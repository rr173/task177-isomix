package solve

import (
	"task177-isomix/internal/hashutil"
	"task177-isomix/internal/model"
)

// ResumeIncomplete 重启后恢复未完成求解（状态为 queued / running 的记录）。
//
// 恢复语义：
//  1. 从持久化的样品/端元/约束 ID 集合重新加载输入；
//  2. 重新计算输入哈希并与记录比对——哈希失配意味着输入在求解途中被篡改，
//     标记为数值不稳定并附证据，绝不静默重算；
//  3. 哈希一致则重新执行求解，更新终态。
//
// 返回恢复的记录数与遇到的错误（错误为软错误，不中断其余恢复）。
func (s *Service) ResumeIncomplete() (recovered int, errs []error) {
	all, err := s.solutions.ListSolutions()
	if err != nil {
		return 0, []error{err}
	}
	for i := range all {
		sol := &all[i]
		if sol.Status != model.SolutionQueued && sol.Status != model.SolutionRunning {
			continue
		}
		if err := s.resumeOne(sol); err != nil {
			errs = append(errs, err)
			continue
		}
		recovered++
	}
	return recovered, errs
}

// resumeOne 恢复单条求解记录。
func (s *Service) resumeOne(sol *model.Solution) error {
	sp, err := s.meas.Get(sol.SampleID)
	if err != nil {
		sol.Status = model.SolutionUnstable
		sol.Error = "resume: sample reload failed: " + err.Error()
		return s.solutions.UpdateSolution(sol)
	}
	ems := make([]model.Endmember, 0, len(sol.EndmemberIDs))
	for _, id := range sol.EndmemberIDs {
		e, err := s.endm.Get(id)
		if err != nil {
			sol.Status = model.SolutionUnstable
			sol.Error = "resume: endmember reload failed: " + err.Error()
			return s.solutions.UpdateSolution(sol)
		}
		ems = append(ems, *e)
	}
	cs := make([]model.Constraint, 0, len(sol.ConstraintIDs))
	for _, id := range sol.ConstraintIDs {
		c, err := s.cons.Get(id)
		if err != nil {
			sol.Status = model.SolutionUnstable
			sol.Error = "resume: constraint reload failed: " + err.Error()
			return s.solutions.UpdateSolution(sol)
		}
		cs = append(cs, *c)
	}

	// 输入哈希校验：持久化摘要必须与重算一致。
	h, err := hashutil.InputHash(*sp, ems, cs)
	if err != nil {
		sol.Status = model.SolutionUnstable
		sol.Error = "resume: hash recompute failed: " + err.Error()
		return s.solutions.UpdateSolution(sol)
	}
	if h != sol.InputHash {
		sol.Status = model.SolutionUnstable
		sol.Error = "resume: input hash mismatch; persisted input was mutated mid-flight"
		return s.solutions.UpdateSolution(sol)
	}

	// 重新执行求解。
	s.execute(sol)
	return s.solutions.UpdateSolution(sol)
}
