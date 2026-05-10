// Package chain — Phase 16. Use-case service для Prompt Chains.
//
// Цепочка — граф промпт-шагов с явными переходами (next_step_id для prompt-шагов
// и branches[].next_step_id для fork-шагов). Tree-структура: одна корневая
// последовательность + ветки в каждом fork. Любая вложенность.
//
// Run-mode проходит шаги по графу: prompt → next_step_id, fork → юзер выбирает
// ветку в UI. Output текущего шага становится переменной для следующего.
package chain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"promptvault/internal/infrastructure/metrics"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	activityuc "promptvault/internal/usecases/activity"
	quotauc "promptvault/internal/usecases/quota"
	"promptvault/internal/usecases/teamcheck"
)

const (
	maxNameLen        = 255
	maxDescriptionLen = 2000
	maxStepNameLen    = 255
)

type Service struct {
	chains  repo.ChainRepository
	prompts repo.PromptRepository
	teams   repo.TeamRepository
	quotas  *quotauc.Service
	// activity — опциональный team activity feed (Phase 14).
	// Nil-safe через LogSafe; nil в тестах и в personal-only сборках.
	activity *activityuc.Service
}

func NewService(chains repo.ChainRepository, prompts repo.PromptRepository, teams repo.TeamRepository, quotas *quotauc.Service) *Service {
	return &Service{chains: chains, prompts: prompts, teams: teams, quotas: quotas}
}

// SetActivity подключает team_activity_log хуки. Не передаём в NewService,
// потому что не все callers (тесты) хотят активити-зависимость.
func (s *Service) SetActivity(activity *activityuc.Service) {
	s.activity = activity
}

// --- Chain CRUD ---

func (s *Service) Create(ctx context.Context, userID uint, name, description string, teamID *uint) (*models.PromptChain, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxNameLen {
		return nil, ErrInvalidName
	}
	if len(description) > maxDescriptionLen {
		return nil, ErrInvalidDescription
	}
	// Pack T: scope-aware квота. Цепочки в команде идут в team-pool по плану owner'а.
	if s.quotas != nil {
		if teamID != nil {
			team, err := s.teams.GetByID(ctx, *teamID)
			if err != nil {
				return nil, err
			}
			if err := s.quotas.CheckTeamChainQuota(ctx, *teamID, team.CreatedBy); err != nil {
				return nil, err
			}
		} else {
			if err := s.quotas.CheckChainQuota(ctx, userID); err != nil {
				return nil, err
			}
		}
	}
	if err := teamcheck.RequireEditor(ctx, s.teams, teamID, userID); err != nil {
		return nil, mapTeamError(err)
	}

	c := &models.PromptChain{
		UserID:      userID,
		TeamID:      teamID,
		Name:        name,
		Description: description,
	}
	if err := s.chains.Create(ctx, c); err != nil {
		return nil, err
	}
	scope := "personal"
	if teamID != nil {
		scope = "team"
	}
	metrics.ChainsCreated.WithLabelValues(scope).Inc()
	slog.Info("chain.created", "chain_id", c.ID, "user_id", userID, "team_id", teamID)
	if teamID != nil {
		s.activity.LogSafe(ctx, activityuc.Event{
			TeamID:      *teamID,
			ActorID:     userID,
			EventType:   models.ActivityChainCreated,
			TargetType:  models.TargetChain,
			TargetID:    &c.ID,
			TargetLabel: c.Name,
		})
	}
	return c, nil
}

