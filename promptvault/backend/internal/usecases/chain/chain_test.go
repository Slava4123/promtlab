// Package chain — service-level unit-tests (CR-7 part 2).
//
// Покрываем:
//   - Create: tier-bounds (CheckChainQuota), team RBAC (viewer denied),
//     name/description validation
//   - Fork tier-gate: personal vs team (Max-owner-grants-team), Pro/Free blocked
//   - StartExecution: snapshot (chain + steps + prompt content),
//     empty chain refused, soft-deleted prompt fallback
//   - GetExecution: initiator-only, кикнут из команды между Start и Advance →
//     ErrForbidden (security regression W3)
//   - AdvanceStep: corrupt step_outputs → explicit error not silent (MN-15),
//     уже completed → refused
//   - Delete: HasActiveExecutions → refused
//
// Cycle detection / fork validation — pure functions, покрыты в conditions_test.go.
package chain

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	quotauc "promptvault/internal/usecases/quota"
)

// ----------------------- Helpers -----------------------

// testHarness — общий каркас для chain.Service тестов:
// собирает mock-репо + реальный quota.Service на mock-репо.
type testHarness struct {
	chains  *mockChainRepo
	prompts *mockPromptRepo
	teams   *mockTeamRepo
	plans   *mockPlanRepo
	quotas  *mockQuotaRepo
	users   *mockUserRepo
	svc     *Service
}

// newTestHarness — создаёт harness с пустыми моками.
// Activity передаём nil — LogSafe nil-safe (см. activity/service.go:79).
func newTestHarness() *testHarness {
	h := &testHarness{
		chains:  &mockChainRepo{},
		prompts: &mockPromptRepo{},
		teams:   &mockTeamRepo{},
		plans:   &mockPlanRepo{},
		quotas:  &mockQuotaRepo{},
		users:   &mockUserRepo{},
	}
	q := quotauc.NewService(h.plans, h.quotas, h.users)
	h.svc = NewService(h.chains, h.prompts, h.teams, q)
	return h
}

// newPlan — конструирует SubscriptionPlan с tier-лимитами для chain.
func newPlan(planID string, maxChains, maxStepsPerChain int) *models.SubscriptionPlan {
	return &models.SubscriptionPlan{
		ID:               planID,
		Name:             planID,
		MaxChains:        maxChains,
		MaxStepsPerChain: maxStepsPerChain,
		// Pack T: team-pool — для tests используем те же значения что personal
		// (тесты проверяют чисто bounds-логику, конкретные числа не важны).
		MaxTeamChains: maxChains,
	}
}

// makeTeam — тестовая команда. CreatedBy задаёт owner — это важно для Pack T
// (team-pool квота резолвит план owner'а через teams.GetByID).
func makeTeam(teamID, ownerID uint) *models.Team {
	return &models.Team{ID: teamID, CreatedBy: ownerID}
}

// newUser — User с указанным planID.
func newUser(id uint, planID string) *models.User {
	return &models.User{ID: id, PlanID: planID}
}

// expectQuotaCheck — настраивает моки plans+users+quotas для CheckChainQuota:
// users.GetByID(uid) → user with planID
// plans.GetByID(planID) → plan
// quotas.CountPersonalChains(uid) → currentCount (Pack T: только личные)
// quotas.CountTeamChains(*) → currentCount (для team-create flow)
// teams.GetByID(*) → team(creator=uid) — Pack T: chain.Service в team-mode
//   резолвит owner для team-pool квоты. Owner=uid в тестах для простоты.
func (h *testHarness) expectQuotaCheck(uid uint, planID string, plan *models.SubscriptionPlan, currentChains int64) {
	h.users.On("GetByID", mock.Anything, uid).Return(newUser(uid, planID), nil).Maybe()
	h.plans.On("GetByID", mock.Anything, planID).Return(plan, nil).Maybe()
	h.quotas.On("CountPersonalChains", mock.Anything, uid).Return(currentChains, nil).Maybe()
	h.quotas.On("CountTeamChains", mock.Anything, mock.Anything).Return(currentChains, nil).Maybe()
	h.teams.On("GetByID", mock.Anything, mock.AnythingOfType("uint")).
		Return(makeTeam(0, uid), nil).Maybe()
}