func (s *Service) GetByID(ctx context.Context, chainID, userID uint) (*models.PromptChain, error) {
	c, err := s.chains.GetByID(ctx, chainID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := s.checkReadAccess(ctx, c, userID); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) GetByIDWithSteps(ctx context.Context, chainID, userID uint) (*models.PromptChain, error) {
	c, err := s.chains.GetByIDWithSteps(ctx, chainID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := s.checkReadAccess(ctx, c, userID); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) List(ctx context.Context, userID uint, teamIDs []uint, limit, offset int) ([]models.PromptChain, int64, error) {
	if len(teamIDs) > 0 {
		if err := teamcheck.RequireMembership(ctx, s.teams, teamIDs, userID); err != nil {
			return nil, 0, mapTeamError(err)
		}
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.chains.ListByUser(ctx, userID, teamIDs, limit, offset)
}

// ListWithStats — расширенный list для UI карточек: каждая цепочка идёт со
// step_count, has_branching, saved_runs_count и steps_preview. Один SELECT
// через LATERAL в репозитории — без N+1.
func (s *Service) ListWithStats(ctx context.Context, userID uint, teamIDs []uint, limit, offset int) ([]models.PromptChainListRow, int64, error) {
	if len(teamIDs) > 0 {
		if err := teamcheck.RequireMembership(ctx, s.teams, teamIDs, userID); err != nil {
			return nil, 0, mapTeamError(err)
		}
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.chains.ListByUserWithStats(ctx, userID, teamIDs, limit, offset)
}

func (s *Service) Update(ctx context.Context, chainID, userID uint, name, description string) (*models.PromptChain, error) {
	c, err := s.GetByID(ctx, chainID, userID)
	if err != nil {
		return nil, err
	}
	if err := s.checkEditAccess(ctx, c, userID); err != nil {
		return nil, err
	}

	if name != "" {
		name = strings.TrimSpace(name)
		if len(name) > maxNameLen {
			return nil, ErrInvalidName
		}
		c.Name = name
	}
	if len(description) > maxDescriptionLen {
		return nil, ErrInvalidDescription
	}
	c.Description = description

	if err := s.chains.Update(ctx, c); err != nil {
		return nil, err
	}
	if c.TeamID != nil {
		s.activity.LogSafe(ctx, activityuc.Event{
			TeamID:      *c.TeamID,
			ActorID:     userID,
			EventType:   models.ActivityChainUpdated,
			TargetType:  models.TargetChain,
			TargetID:    &c.ID,
			TargetLabel: c.Name,
		})
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, chainID, userID uint) error {
	c, err := s.GetByID(ctx, chainID, userID)
	if err != nil {
		return err
	}
	if err := s.checkEditAccess(ctx, c, userID); err != nil {
		return err
	}
	hasActive, err := s.chains.HasActiveExecutions(ctx, chainID)
	if err != nil {
		return err
	}
	if hasActive {
		return ErrChainHasActiveExecutions
	}
	if err := s.chains.SoftDelete(ctx, chainID); err != nil {
		return err
	}
	if c.TeamID != nil {
		s.activity.LogSafe(ctx, activityuc.Event{
			TeamID:      *c.TeamID,
			ActorID:     userID,
			EventType:   models.ActivityChainDeleted,
			TargetType:  models.TargetChain,
			TargetID:    &c.ID,
			TargetLabel: c.Name,
		})
	}
	return nil
}

// --- Steps ---

// AddStepInput — input для AddStep. StepType="" → "prompt".
//
// Куда вставить новый шаг (взаимоисключающие, для tree-editor):
//   - AfterStepID:                 после указанного prompt-шага (новый забирает
//     старый next_step_id; afterStep.NextStepID = новый.id).
//   - ParentForkID + BranchIndex:  как первый шаг указанной ветки fork-шага
//     (новый забирает старый branch.next_step_id; branch.next_step_id = новый.id).
//   - Ничего не задано:            tail-mode — добавить в конец главной линии
//     (после последнего prompt-шага без next_step_id). Если шагов нет — root.
//
// Если новый шаг сам — fork, anchor.next должен быть nil (иначе хвост потеряется).
type AddStepInput struct {
	// PromptID — обязателен для prompt-шага; для fork-шага — nil (fork это
	// контейнер с ветками, без своего промпта).
	PromptID         *uint
	Name             string
	VariableMapping  json.RawMessage
	ManualCheckpoint bool
	StepType         string
	Conditions       json.RawMessage

	AfterStepID  *uint
	ParentForkID *uint
	BranchIndex  *int
}

func (s *Service) AddStep(ctx context.Context, chainID, userID uint, in AddStepInput) (*models.PromptChainStep, error) {
	c, err := s.GetByID(ctx, chainID, userID)
	if err != nil {
		return nil, err
	}
	if err := s.checkEditAccess(ctx, c, userID); err != nil {
		return nil, err
	}

	stepType := in.StepType
	if stepType == "" {
		stepType = models.StepTypePrompt
	}
	switch stepType {
	case models.StepTypePrompt, models.StepTypeFork:
	default:
		return nil, ErrInvalidForkStep
	}

	// prompt-шаг — promptID обязателен и должен существовать; fork-шаг — nil.
	if stepType == models.StepTypePrompt {
		if in.PromptID == nil || *in.PromptID == 0 {
			return nil, ErrPromptNotFound
		}
		// MN-37: GetMeta вместо GetByID — нам нужно только подтверждение
		// существования промпта, Tags/Collections не используются. Экономит
		// 2 SELECT'а на каждый AddStep.
		if _, err := s.prompts.GetMeta(ctx, *in.PromptID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, ErrPromptNotFound
			}
			return nil, err
		}
	}

	if s.quotas != nil {
		if err := s.quotas.CheckChainStepQuota(ctx, userID, chainID); err != nil {
			return nil, err
		}
	}

	if err := validateVariableMapping(in.VariableMapping); err != nil {
		return nil, err
	}
	if len(in.Name) > maxStepNameLen {
		return nil, ErrInvalidName
	}

	if stepType == models.StepTypeFork {
		if !s.isMaxTierForChain(ctx, c, userID) {
			return nil, ErrForkRequiresMax
		}
		if len(in.Conditions) == 0 {
			return nil, ErrInvalidForkStep
		}
	}

	steps, err := s.chains.ListStepsByChain(ctx, chainID)
	if err != nil {
		return nil, err
	}

	// Определить anchor: prompt-шаг (anchorPromptStep) или fork-шаг + индекс
	// ветки (anchorFork/anchorBranchIdx). Унаследовать nextStepID из anchor.
	var (
		anchorPromptStep *models.PromptChainStep
		anchorFork       *models.PromptChainStep
		anchorBranchIdx  int
		nextStepID       *uint
	)

	switch {
	case in.AfterStepID != nil:
		for i := range steps {
			if steps[i].ID == *in.AfterStepID {
				anchorPromptStep = &steps[i]
				break
			}
		}
		if anchorPromptStep == nil {
			return nil, ErrStepNotFound
		}
		if anchorPromptStep.StepType != models.StepTypePrompt {
			return nil, ErrCannotInsertAfterFork
		}
		nextStepID = anchorPromptStep.NextStepID

	case in.ParentForkID != nil && in.BranchIndex != nil:
		for i := range steps {
			if steps[i].ID == *in.ParentForkID {
				anchorFork = &steps[i]
				break
			}
		}
		if anchorFork == nil {
			return nil, ErrStepNotFound
		}
		if anchorFork.StepType != models.StepTypeFork {
			return nil, ErrParentNotFork
		}
		anchorBranchIdx = *in.BranchIndex
		var conds models.Conditions
		if err := json.Unmarshal(anchorFork.Conditions, &conds); err != nil {
			return nil, ErrInvalidConditions
		}
		if anchorBranchIdx < 0 || anchorBranchIdx >= len(conds.Branches) {
			return nil, ErrChosenBranchNotFound
		}
		nextStepID = conds.Branches[anchorBranchIdx].NextStepID

	default:
		// tail-mode: ищем prompt-шаг с NextStepID==nil. Среди нескольких таких
		// (напр. концы веток) берём с наибольшей position — это последний
		// «по времени создания» tail главной линии.
		var lastTail *models.PromptChainStep
		for i := range steps {
			if steps[i].StepType == models.StepTypePrompt && steps[i].NextStepID == nil {
				if lastTail == nil || steps[i].Position > lastTail.Position {
					lastTail = &steps[i]
				}
			}
		}
		if lastTail != nil {
			anchorPromptStep = lastTail
		}
		// Если шагов нет (или все — fork без хвостов) — новый шаг становится
		// корнем; nextStepID остаётся nil; anchor* nil.
	}

	// Если новый шаг — fork, у anchor не должно быть хвоста, который был бы потерян.
	if stepType == models.StepTypeFork && nextStepID != nil {
		return nil, ErrInsertForkLosesTail
	}

	// Если новый — fork, валидируем branches против существующих ID.
	if stepType == models.StepTypeFork {
		validIDs := make(map[uint]struct{}, len(steps))
		for _, st := range steps {
			validIDs[st.ID] = struct{}{}
		}
		if err := validateForkBranches(in.Conditions, validIDs); err != nil {
			return nil, err
		}
	}

	mapping := in.VariableMapping
	if len(mapping) == 0 {
		mapping = json.RawMessage(`{}`)
	}

	count, err := s.chains.CountStepsByChain(ctx, chainID)
	if err != nil {
		return nil, err
	}

	step := &models.PromptChainStep{
		ChainID:          chainID,
		Position:         int(count) + 1,
		PromptID:         in.PromptID,
		Name:             in.Name,
		VariableMapping:  mapping,
		ManualCheckpoint: in.ManualCheckpoint,
		StepType:         stepType,
		Conditions:       in.Conditions,
		NextStepID:       nextStepID,
	}
	// fork-шаг — без промпта.
	if stepType == models.StepTypeFork {
		step.PromptID = nil
	}

	// Все mutate-операции (insert step, update anchor, cycle-check) внутри одной
	// транзакции. При cycle-detect ошибке tx откатится — шаг не появится, anchor
	// не сместится, граф останется в исходном состоянии.
	txErr := s.chains.InTransaction(ctx, func(txRepo repo.ChainRepository) error {
		if err := txRepo.AddStep(ctx, step); err != nil {
			return err
		}
		if anchorPromptStep != nil {
			anchorPromptStep.NextStepID = &step.ID
			if err := txRepo.UpdateStep(ctx, anchorPromptStep); err != nil {
				return err
			}
		}
		if anchorFork != nil {
			var conds models.Conditions
			if err := json.Unmarshal(anchorFork.Conditions, &conds); err != nil {
				return ErrInvalidConditions
			}
			conds.Branches[anchorBranchIdx].NextStepID = &step.ID
			raw, err := json.Marshal(conds)
			if err != nil {
				return err
			}
			anchorFork.Conditions = raw
			if err := txRepo.UpdateStep(ctx, anchorFork); err != nil {
				return err
			}
		}
		// Cycle check внутри tx — если граф зацикливается, tx откатится.
		allSteps, err := txRepo.ListStepsByChain(ctx, chainID)
		if err != nil {
			return err
		}
		if cycleErr := detectCycles(allSteps); cycleErr != nil {
			return cycleErr
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return step, nil
}

// UpdateStepInput — изменяемые поля шага. PromptID нельзя менять (создайте новый шаг).
// StepType: "" = не менять; "prompt"/"fork" — переключить тип.
type UpdateStepInput struct {
	Name             string
	VariableMapping  json.RawMessage
	ManualCheckpoint bool
	StepType         string
	Conditions       json.RawMessage
}

func (s *Service) UpdateStep(ctx context.Context, chainID, stepID, userID uint, in UpdateStepInput) (*models.PromptChainStep, error) {
	c, err := s.GetByID(ctx, chainID, userID)
	if err != nil {
		return nil, err
	}
	if err := s.checkEditAccess(ctx, c, userID); err != nil {
		return nil, err
	}

	step, err := s.chains.GetStepByID(ctx, stepID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrStepNotFound
		}
		return nil, err
	}
	if step.ChainID != chainID {
		return nil, ErrStepNotFound
	}

	if err := validateVariableMapping(in.VariableMapping); err != nil {
		return nil, err
	}
	if len(in.Name) > maxStepNameLen {
		return nil, ErrInvalidName
	}

	step.Name = in.Name
	if len(in.VariableMapping) > 0 {
		step.VariableMapping = in.VariableMapping
	}
	step.ManualCheckpoint = in.ManualCheckpoint

	newType := in.StepType
	if newType == "" {
		newType = step.StepType
	}
	switch newType {
	case models.StepTypeFork:
		if !s.isMaxTierForChain(ctx, c, userID) {
			return nil, ErrForkRequiresMax
		}
		if len(in.Conditions) == 0 {
			return nil, ErrInvalidForkStep
		}
		existingSteps, err := s.chains.ListStepsByChain(ctx, chainID)
		if err != nil {
			return nil, err
		}
		validIDs := make(map[uint]struct{}, len(existingSteps))
		for _, st := range existingSteps {
			validIDs[st.ID] = struct{}{}
		}
		if err := validateForkBranches(in.Conditions, validIDs); err != nil {
			return nil, err
		}
		step.StepType = models.StepTypeFork
		step.Conditions = in.Conditions
	case models.StepTypePrompt:
		step.StepType = models.StepTypePrompt
		step.Conditions = nil
	default:
		return nil, ErrInvalidForkStep
	}

	if err := s.chains.UpdateStep(ctx, step); err != nil {
		return nil, err
	}
	allSteps, err := s.chains.ListStepsByChain(ctx, chainID)
	if err == nil {
		if cycleErr := detectCycles(allSteps); cycleErr != nil {
			return nil, cycleErr
		}
	}
	return step, nil
}

// RemoveStep — удаляет шаг и «зашивает» граф: prompt-предшественники переключаются
// на step.NextStepID; fork-предшественники у которых ветка указывала на step,
// переключают branch.next_step_id на step.NextStepID. Если на удаляемый шаг
// никто не ссылается — просто DELETE.
func (s *Service) RemoveStep(ctx context.Context, chainID, stepID, userID uint) error {
	c, err := s.GetByID(ctx, chainID, userID)
	if err != nil {
		return err
	}
	if err := s.checkEditAccess(ctx, c, userID); err != nil {
		return err
	}
	target, err := s.chains.GetStepByID(ctx, stepID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrStepNotFound
		}
		return err
	}
	if target.ChainID != chainID {
		return ErrStepNotFound
	}

	// Re-link prompt-предшественников.
	if err := s.chains.RelinkPromptPredecessors(ctx, chainID, target.ID, target.NextStepID); err != nil {
		return err
	}

	// Re-link fork-предшественников: загружаем все шаги цепочки и обновляем
	// conditions.branches у тех, кто ссылается на target.ID.
	allSteps, err := s.chains.ListStepsByChain(ctx, chainID)
	if err != nil {
		return err
	}
	for i := range allSteps {
		st := &allSteps[i]
		if st.ID == target.ID || st.StepType != models.StepTypeFork {
			continue
		}
		var conds models.Conditions
		if err := json.Unmarshal(st.Conditions, &conds); err != nil {
			continue
		}
		changed := false
		for j := range conds.Branches {
			if conds.Branches[j].NextStepID != nil && *conds.Branches[j].NextStepID == target.ID {
				conds.Branches[j].NextStepID = target.NextStepID
				changed = true
			}
		}
		if !changed {
			continue
		}
		raw, err := json.Marshal(conds)
		if err != nil {
			return err
		}
		st.Conditions = raw
		if err := s.chains.UpdateStep(ctx, st); err != nil {
			return err
		}
	}

	return s.chains.RemoveStep(ctx, target.ID)
}

// MoveStepUp меняет местами prompt-шаг и его prompt-предшественника в графе.
// Семантика: если до операции X → A → S → B (X — prompt или fork-branch,
// A — prompt-предшественник S, S — двигаемый шаг, B — то что после S),
// после операции: X → S → A → B. Все ссылки переписываются в одной транзакции.
//
// Случай когда у A нет prompt-предшественника (A был root или первым в ветке
// fork): X — это fork-ветка, указывающая на A. Её next_step_id переключается
// на S; A становится вторым в подцепочке.
//
// Fork-шаги двигать нельзя (ErrCannotMoveFork) — фактически они занимают
// «слот ветвления» и при перемещении пришлось бы переписывать целый поддерево.
func (s *Service) MoveStepUp(ctx context.Context, chainID, stepID, userID uint) error {
	c, err := s.GetByID(ctx, chainID, userID)
	if err != nil {
		return err
	}
	if err := s.checkEditAccess(ctx, c, userID); err != nil {
		return err
	}

	return s.chains.InTransaction(ctx, func(txRepo repo.ChainRepository) error {
		steps, err := txRepo.ListStepsByChain(ctx, chainID)
		if err != nil {
			return err
		}
		var target *models.PromptChainStep
		for i := range steps {
			if steps[i].ID == stepID {
				target = &steps[i]
				break
			}
		}
		if target == nil || target.ChainID != chainID {
			return ErrStepNotFound
		}
		if target.StepType != models.StepTypePrompt {
			return ErrCannotMoveFork
		}

		// A — prompt-предшественник target (его next_step_id указывает на target).
		var prev *models.PromptChainStep
		for i := range steps {
			if steps[i].StepType == models.StepTypePrompt && steps[i].NextStepID != nil && *steps[i].NextStepID == target.ID {
				prev = &steps[i]
				break
			}
		}
		if prev == nil {
			// target — корень линейной подцепочки. Перемещать вверх некуда.
			return ErrCannotMoveAtBoundary
		}

		// X — кто указывает на prev. Может быть prompt или fork-ветка.
		var xPrompt *models.PromptChainStep
		var xFork *models.PromptChainStep
		var xForkBranchIdx int
		for i := range steps {
			st := &steps[i]
			if st.StepType == models.StepTypePrompt && st.NextStepID != nil && *st.NextStepID == prev.ID {
				xPrompt = st
				break
			}
		}
		if xPrompt == nil {
			for i := range steps {
				st := &steps[i]
				if st.StepType != models.StepTypeFork || len(st.Conditions) == 0 {
					continue
				}
				var conds models.Conditions
				if err := json.Unmarshal(st.Conditions, &conds); err != nil {
					continue
				}
				for j, b := range conds.Branches {
					if b.NextStepID != nil && *b.NextStepID == prev.ID {
						xFork = st
						xForkBranchIdx = j
						break
					}
				}
				if xFork != nil {
					break
				}
			}
		}

		targetOldNext := target.NextStepID

		// prev.next = target.old_next (B)
		prev.NextStepID = targetOldNext
		if err := txRepo.UpdateStep(ctx, prev); err != nil {
			return err
		}
		// target.next = prev.id
		prevID := prev.ID
		target.NextStepID = &prevID
		if err := txRepo.UpdateStep(ctx, target); err != nil {
			return err
		}
		// X.next = target.id (или X.fork-branch.next = target.id)
		targetID := target.ID
		if xPrompt != nil {
			xPrompt.NextStepID = &targetID
			if err := txRepo.UpdateStep(ctx, xPrompt); err != nil {
				return err
			}
		} else if xFork != nil {
			var conds models.Conditions
			if err := json.Unmarshal(xFork.Conditions, &conds); err != nil {
				return err
			}
			conds.Branches[xForkBranchIdx].NextStepID = &targetID
			raw, err := json.Marshal(conds)
			if err != nil {
				return err
			}
			xFork.Conditions = raw
			if err := txRepo.UpdateStep(ctx, xFork); err != nil {
				return err
			}
		}
		// Если ни prompt-предшественника, ни fork-ветки — prev был root цепочки;
		// после swap target становится root (никто на него не указывает).

		return nil
	})
}

// MoveStepDown — двинуть шаг вниз, выраженный через MoveStepUp следующего.
func (s *Service) MoveStepDown(ctx context.Context, chainID, stepID, userID uint) error {
	c, err := s.GetByID(ctx, chainID, userID)
	if err != nil {
		return err
	}
	if err := s.checkEditAccess(ctx, c, userID); err != nil {
		return err
	}
	target, err := s.chains.GetStepByID(ctx, stepID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrStepNotFound
		}
		return err
	}
	if target.ChainID != chainID {
		return ErrStepNotFound
	}
	if target.StepType != models.StepTypePrompt {
		return ErrCannotMoveFork
	}
	if target.NextStepID == nil {
		return ErrCannotMoveAtBoundary
	}
	next, err := s.chains.GetStepByID(ctx, *target.NextStepID)
	if err != nil {
		return err
	}
	if next.StepType != models.StepTypePrompt {
		// Следом fork — двигать вниз нельзя (мы бы должны были «перепрыгнуть» fork).
		return ErrCannotMoveAtBoundary
	}
	return s.MoveStepUp(ctx, chainID, next.ID, userID)
}

func (s *Service) ReorderSteps(ctx context.Context, chainID, userID uint, stepIDs []uint) error {
	c, err := s.GetByID(ctx, chainID, userID)
	if err != nil {
		return err
	}
	if err := s.checkEditAccess(ctx, c, userID); err != nil {
		return err
	}

	current, err := s.chains.ListStepsByChain(ctx, chainID)
	if err != nil {
		return err
	}
	if len(current) != len(stepIDs) {
		return ErrReorderMismatch
	}
	known := make(map[uint]struct{}, len(current))
	for _, ss := range current {
		known[ss.ID] = struct{}{}
	}
	for _, id := range stepIDs {
		if _, ok := known[id]; !ok {
			return ErrReorderMismatch
		}
	}
	return s.chains.ReorderSteps(ctx, chainID, stepIDs)
}

// --- Executions ---

// StartExecution создаёт новый запуск. Снимает snapshot цепочки + контент всех
// используемых промптов: редактирование во время run не ломает execution.
//
// Поле CurrentStep сохраняется для backward-compat (старые executions могут
// существовать), но обход теперь по step.ID, см. AdvanceStep.
func (s *Service) StartExecution(ctx context.Context, chainID, userID uint, initialVars json.RawMessage) (*models.PromptChainExecution, error) {
	c, err := s.GetByIDWithSteps(ctx, chainID, userID)
	if err != nil {
		return nil, err
	}
	if len(c.Steps) == 0 {
		return nil, ErrEmptyChain
	}

	root := findRootStep(c.Steps)
	if root == nil {
		return nil, ErrEmptyChain
	}

	// CR-11: N+1 fix. GetByIDWithSteps делает Preload("Steps.Prompt"),
	// поэтому промпты уже подгружены — раньше тут шёл per-step
	// `s.prompts.GetByID()` (51 query на цепочке из 50 шагов × 3 SELECT'а
	// внутри = ~150 round-trips на один Start). Теперь читаем preloaded.
	// Если step.Prompt == nil (soft-delete после Preload или missing FK)
	// — fallback на GetByID для корректной диагностики ErrPromptNotFound.
	promptContents := make(map[uint]string, len(c.Steps))
	for _, step := range c.Steps {
		if step.PromptID == nil {
			continue // fork-шаг — без промпта
		}
		if _, ok := promptContents[*step.PromptID]; ok {
			continue
		}
		if step.Prompt != nil && step.Prompt.ID != 0 {
			promptContents[*step.PromptID] = step.Prompt.Content
			continue
		}
		// Fallback: prompt не preloaded или soft-deleted — proverим явно.
		p, err := s.prompts.GetByID(ctx, *step.PromptID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, ErrPromptNotFound
			}
			return nil, err
		}
		promptContents[*step.PromptID] = p.Content
	}

	snapshot := models.ChainSnapshot{
		Chain:          *c,
		Steps:          c.Steps,
		PromptContents: promptContents,
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}

	vars := initialVars
	if len(vars) == 0 {
		vars = json.RawMessage(`{}`)
	}

	exec := &models.PromptChainExecution{
		ChainID:       chainID,
		UserID:        userID,
		CurrentStep:   root.Position,
		Variables:     vars,
		StepOutputs:   json.RawMessage(`{}`),
		ChainSnapshot: snapshotJSON,
		Status:        models.ChainExecutionStatusInProgress,
	}
	if err := s.chains.CreateExecution(ctx, exec); err != nil {
		return nil, err
	}
	metrics.ChainExecutionsStarted.Inc()
	slog.Info("chain.execution.started",
		"chain_id", chainID,
		"exec_id", exec.ID,
		"user_id", userID,
		"steps_count", len(c.Steps),
	)
	if c.TeamID != nil {
		s.activity.LogSafe(ctx, activityuc.Event{
			TeamID:      *c.TeamID,
			ActorID:     userID,
			EventType:   models.ActivityChainExecutionStarted,
			TargetType:  models.TargetChain,
			TargetID:    &c.ID,
			TargetLabel: c.Name,
		})
	}
	return exec, nil
}

// ListExecutions — последние N запусков цепочки. Видны всем кто имеет
// read-access к chain (owner личной цепочки; owner/editor/viewer команды).
// Initiator-фильтр НЕ применяется: история запусков — team-property,
// в отличие от GetExecution/AdvanceStep (initiator-only, эти меняют state).
//
// limit — page size (1..100). Реальный объём упирается в plan.MaxSavedExecutions
// (Free=3, Pro=50, Max=1000 — после миграции 000067) — за пределами лимита
// старые записи уже удалены planом ретеншна (TODO: ретеншн в repo).
func (s *Service) ListExecutions(ctx context.Context, chainID, userID uint, limit int) ([]models.PromptChainExecution, error) {
	chain, err := s.chains.GetByID(ctx, chainID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := s.checkReadAccess(ctx, chain, userID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.chains.ListExecutionsByChain(ctx, chainID, limit)
}

func (s *Service) GetExecution(ctx context.Context, execID, userID uint) (*models.PromptChainExecution, error) {
	exec, err := s.chains.GetExecutionByID(ctx, execID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrExecutionNotFound
		}
		return nil, err
	}
	// Initiator-only: execution stateful, advance делает только тот, кто запустил.
	// Editor команды не может «перехватить» чужой in-progress execution.
	if exec.UserID != userID {
		return nil, ErrForbidden
	}
	// Актуальный read-access к цепочке (security fix W3): если юзера выгнали
	// из команды между Start и Advance — должен получить 403. Snapshot
	// защищает структуру, но не право доступа.
	chain, err := s.chains.GetByID(ctx, exec.ChainID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			// Цепочка soft-deleted — initiator всё ещё может завершить execution
			// (snapshot живой). Это by design: удаление chain не прерывает run.
			return exec, nil
		}
		return nil, err
	}
	if err := s.checkReadAccess(ctx, chain, userID); err != nil {
		return nil, err
	}
	return exec, nil
}

// AdvanceStep записывает output текущего шага и переводит execution на следующий.
// Для fork-шагов нужен chosenBranchIdx (юзер выбрал ветку в UI).
// Для prompt-шагов chosenBranchIdx игнорируется. Использует snapshot, не текущее
// состояние chain. Если следующего шага нет — статус → completed.
func (s *Service) AdvanceStep(ctx context.Context, execID, userID uint, stepOutput string, chosenBranchIdx *int) (*models.PromptChainExecution, error) {
	exec, err := s.GetExecution(ctx, execID, userID)
	if err != nil {
		return nil, err
	}
	if exec.Status != models.ChainExecutionStatusInProgress {
		return nil, ErrExecutionAlreadyCompleted
	}

	var snap models.ChainSnapshot
	if err := json.Unmarshal(exec.ChainSnapshot, &snap); err != nil {
		return nil, err
	}

	var currentStep *models.PromptChainStep
	for i := range snap.Steps {
		if snap.Steps[i].Position == exec.CurrentStep {
			currentStep = &snap.Steps[i]
			break
		}
	}
	if currentStep == nil {
		return nil, ErrInvalidStepPosition
	}

	// MN-15: corrupt step_outputs больше не silent — раньше unmarshal-error
	// возвращал outputs={} и затем перезаписывал StepOutputs новой записью,
	// тихо теряя outputs шагов 1..(N-1). Теперь возвращаем явную ошибку,
	// AdvanceStep отказывает — юзер видит сообщение, SRE может посмотреть
	// почему JSONB сломан.
	outputs := map[string]string{}
	if len(exec.StepOutputs) > 0 {
		if uErr := json.Unmarshal(exec.StepOutputs, &outputs); uErr != nil {
			slog.Error("chain.advance.corrupt_step_outputs",
				"execution_id", exec.ID, "error", uErr)
			return nil, fmt.Errorf("chain.advance: corrupt step_outputs: %w", uErr)
		}
	}
	outputs[stepOutputKey(currentStep.ID)] = stepOutput
	outputsJSON, err := json.Marshal(outputs)
	if err != nil {
		return nil, err
	}
	exec.StepOutputs = outputsJSON

	nextStep, err := resolveNextStep(snap.Steps, currentStep, chosenBranchIdx)
	if err != nil {
		return nil, err
	}

	if nextStep != nil {
		exec.CurrentStep = nextStep.Position
	} else {
		exec.Status = models.ChainExecutionStatusCompleted
		now := time.Now()
		exec.CompletedAt = &now
	}

	if err := s.chains.UpdateExecution(ctx, exec); err != nil {
		if errors.Is(err, repo.ErrConflict) {
			return nil, ErrConcurrentAdvance
		}
		return nil, err
	}
	if exec.Status == models.ChainExecutionStatusCompleted {
		metrics.ChainExecutionsCompleted.WithLabelValues("completed").Inc()
		slog.Info("chain.execution.completed",
			"chain_id", exec.ChainID,
			"exec_id", exec.ID,
			"user_id", userID,
			"steps_count", len(snap.Steps),
		)
		// Activity для team-цепочек: snap.Chain содержит team_id из момента старта
		// (snapshot замороженный). Если команда удалена — LogSafe всё равно
		// сделает попытку, но FK на team может отсутствовать → no-op в репо.
		if snap.Chain.TeamID != nil {
			s.activity.LogSafe(ctx, activityuc.Event{
				TeamID:      *snap.Chain.TeamID,
				ActorID:     userID,
				EventType:   models.ActivityChainExecutionCompleted,
				TargetType:  models.TargetChain,
				TargetID:    &snap.Chain.ID,
				TargetLabel: snap.Chain.Name,
			})
		}
	}
	return exec, nil
}

// resolveNextStep — переход по графу.
//
//	prompt-шаг: переход = currentStep.NextStepID (nil → конец).
//	fork-шаг:   переход = branches[chosenBranchIdx].NextStepID (nil → конец ветки).
func resolveNextStep(steps []models.PromptChainStep, currentStep *models.PromptChainStep, chosenBranchIdx *int) (*models.PromptChainStep, error) {
	if currentStep.StepType == models.StepTypeFork {
		if chosenBranchIdx == nil {
			return nil, ErrChooseBranchRequired
		}
		branch, err := findBranchByIndex(currentStep.Conditions, *chosenBranchIdx)
		if err != nil {
			return nil, err
		}
		if branch.NextStepID == nil {
			return nil, nil
		}
		for i := range steps {
			if steps[i].ID == *branch.NextStepID {
				return &steps[i], nil
			}
		}
		return nil, ErrInvalidNextStep
	}
	if currentStep.NextStepID == nil {
		return nil, nil
	}
	for i := range steps {
		if steps[i].ID == *currentStep.NextStepID {
			return &steps[i], nil
		}
	}
	return nil, ErrInvalidNextStep
}

// findRootStep — шаг, на который никто не указывает (ни как next_step_id, ни
// как fork branch). При обычной работе ровно один такой; при пустых ветках или
// частично собранных цепочках — берём ближайший к началу по position.
func findRootStep(steps []models.PromptChainStep) *models.PromptChainStep {
	if len(steps) == 0 {
		return nil
	}
	incoming := make(map[uint]struct{}, len(steps))
	for i := range steps {
		st := &steps[i]
		if st.NextStepID != nil {
			incoming[*st.NextStepID] = struct{}{}
		}
		if st.StepType == models.StepTypeFork && len(st.Conditions) > 0 {
			var conds models.Conditions
			if err := json.Unmarshal(st.Conditions, &conds); err == nil {
				for _, b := range conds.Branches {
					if b.NextStepID != nil {
						incoming[*b.NextStepID] = struct{}{}
					}
				}
			}
		}
	}
	var best *models.PromptChainStep
	for i := range steps {
		if _, isReferenced := incoming[steps[i].ID]; isReferenced {
			continue
		}
		if best == nil || steps[i].Position < best.Position {
			best = &steps[i]
		}
	}
	return best
}

// stepOutputKey — формат ключа в StepOutputs JSONB. step_<id>, чтобы переменные
// в шаблонах могли ссылаться на конкретный шаг по уникальному id.
func stepOutputKey(stepID uint) string {
	return "step_" + uintToString(stepID)
}

func uintToString(v uint) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

// --- Access helpers ---

func (s *Service) checkReadAccess(ctx context.Context, c *models.PromptChain, userID uint) error {
	if c.TeamID != nil {
		if _, err := s.teams.GetMember(ctx, *c.TeamID, userID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return ErrForbidden
			}
			return err
		}
		return nil
	}
	if c.UserID != userID {
		return ErrForbidden
	}
	return nil
}

func (s *Service) checkEditAccess(ctx context.Context, c *models.PromptChain, userID uint) error {
	return mapTeamError(teamcheck.RequireEditor(ctx, s.teams, c.TeamID, userID))
}

// isMaxTierForChain — fork-фича доступна, если:
//   - Personal-цепочка: юзер сам на тарифе Max/Max-Yearly
//   - Team-цепочка: хотя бы один owner команды на тарифе Max/Max-Yearly
//
// Это «справедливее» предыдущей реализации (где проверялся юзер-создатель).
// Pro-editor в Max-команде получает fork-фичу автоматически — owner «дарит»
// её всей команде через свой план.
func (s *Service) isMaxTierForChain(ctx context.Context, c *models.PromptChain, userID uint) bool {
	if s.quotas == nil {
		return false
	}
	if c.TeamID == nil {
		return s.quotas.IsMaxTierUser(ctx, userID)
	}
	members, err := s.teams.ListMembers(ctx, *c.TeamID)
	if err != nil {
		return false
	}
	for _, m := range members {
		if m.Role != models.RoleOwner {
			continue
		}
		if s.quotas.IsMaxTierUser(ctx, m.UserID) {
			return true
		}
	}
	return false
}

func mapTeamError(err error) error {
	return teamcheck.MapError(err, ErrForbidden, ErrViewerReadOnly)
}

// validateVariableMapping проверяет JSON корректность и whitelist sources.
// Не валидирует существование step_id/var_name — это ответственность UI/формы;
// при rendering неверный mapping приведёт к пустой подстановке.
func validateVariableMapping(raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]models.VariableSource
	if err := json.Unmarshal(raw, &m); err != nil {
		return ErrInvalidVariableMapping
	}
	for _, src := range m {
		switch src.Type {
		case models.VariableSourceManual, models.VariableSourceStepOutput, models.VariableSourceChainVar:
			// ok
		default:
			return ErrInvalidVariableMapping
		}
	}
	return nil
}