// expectIsMaxTier — настраивает users+plans для IsMaxTierUser.
// Использует Maybe() так как IsMaxTierUser может быть вызван несколько раз.
func (h *testHarness) expectIsMaxTier(uid uint, planID string) {
	h.users.On("GetByID", mock.Anything, uid).Return(newUser(uid, planID), nil).Maybe()
	h.plans.On("GetByID", mock.Anything, planID).Return(newPlan(planID, 100, 50), nil).Maybe()
}

// makeTeamMember — конструктор TeamMember с ролью.
func makeTeamMember(teamID, userID uint, role models.TeamRole) *models.TeamMember {
	return &models.TeamMember{
		ID:     userID,
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	}
}

// ----------------------- Create -----------------------

func TestCreate_FreeUserBlocked_AfterFirstChain(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)

	// Free план: MaxChains=1; уже создана 1.
	h.expectQuotaCheck(uid, "free", newPlan("free", 1, 3), 1)

	_, err := h.svc.Create(context.Background(), uid, "Test Chain", "", nil)
	if err == nil {
		t.Fatal("expected quota error, got nil")
	}
	var qe *quotauc.QuotaExceededError
	if !errors.As(err, &qe) {
		t.Fatalf("expected QuotaExceededError, got %T: %v", err, err)
	}
	if qe.QuotaType != "chains" {
		t.Errorf("expected QuotaType=chains, got %q", qe.QuotaType)
	}
}

func TestCreate_ProUser_OK(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)

	// Pro план: MaxChains=5; уже 2.
	h.expectQuotaCheck(uid, "pro", newPlan("pro", 5, 10), 2)
	h.chains.On("Create", mock.Anything, mock.AnythingOfType("*models.PromptChain")).
		Return(nil)

	c, err := h.svc.Create(context.Background(), uid, "My Chain", "desc", nil)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if c.UserID != uid {
		t.Errorf("expected UserID=%d, got %d", uid, c.UserID)
	}
	if c.Name != "My Chain" {
		t.Errorf("expected Name='My Chain', got %q", c.Name)
	}
	if c.TeamID != nil {
		t.Errorf("expected nil TeamID for personal chain, got %v", *c.TeamID)
	}
}

func TestCreate_TeamMode_ViewerDenied(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	teamID := uint(7)

	h.expectQuotaCheck(uid, "pro", newPlan("pro", 5, 10), 0)
	// Viewer в команде — RequireEditor должен отказать → ErrViewerReadOnly.
	h.teams.On("GetMember", mock.Anything, teamID, uid).
		Return(makeTeamMember(teamID, uid, models.RoleViewer), nil)

	_, err := h.svc.Create(context.Background(), uid, "Chain", "", &teamID)
	if !errors.Is(err, ErrViewerReadOnly) {
		t.Fatalf("expected ErrViewerReadOnly, got %v", err)
	}
}

func TestCreate_TeamMode_EditorOK(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	teamID := uint(7)

	h.expectQuotaCheck(uid, "pro", newPlan("pro", 5, 10), 0)
	h.teams.On("GetMember", mock.Anything, teamID, uid).
		Return(makeTeamMember(teamID, uid, models.RoleEditor), nil)
	h.chains.On("Create", mock.Anything, mock.AnythingOfType("*models.PromptChain")).
		Return(nil)

	c, err := h.svc.Create(context.Background(), uid, "Chain", "", &teamID)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if c.TeamID == nil || *c.TeamID != teamID {
		t.Errorf("expected TeamID=%d, got %v", teamID, c.TeamID)
	}
}

func TestCreate_NonMember_Forbidden(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	teamID := uint(7)

	h.expectQuotaCheck(uid, "pro", newPlan("pro", 5, 10), 0)
	// repo.ErrNotFound из GetMember — юзер не в команде → ErrForbidden.
	h.teams.On("GetMember", mock.Anything, teamID, uid).
		Return(nil, repo.ErrNotFound)

	_, err := h.svc.Create(context.Background(), uid, "Chain", "", &teamID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCreate_EmptyName_Refused(t *testing.T) {
	h := newTestHarness()
	_, err := h.svc.Create(context.Background(), 1, "  ", "", nil)
	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestCreate_TooLongDescription_Refused(t *testing.T) {
	h := newTestHarness()
	long := strings.Repeat("x", maxDescriptionLen+1)
	_, err := h.svc.Create(context.Background(), 1, "Name", long, nil)
	if !errors.Is(err, ErrInvalidDescription) {
		t.Fatalf("expected ErrInvalidDescription, got %v", err)
	}
}

// ----------------------- Fork tier-gate (AddStep) -----------------------

// TestForkTierGate_PersonalChain_FreeUser_Blocked — fork-шаг в personal-цепочке
// требует Max-tier у владельца.
//
// Порядок проверок в AddStep (chain.go:240-292):
//  1. GetByID chain
//  2. checkEditAccess (для personal teamID=nil → passes)
//  3. CheckChainStepQuota (нужны users+plans+quotas mocks)
//  4. validateVariableMapping
//  5. isMaxTierForChain → free → false → ErrForkRequiresMax
func TestForkTierGate_PersonalChain_FreeUser_Blocked(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)

	personalChain := &models.PromptChain{ID: chainID, UserID: uid, TeamID: nil, Name: "Test"}
	h.chains.On("GetByID", mock.Anything, chainID).Return(personalChain, nil)
	// CheckChainStepQuota: free plan MaxStepsPerChain=3, current=0 → проходит.
	h.users.On("GetByID", mock.Anything, uid).Return(newUser(uid, "free"), nil).Maybe()
	h.plans.On("GetByID", mock.Anything, "free").Return(newPlan("free", 1, 3), nil).Maybe()
	h.quotas.On("CountStepsByChain", mock.Anything, chainID).Return(int64(0), nil).Maybe()
	// IsMaxTierUser → free → false → ErrForkRequiresMax.

	conds, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{{Label: "ok", NextStepID: nil}},
	})
	_, err := h.svc.AddStep(context.Background(), chainID, uid, AddStepInput{
		StepType:   models.StepTypeFork,
		Name:       "fork",
		Conditions: conds,
	})
	if !errors.Is(err, ErrForkRequiresMax) {
		t.Fatalf("expected ErrForkRequiresMax, got %v", err)
	}
}

// TestForkTierGate_PersonalChain_MaxUser_Allowed — Max-tier юзер может добавить fork.
func TestForkTierGate_PersonalChain_MaxUser_Allowed(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)

	personalChain := &models.PromptChain{ID: chainID, UserID: uid, TeamID: nil, Name: "Test"}
	h.chains.On("GetByID", mock.Anything, chainID).Return(personalChain, nil)
	h.expectIsMaxTier(uid, "max")
	// CheckChainStepQuota: используем тот же план (max, MaxStepsPerChain=50).
	h.quotas.On("CountStepsByChain", mock.Anything, chainID).Return(int64(0), nil).Maybe()
	// ListStepsByChain — пустой граф.
	h.chains.On("ListStepsByChain", mock.Anything, chainID).Return([]models.PromptChainStep{}, nil)
	h.chains.On("CountStepsByChain", mock.Anything, chainID).Return(int64(0), nil)
	h.chains.On("AddStep", mock.Anything, mock.AnythingOfType("*models.PromptChainStep")).Return(nil)

	conds, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{{Label: "done", NextStepID: nil}},
	})
	step, err := h.svc.AddStep(context.Background(), chainID, uid, AddStepInput{
		StepType:   models.StepTypeFork,
		Name:       "fork",
		Conditions: conds,
	})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if step.StepType != models.StepTypeFork {
		t.Errorf("expected StepType=fork, got %q", step.StepType)
	}
	if step.PromptID != nil {
		t.Errorf("fork step must have nil PromptID, got %v", *step.PromptID)
	}
}

// TestForkTierGate_TeamWithMaxOwner_AllowsProEditor — fork доступен Pro-editor'у,
// если хотя бы один owner команды на тарифе Max ("owner дарит фичу команде").
func TestForkTierGate_TeamWithMaxOwner_AllowsProEditor(t *testing.T) {
	h := newTestHarness()
	editorID := uint(42)
	ownerID := uint(99)
	teamID := uint(7)
	chainID := uint(10)

	teamChain := &models.PromptChain{ID: chainID, UserID: ownerID, TeamID: &teamID, Name: "Team Chain"}
	h.chains.On("GetByID", mock.Anything, chainID).Return(teamChain, nil)
	// Editor → checkEditAccess passes.
	h.teams.On("GetMember", mock.Anything, teamID, editorID).
		Return(makeTeamMember(teamID, editorID, models.RoleEditor), nil)

	// Команда: editor (pro) + owner (max). isMaxTierForChain должен найти Max-owner.
	h.teams.On("ListMembers", mock.Anything, teamID).Return([]models.TeamMember{
		*makeTeamMember(teamID, editorID, models.RoleEditor),
		*makeTeamMember(teamID, ownerID, models.RoleOwner),
	}, nil)
	// IsMaxTierUser(editor) — pro → false; (owner) — max → true.
	h.users.On("GetByID", mock.Anything, editorID).Return(newUser(editorID, "pro"), nil).Maybe()
	h.users.On("GetByID", mock.Anything, ownerID).Return(newUser(ownerID, "max"), nil).Maybe()
	h.plans.On("GetByID", mock.Anything, "pro").Return(newPlan("pro", 5, 10), nil).Maybe()
	h.plans.On("GetByID", mock.Anything, "max").Return(newPlan("max", 100, 50), nil).Maybe()

	// CheckChainStepQuota → editor's pro plan, MaxStepsPerChain=10, current=0.
	h.quotas.On("CountStepsByChain", mock.Anything, chainID).Return(int64(0), nil).Maybe()
	h.chains.On("ListStepsByChain", mock.Anything, chainID).Return([]models.PromptChainStep{}, nil)
	h.chains.On("CountStepsByChain", mock.Anything, chainID).Return(int64(0), nil)
	h.chains.On("AddStep", mock.Anything, mock.AnythingOfType("*models.PromptChainStep")).Return(nil)

	conds, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{{Label: "done", NextStepID: nil}},
	})
	_, err := h.svc.AddStep(context.Background(), chainID, editorID, AddStepInput{
		StepType:   models.StepTypeFork,
		Name:       "fork-by-pro-editor",
		Conditions: conds,
	})
	if err != nil {
		t.Fatalf("expected nil err for Pro-editor in Max-team, got %v", err)
	}
}

// TestForkTierGate_TeamNoMaxOwner_BlocksAll — нет Max-owner'а → fork запрещён всем.
func TestForkTierGate_TeamNoMaxOwner_BlocksAll(t *testing.T) {
	h := newTestHarness()
	editorID := uint(42)
	ownerID := uint(99)
	teamID := uint(7)
	chainID := uint(10)

	teamChain := &models.PromptChain{ID: chainID, UserID: ownerID, TeamID: &teamID}
	h.chains.On("GetByID", mock.Anything, chainID).Return(teamChain, nil)
	h.teams.On("GetMember", mock.Anything, teamID, editorID).
		Return(makeTeamMember(teamID, editorID, models.RoleEditor), nil)

	// Owner на pro (не max) — fork недоступен всей команде.
	h.teams.On("ListMembers", mock.Anything, teamID).Return([]models.TeamMember{
		*makeTeamMember(teamID, ownerID, models.RoleOwner),
	}, nil)
	// CheckChainStepQuota для editor (pro plan).
	h.users.On("GetByID", mock.Anything, editorID).Return(newUser(editorID, "pro"), nil).Maybe()
	h.users.On("GetByID", mock.Anything, ownerID).Return(newUser(ownerID, "pro"), nil).Maybe()
	h.plans.On("GetByID", mock.Anything, "pro").Return(newPlan("pro", 5, 10), nil).Maybe()
	h.quotas.On("CountStepsByChain", mock.Anything, chainID).Return(int64(0), nil).Maybe()

	conds, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{{Label: "done", NextStepID: nil}},
	})
	_, err := h.svc.AddStep(context.Background(), chainID, editorID, AddStepInput{
		StepType:   models.StepTypeFork,
		Name:       "fork",
		Conditions: conds,
	})
	if !errors.Is(err, ErrForkRequiresMax) {
		t.Fatalf("expected ErrForkRequiresMax, got %v", err)
	}
}

// ----------------------- StartExecution -----------------------

func TestStartExecution_SnapshotsCurrentChainState(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)
	promptID := uint(100)
	stepID := uint(1)

	// Цепочка с 1 шагом, Prompt уже preloaded в snapshot.
	chainWithSteps := &models.PromptChain{
		ID:     chainID,
		UserID: uid,
		Name:   "Test Chain",
		Steps: []models.PromptChainStep{
			{
				ID:       stepID,
				ChainID:  chainID,
				Position: 1,
				StepType: models.StepTypePrompt,
				PromptID: &promptID,
				Prompt: &models.Prompt{
					ID:      promptID,
					Content: "Hello {{var}}",
				},
			},
		},
	}
	h.chains.On("GetByIDWithSteps", mock.Anything, chainID).Return(chainWithSteps, nil)
	h.chains.On("CreateExecution", mock.Anything, mock.AnythingOfType("*models.PromptChainExecution")).
		Return(nil)

	exec, err := h.svc.StartExecution(context.Background(), chainID, uid, nil)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if exec.Status != models.ChainExecutionStatusInProgress {
		t.Errorf("expected status=in_progress, got %q", exec.Status)
	}
	// Verify snapshot.
	var snap models.ChainSnapshot
	if err := json.Unmarshal(exec.ChainSnapshot, &snap); err != nil {
		t.Fatalf("snapshot must unmarshal, got %v", err)
	}
	if snap.Chain.ID != chainID {
		t.Errorf("snapshot.Chain.ID=%d, want %d", snap.Chain.ID, chainID)
	}
	if len(snap.Steps) != 1 {
		t.Fatalf("snapshot.Steps len=%d, want 1", len(snap.Steps))
	}
	if got := snap.PromptContents[promptID]; got != "Hello {{var}}" {
		t.Errorf("snapshot.PromptContents[%d]=%q, want %q", promptID, got, "Hello {{var}}")
	}
}

func TestStartExecution_EmptyChain_Refused(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)

	emptyChain := &models.PromptChain{ID: chainID, UserID: uid, Steps: nil}
	h.chains.On("GetByIDWithSteps", mock.Anything, chainID).Return(emptyChain, nil)

	_, err := h.svc.StartExecution(context.Background(), chainID, uid, nil)
	if !errors.Is(err, ErrEmptyChain) {
		t.Fatalf("expected ErrEmptyChain, got %v", err)
	}
}

// TestStartExecution_PromptSoftDeleted_Fallback — prompt не preloaded (soft-deleted),
// fallback на prompts.GetByID должен вернуть ErrPromptNotFound.
func TestStartExecution_PromptSoftDeleted_Fallback(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)
	promptID := uint(100)

	// Step.Prompt == nil → fallback на GetByID, который вернёт ErrNotFound.
	chainWithSteps := &models.PromptChain{
		ID:     chainID,
		UserID: uid,
		Steps: []models.PromptChainStep{
			{
				ID:       1,
				ChainID:  chainID,
				Position: 1,
				StepType: models.StepTypePrompt,
				PromptID: &promptID,
				Prompt:   nil, // не preloaded
			},
		},
	}
	h.chains.On("GetByIDWithSteps", mock.Anything, chainID).Return(chainWithSteps, nil)
	h.prompts.On("GetByID", mock.Anything, promptID).Return(nil, repo.ErrNotFound)

	_, err := h.svc.StartExecution(context.Background(), chainID, uid, nil)
	if !errors.Is(err, ErrPromptNotFound) {
		t.Fatalf("expected ErrPromptNotFound on soft-deleted prompt, got %v", err)
	}
}

// ----------------------- GetExecution / Initiator-only / RBAC W3 -----------------------

func TestGetExecution_NotInitiator_Forbidden(t *testing.T) {
	h := newTestHarness()
	initiatorID := uint(42)
	otherID := uint(99)
	execID := uint(50)

	exec := &models.PromptChainExecution{ID: execID, UserID: initiatorID, ChainID: 10}
	h.chains.On("GetExecutionByID", mock.Anything, execID).Return(exec, nil)

	_, err := h.svc.GetExecution(context.Background(), execID, otherID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for non-initiator, got %v", err)
	}
}

// TestGetExecution_UserKickedFromTeam_Forbidden — security regression W3:
// initiator запустил team-chain execution, потом был выгнан из команды между
// Start и Advance — checkReadAccess должен вернуть ErrForbidden.
func TestGetExecution_UserKickedFromTeam_Forbidden(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	teamID := uint(7)
	chainID := uint(10)
	execID := uint(50)

	exec := &models.PromptChainExecution{ID: execID, UserID: uid, ChainID: chainID}
	teamChain := &models.PromptChain{ID: chainID, UserID: 99, TeamID: &teamID}
	h.chains.On("GetExecutionByID", mock.Anything, execID).Return(exec, nil)
	h.chains.On("GetByID", mock.Anything, chainID).Return(teamChain, nil)
	// Юзер больше не в команде.
	h.teams.On("GetMember", mock.Anything, teamID, uid).Return(nil, repo.ErrNotFound)

	_, err := h.svc.GetExecution(context.Background(), execID, uid)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden after kick from team, got %v", err)
	}
}

// TestGetExecution_ChainSoftDeleted_StillReturnsExec — chain удалён после Start,
// но execution ещё in-progress; initiator должен мочь продолжить (snapshot живой).
func TestGetExecution_ChainSoftDeleted_StillReturnsExec(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)
	execID := uint(50)

	exec := &models.PromptChainExecution{ID: execID, UserID: uid, ChainID: chainID}
	h.chains.On("GetExecutionByID", mock.Anything, execID).Return(exec, nil)
	h.chains.On("GetByID", mock.Anything, chainID).Return(nil, repo.ErrNotFound)

	got, err := h.svc.GetExecution(context.Background(), execID, uid)
	if err != nil {
		t.Fatalf("expected nil err (snapshot keeps execution alive), got %v", err)
	}
	if got.ID != execID {
		t.Errorf("expected exec.ID=%d, got %d", execID, got.ID)
	}
}

// ----------------------- AdvanceStep -----------------------

// TestAdvanceStep_CorruptStepOutputs_RefusedNotSilent — MN-15: corrupt JSONB
// больше не игнорируется silent. Раньше unmarshal-error → outputs={}, перезапись
// тихо теряла outputs шагов 1..(N-1). Теперь — explicit error.
func TestAdvanceStep_CorruptStepOutputs_RefusedNotSilent(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)
	execID := uint(50)
	stepID := uint(1)
	promptID := uint(100)

	// Snapshot с одним шагом, но StepOutputs содержит invalid JSON.
	snap := models.ChainSnapshot{
		Chain: models.PromptChain{ID: chainID, UserID: uid},
		Steps: []models.PromptChainStep{
			{ID: stepID, ChainID: chainID, Position: 1, StepType: models.StepTypePrompt, PromptID: &promptID},
		},
	}
	snapJSON, _ := json.Marshal(snap)
	exec := &models.PromptChainExecution{
		ID:            execID,
		UserID:        uid,
		ChainID:       chainID,
		CurrentStep:   1,
		Status:        models.ChainExecutionStatusInProgress,
		ChainSnapshot: snapJSON,
		StepOutputs:   json.RawMessage(`{not-json`), // corrupt
	}
	h.chains.On("GetExecutionByID", mock.Anything, execID).Return(exec, nil)
	h.chains.On("GetByID", mock.Anything, chainID).
		Return(&models.PromptChain{ID: chainID, UserID: uid}, nil)

	_, err := h.svc.AdvanceStep(context.Background(), execID, uid, "output text", nil)
	if err == nil {
		t.Fatal("expected error on corrupt step_outputs, got nil")
	}
	if !strings.Contains(err.Error(), "corrupt step_outputs") {
		t.Errorf("expected 'corrupt step_outputs' in error, got %v", err)
	}
}

func TestAdvanceStep_AlreadyCompleted_Refused(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)
	execID := uint(50)

	exec := &models.PromptChainExecution{
		ID:      execID,
		UserID:  uid,
		ChainID: chainID,
		Status:  models.ChainExecutionStatusCompleted,
	}
	h.chains.On("GetExecutionByID", mock.Anything, execID).Return(exec, nil)
	h.chains.On("GetByID", mock.Anything, chainID).
		Return(&models.PromptChain{ID: chainID, UserID: uid}, nil)

	_, err := h.svc.AdvanceStep(context.Background(), execID, uid, "out", nil)
	if !errors.Is(err, ErrExecutionAlreadyCompleted) {
		t.Fatalf("expected ErrExecutionAlreadyCompleted, got %v", err)
	}
}

// TestAdvanceStep_PromptStep_HappyPath — линейная цепочка из 2 шагов;
// первый Advance переводит на step 2, второй — завершает execution.
func TestAdvanceStep_PromptStep_HappyPath(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)
	execID := uint(50)
	pid := uint(100)
	step1ID := uint(1)
	step2ID := uint(2)

	// step1.next = step2; step2.next = nil (конец).
	snap := models.ChainSnapshot{
		Chain: models.PromptChain{ID: chainID, UserID: uid},
		Steps: []models.PromptChainStep{
			{ID: step1ID, Position: 1, StepType: models.StepTypePrompt, PromptID: &pid, NextStepID: &step2ID},
			{ID: step2ID, Position: 2, StepType: models.StepTypePrompt, PromptID: &pid, NextStepID: nil},
		},
	}
	snapJSON, _ := json.Marshal(snap)
	exec := &models.PromptChainExecution{
		ID:            execID,
		UserID:        uid,
		ChainID:       chainID,
		CurrentStep:   1,
		Status:        models.ChainExecutionStatusInProgress,
		ChainSnapshot: snapJSON,
		StepOutputs:   json.RawMessage(`{}`),
	}
	h.chains.On("GetExecutionByID", mock.Anything, execID).Return(exec, nil)
	h.chains.On("GetByID", mock.Anything, chainID).
		Return(&models.PromptChain{ID: chainID, UserID: uid}, nil)
	h.chains.On("UpdateExecution", mock.Anything, mock.AnythingOfType("*models.PromptChainExecution")).
		Return(nil)

	updated, err := h.svc.AdvanceStep(context.Background(), execID, uid, "step1 result", nil)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if updated.CurrentStep != 2 {
		t.Errorf("expected CurrentStep=2 after first advance, got %d", updated.CurrentStep)
	}
	if updated.Status != models.ChainExecutionStatusInProgress {
		t.Errorf("expected status=in_progress after first advance, got %q", updated.Status)
	}
}

func TestAdvanceStep_ForkStep_RequiresChosenBranch(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)
	execID := uint(50)
	forkID := uint(1)

	conds, _ := json.Marshal(models.Conditions{
		Branches: []models.ConditionBranch{
			{Label: "left", NextStepID: nil},
			{Label: "right", NextStepID: nil},
		},
	})
	snap := models.ChainSnapshot{
		Chain: models.PromptChain{ID: chainID, UserID: uid},
		Steps: []models.PromptChainStep{
			{ID: forkID, Position: 1, StepType: models.StepTypeFork, Conditions: conds},
		},
	}
	snapJSON, _ := json.Marshal(snap)
	exec := &models.PromptChainExecution{
		ID:            execID,
		UserID:        uid,
		ChainID:       chainID,
		CurrentStep:   1,
		Status:        models.ChainExecutionStatusInProgress,
		ChainSnapshot: snapJSON,
		StepOutputs:   json.RawMessage(`{}`),
	}
	h.chains.On("GetExecutionByID", mock.Anything, execID).Return(exec, nil)
	h.chains.On("GetByID", mock.Anything, chainID).
		Return(&models.PromptChain{ID: chainID, UserID: uid}, nil)

	// chosenBranchIdx=nil → ErrChooseBranchRequired.
	_, err := h.svc.AdvanceStep(context.Background(), execID, uid, "out", nil)
	if !errors.Is(err, ErrChooseBranchRequired) {
		t.Fatalf("expected ErrChooseBranchRequired on nil branch idx, got %v", err)
	}
}

// ----------------------- Delete (HasActiveExecutions guard) -----------------------

func TestDelete_HasActiveExecutions_Refused(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)

	personalChain := &models.PromptChain{ID: chainID, UserID: uid}
	h.chains.On("GetByID", mock.Anything, chainID).Return(personalChain, nil)
	h.chains.On("HasActiveExecutions", mock.Anything, chainID).Return(true, nil)

	err := h.svc.Delete(context.Background(), chainID, uid)
	if !errors.Is(err, ErrChainHasActiveExecutions) {
		t.Fatalf("expected ErrChainHasActiveExecutions, got %v", err)
	}
}

func TestDelete_NoActiveExecutions_OK(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	chainID := uint(10)

	personalChain := &models.PromptChain{ID: chainID, UserID: uid}
	h.chains.On("GetByID", mock.Anything, chainID).Return(personalChain, nil)
	h.chains.On("HasActiveExecutions", mock.Anything, chainID).Return(false, nil)
	h.chains.On("SoftDelete", mock.Anything, chainID).Return(nil)

	if err := h.svc.Delete(context.Background(), chainID, uid); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// TestDelete_NotOwner_Personal_Forbidden — для personal-цепочки доступ только у owner'а.
func TestDelete_NotOwner_Personal_Forbidden(t *testing.T) {
	h := newTestHarness()
	uid := uint(42)
	otherID := uint(99)
	chainID := uint(10)

	personalChain := &models.PromptChain{ID: chainID, UserID: otherID}
	h.chains.On("GetByID", mock.Anything, chainID).Return(personalChain, nil)

	err := h.svc.Delete(context.Background(), chainID, uid)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for non-owner, got %v", err)
	}
}
