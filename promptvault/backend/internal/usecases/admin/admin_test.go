package admin

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	auditsvc "promptvault/internal/usecases/audit"
	badgeuc "promptvault/internal/usecases/badge"
)

// ========== Fakes ==========

type fakeUserRepo struct {
	users  map[uint]*models.User
	update func(u *models.User) error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[uint]*models.User)}
}

func (f *fakeUserRepo) Create(_ context.Context, user *models.User) error {
	f.users[user.ID] = user
	return nil
}
func (f *fakeUserRepo) GetByID(_ context.Context, id uint) (*models.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return u, nil
}
func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (*models.User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, repo.ErrNotFound
}
func (f *fakeUserRepo) GetByUsername(_ context.Context, username string) (*models.User, error) {
	for _, u := range f.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, repo.ErrNotFound
}
func (f *fakeUserRepo) SearchUsers(_ context.Context, _ string, _ int) ([]models.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) Update(_ context.Context, user *models.User) error {
	if f.update != nil {
		return f.update(user)
	}
	f.users[user.ID] = user
	return nil
}
func (f *fakeUserRepo) SetPlan(_ context.Context, userID uint, planID string) error {
	if u, ok := f.users[userID]; ok {
		u.PlanID = planID
		return nil
	}
	return repo.ErrNotFound
}
func (f *fakeUserRepo) SetQuotaWarningSentOn(_ context.Context, _ uint, _ time.Time) error {
	return nil
}
func (f *fakeUserRepo) TouchLastLogin(_ context.Context, _ uint) error {
	return nil
}
func (f *fakeUserRepo) ListInactiveForReengagement(_ context.Context, _, _ time.Time, _ int) ([]models.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) MarkReengagementSent(_ context.Context, _ uint) error {
	return nil
}
func (f *fakeUserRepo) CountReferredBy(_ context.Context, _ string) (int64, error) { return 0, nil }
func (f *fakeUserRepo) GetByReferralCode(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) MarkReferralRewarded(_ context.Context, _ uint) (bool, error) {
	return false, nil
}
func (f *fakeUserRepo) ListPaidUsers(_ context.Context) ([]uint, error) { return nil, nil }
func (f *fakeUserRepo) SetInsightEmailsEnabled(_ context.Context, _ uint, _ bool) error {
	return nil
}

type fakeAdminRepo struct {
	users          []repo.UserSummary
	detailByID     map[uint]*repo.UserDetail
	updatedStatus  map[uint]models.UserStatus
	updateStatusErr error
}

func newFakeAdminRepo() *fakeAdminRepo {
	return &fakeAdminRepo{
		detailByID:    make(map[uint]*repo.UserDetail),
		updatedStatus: make(map[uint]models.UserStatus),
	}
}

func (f *fakeAdminRepo) ListUsers(_ context.Context, _ repo.UserListFilter) ([]repo.UserSummary, int64, error) {
	return f.users, int64(len(f.users)), nil
}
func (f *fakeAdminRepo) GetUserDetail(_ context.Context, userID uint) (*repo.UserDetail, error) {
	d, ok := f.detailByID[userID]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return d, nil
}
func (f *fakeAdminRepo) UpdateStatus(_ context.Context, userID uint, status models.UserStatus) error {
	if f.updateStatusErr != nil {
		return f.updateStatusErr
	}
	f.updatedStatus[userID] = status
	return nil
}
func (f *fakeAdminRepo) CountUsers(_ context.Context) (total, admins, active, frozen int64, err error) {
	return 0, 0, 0, 0, nil
}

type fakeAuditRepo struct {
	entries []models.AuditLog
}

func (f *fakeAuditRepo) Log(_ context.Context, entry *models.AuditLog) error {
	entry.ID = uint(len(f.entries) + 1)
	f.entries = append(f.entries, *entry)
	return nil
}
func (f *fakeAuditRepo) List(_ context.Context, _ repo.AuditLogFilter) ([]models.AuditLog, int64, error) {
	return f.entries, int64(len(f.entries)), nil
}

type fakeBadgeRepo struct {
	unlocked         map[uint]map[string]time.Time
	versionedPrompts map[int]int64
}

func newFakeBadgeRepo() *fakeBadgeRepo {
	return &fakeBadgeRepo{
		unlocked:         make(map[uint]map[string]time.Time),
		versionedPrompts: make(map[int]int64),
	}
}

func (f *fakeBadgeRepo) Unlock(_ context.Context, userID uint, badgeID string) error {
	if f.unlocked[userID] == nil {
		f.unlocked[userID] = make(map[string]time.Time)
	}
	if _, ok := f.unlocked[userID][badgeID]; ok {
		return repo.ErrBadgeAlreadyUnlocked
	}
	f.unlocked[userID][badgeID] = time.Now()
	return nil
}
func (f *fakeBadgeRepo) UnlockedIDs(_ context.Context, userID uint) (map[string]struct{}, error) {
	set := make(map[string]struct{})
	for id := range f.unlocked[userID] {
		set[id] = struct{}{}
	}
	return set, nil
}
func (f *fakeBadgeRepo) ListByUser(_ context.Context, userID uint) ([]models.UserBadge, error) {
	var out []models.UserBadge
	for id, t := range f.unlocked[userID] {
		out = append(out, models.UserBadge{UserID: userID, BadgeID: id, UnlockedAt: t})
	}
	return out, nil
}
func (f *fakeBadgeRepo) DeleteByUserAndBadge(_ context.Context, userID uint, badgeID string) error {
	delete(f.unlocked[userID], badgeID)
	return nil
}
func (f *fakeBadgeRepo) CountSoloPrompts(_ context.Context, _ uint) (int64, error)    { return 0, nil }
func (f *fakeBadgeRepo) CountTeamPrompts(_ context.Context, _ uint) (int64, error)    { return 0, nil }
func (f *fakeBadgeRepo) CountAllPrompts(_ context.Context, _ uint) (int64, error)     { return 0, nil }
func (f *fakeBadgeRepo) CountSoloCollections(_ context.Context, _ uint) (int64, error) {
	return 0, nil
}
func (f *fakeBadgeRepo) CountTeamCollections(_ context.Context, _ uint) (int64, error) {
	return 0, nil
}
func (f *fakeBadgeRepo) SumUsage(_ context.Context, _ uint) (int64, error) { return 0, nil }
func (f *fakeBadgeRepo) CountVersionedPrompts(_ context.Context, _ uint, minV int) (int64, error) {
	return f.versionedPrompts[minV], nil
}

// ========== Plan / Subscription / Notifier fakes ==========

type fakePlansRepo struct {
	plans map[string]*models.SubscriptionPlan
}

func newFakePlansRepo(ids ...string) *fakePlansRepo {
	m := make(map[string]*models.SubscriptionPlan, len(ids))
	for _, id := range ids {
		m[id] = &models.SubscriptionPlan{ID: id, Name: id, IsActive: true}
	}
	return &fakePlansRepo{plans: m}
}

func (f *fakePlansRepo) GetAll(_ context.Context) ([]models.SubscriptionPlan, error) {
	out := make([]models.SubscriptionPlan, 0, len(f.plans))
	for _, p := range f.plans {
		out = append(out, *p)
	}
	return out, nil
}
func (f *fakePlansRepo) GetByID(_ context.Context, id string) (*models.SubscriptionPlan, error) {
	p, ok := f.plans[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return p, nil
}
func (f *fakePlansRepo) GetActive(_ context.Context) ([]models.SubscriptionPlan, error) {
	return f.GetAll(context.Background())
}

// fakeSubsRepo — минимальный fake для SubscriptionRepository, реализует только
// методы, нужные admin.ChangeTier (GetCurrentByUserID, MarkExpired). Остальные
// методы паникуют — тесты admin не должны их трогать.
type fakeSubsRepo struct {
	subsByUser   map[uint]*models.Subscription
	markedExpired []uint
	getCurrentErr error
	markExpiredErr error
}

func newFakeSubsRepo() *fakeSubsRepo {
	return &fakeSubsRepo{subsByUser: make(map[uint]*models.Subscription)}
}

func (f *fakeSubsRepo) GetCurrentByUserID(_ context.Context, userID uint) (*models.Subscription, error) {
	if f.getCurrentErr != nil {
		return nil, f.getCurrentErr
	}
	s, ok := f.subsByUser[userID]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return s, nil
}
func (f *fakeSubsRepo) MarkExpired(_ context.Context, subID uint) error {
	if f.markExpiredErr != nil {
		return f.markExpiredErr
	}
	f.markedExpired = append(f.markedExpired, subID)
	return nil
}

// not used in admin tests — paniciem чтобы нечаянно не вызвали
func (f *fakeSubsRepo) Create(context.Context, *models.Subscription) error { panic("unused") }
func (f *fakeSubsRepo) GetActiveByUserID(context.Context, uint) (*models.Subscription, error) {
	panic("unused")
}
func (f *fakeSubsRepo) Update(context.Context, *models.Subscription) error { panic("unused") }
func (f *fakeSubsRepo) ListExpiring(context.Context, time.Time) ([]models.Subscription, error) {
	panic("unused")
}
func (f *fakeSubsRepo) ActivateWithPlanUpdate(context.Context, *models.Subscription, uint, string) error {
	panic("unused")
}
func (f *fakeSubsRepo) CancelAtPeriodEnd(context.Context, uint) error           { panic("unused") }
func (f *fakeSubsRepo) ExpireAndDowngrade(context.Context, uint, uint) error    { panic("unused") }
func (f *fakeSubsRepo) SetRebillId(context.Context, uint, string) error         { panic("unused") }
func (f *fakeSubsRepo) SetAutoRenew(context.Context, uint, bool) error          { panic("unused") }
func (f *fakeSubsRepo) ListReadyForRenewal(context.Context, time.Time, time.Time, int) ([]models.Subscription, error) {
	panic("unused")
}
func (f *fakeSubsRepo) ExtendPeriod(context.Context, uint, time.Time) error    { panic("unused") }
func (f *fakeSubsRepo) UpdatePeriodEnd(context.Context, uint, time.Time) error { panic("unused") }
func (f *fakeSubsRepo) RecordRenewalFailure(context.Context, uint) error    { panic("unused") }
func (f *fakeSubsRepo) ListPreExpiring(context.Context, time.Time, time.Time, int16) ([]models.Subscription, error) {
	panic("unused")
}
func (f *fakeSubsRepo) SetPreExpireStage(context.Context, uint, int16) error { panic("unused") }
func (f *fakeSubsRepo) Pause(context.Context, uint, uint, time.Time, time.Time) error {
	panic("unused")
}
func (f *fakeSubsRepo) Resume(context.Context, uint, uint, time.Time, time.Time) error {
	panic("unused")
}
func (f *fakeSubsRepo) ListExpiredPauses(context.Context, time.Time) ([]models.Subscription, error) {
	panic("unused")
}
func (f *fakeSubsRepo) RecordCancellation(context.Context, *models.SubscriptionCancellation) error {
	panic("unused")
}

type tierEmailCall struct {
	to, name, oldPlan, newPlan, reason, frontendURL string
}

type fakeNotifier struct {
	calls   []tierEmailCall
	sendErr error
}

func (f *fakeNotifier) SendAdminTierChanged(to, name, oldPlan, newPlan, reason, frontendURL string) error {
	f.calls = append(f.calls, tierEmailCall{to, name, oldPlan, newPlan, reason, frontendURL})
	return f.sendErr
}

// ========== Test helpers ==========

type fixture struct {
	svc        *Service
	users      *fakeUserRepo
	adminRepo  *fakeAdminRepo
	auditRepo  *fakeAuditRepo
	badgeRepo  *fakeBadgeRepo
	badgeSvc   *badgeuc.Service
	plansRepo  *fakePlansRepo
	subsRepo   *fakeSubsRepo
	notifier   *fakeNotifier
	adminUser  *models.User
	targetUser *models.User
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	users := newFakeUserRepo()
	adminUser := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin, Status: models.StatusActive}
	target := &models.User{ID: 2, Email: "user@example.com", Name: "Тестовый Пользователь", Role: models.RoleUser, Status: models.StatusActive, PlanID: "free"}
	users.users[1] = adminUser
	users.users[2] = target

	adminRepo := newFakeAdminRepo()
	adminRepo.detailByID[2] = &repo.UserDetail{User: target, PromptCount: 10}

	auditRepo := &fakeAuditRepo{}
	auditSvc := auditsvc.NewService(auditRepo)

	badgeRepo := newFakeBadgeRepo()
	badgeSvc, err := badgeuc.NewService(badgeRepo, nil)
	require.NoError(t, err)

	plansRepo := newFakePlansRepo("free", "pro", "max")
	subsRepo := newFakeSubsRepo()
	notifier := &fakeNotifier{}

	svc := NewService(adminRepo, users, auditSvc, nil, badgeSvc, plansRepo, subsRepo)
	svc.SetTierChangeNotifier(notifier, "https://promtlabs.ru")

	return &fixture{
		svc:        svc,
		users:      users,
		adminRepo:  adminRepo,
		auditRepo:  auditRepo,
		badgeRepo:  badgeRepo,
		badgeSvc:   badgeSvc,
		plansRepo:  plansRepo,
		subsRepo:   subsRepo,
		notifier:   notifier,
		adminUser:  adminUser,
		targetUser: target,
	}
}

func ctxWithAdmin(adminID uint) context.Context {
	return auditsvc.WithContext(context.Background(), auditsvc.AdminRequestInfo{
		AdminID: adminID, IP: "127.0.0.1", UserAgent: "test",
	})
}

// ========== ListUsers / GetUserDetail ==========

func TestListUsers_DefaultPaging(t *testing.T) {
	fx := newFixture(t)
	fx.adminRepo.users = []repo.UserSummary{{ID: 1, Email: "a@b.com"}}
	out, err := fx.svc.ListUsers(context.Background(), UserListFilter{})
	require.NoError(t, err)
	assert.Equal(t, 1, out.Page)
	assert.Equal(t, 20, out.PageSize)
	assert.Equal(t, int64(1), out.Total)
	assert.Len(t, out.Items, 1)
}

func TestGetUserDetail_NotFound(t *testing.T) {
	fx := newFixture(t)
	_, err := fx.svc.GetUserDetail(context.Background(), 99)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

// ========== FreezeUser ==========

func TestFreezeUser_Success(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	err := fx.svc.FreezeUser(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, models.StatusFrozen, fx.adminRepo.updatedStatus[2])
	require.Len(t, fx.auditRepo.entries, 1)
	assert.Equal(t, "freeze_user", string(fx.auditRepo.entries[0].Action))
}

func TestFreezeUser_CannotFreezeSelf(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	err := fx.svc.FreezeUser(ctx, 1) // admin freezing himself
	assert.ErrorIs(t, err, ErrCannotFreezeSelf)
	assert.Empty(t, fx.adminRepo.updatedStatus)
	assert.Empty(t, fx.auditRepo.entries)
}

func TestFreezeUser_MissingContext(t *testing.T) {
	fx := newFixture(t)
	err := fx.svc.FreezeUser(context.Background(), 2) // no admin ctx
	assert.ErrorIs(t, err, auditsvc.ErrMissingRequestInfo)
}

func TestFreezeUser_IdempotentWhenAlreadyFrozen(t *testing.T) {
	fx := newFixture(t)
	fx.targetUser.Status = models.StatusFrozen
	ctx := ctxWithAdmin(1)
	err := fx.svc.FreezeUser(ctx, 2)
	require.NoError(t, err)
	assert.Empty(t, fx.adminRepo.updatedStatus, "no-op when already frozen")
	assert.Empty(t, fx.auditRepo.entries, "no audit when no-op")
}

func TestFreezeUser_NotFound(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	err := fx.svc.FreezeUser(ctx, 999)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

// ========== UnfreezeUser ==========

func TestUnfreezeUser_Success(t *testing.T) {
	fx := newFixture(t)
	fx.targetUser.Status = models.StatusFrozen
	ctx := ctxWithAdmin(1)
	err := fx.svc.UnfreezeUser(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, models.StatusActive, fx.adminRepo.updatedStatus[2])
	require.Len(t, fx.auditRepo.entries, 1)
	assert.Equal(t, "unfreeze_user", string(fx.auditRepo.entries[0].Action))
}

func TestUnfreezeUser_IdempotentWhenActive(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	err := fx.svc.UnfreezeUser(ctx, 2)
	require.NoError(t, err)
	assert.Empty(t, fx.adminRepo.updatedStatus)
	assert.Empty(t, fx.auditRepo.entries)
}

// ========== GrantBadge / RevokeBadge ==========

func TestGrantBadge_Success(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	badge, err := fx.svc.GrantBadge(ctx, 2, "first_prompt")
	require.NoError(t, err)
	assert.Equal(t, "first_prompt", badge.ID)
	_, unlocked := fx.badgeRepo.unlocked[2]["first_prompt"]
	assert.True(t, unlocked)
	require.Len(t, fx.auditRepo.entries, 1)
	assert.Equal(t, "grant_badge", string(fx.auditRepo.entries[0].Action))
}

func TestGrantBadge_UnknownBadge(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	_, err := fx.svc.GrantBadge(ctx, 2, "nonexistent")
	assert.ErrorIs(t, err, ErrBadgeNotFound)
}

func TestGrantBadge_AlreadyUnlocked(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	_, err := fx.svc.GrantBadge(ctx, 2, "first_prompt")
	require.NoError(t, err)

	_, err = fx.svc.GrantBadge(ctx, 2, "first_prompt")
	assert.ErrorIs(t, err, ErrBadgeAlreadyUnlocked)
}

func TestGrantBadge_UserNotFound(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	_, err := fx.svc.GrantBadge(ctx, 999, "first_prompt")
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestRevokeBadge_Success(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	_, err := fx.svc.GrantBadge(ctx, 2, "first_prompt")
	require.NoError(t, err)

	err = fx.svc.RevokeBadge(ctx, 2, "first_prompt")
	require.NoError(t, err)
	assert.Empty(t, fx.badgeRepo.unlocked[2])
	require.Len(t, fx.auditRepo.entries, 2)
	assert.Equal(t, "revoke_badge", string(fx.auditRepo.entries[1].Action))
}

func TestRevokeBadge_UnknownBadge(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	err := fx.svc.RevokeBadge(ctx, 2, "nonexistent")
	assert.ErrorIs(t, err, ErrBadgeNotFound)
}

// ========== ChangeTier ==========

func TestChangeTier_Success_NoActiveSub(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	// User free → max, без подписки.
	err := fx.svc.ChangeTier(ctx, 2, "max", "Manual upgrade per ticket #42")
	require.NoError(t, err)

	assert.Equal(t, "max", fx.users.users[2].PlanID, "plan_id обновлён")
	assert.Empty(t, fx.subsRepo.markedExpired, "MarkExpired не вызван если подписки нет")
	require.Len(t, fx.auditRepo.entries, 1)

	entry := fx.auditRepo.entries[0]
	assert.Equal(t, "change_tier", string(entry.Action))
	// AuditLog payload — JSON-marshalled, сравним поля через json.Unmarshal.
	var after map[string]any
	require.NoError(t, json.Unmarshal(entry.AfterState, &after))
	assert.Equal(t, "max", after["plan_id"])
	assert.Equal(t, "Manual upgrade per ticket #42", after["reason"])
	assert.Equal(t, "admin_override", after["source"])

	require.Len(t, fx.notifier.calls, 1)
	assert.Equal(t, "user@example.com", fx.notifier.calls[0].to)
	assert.Equal(t, "free", fx.notifier.calls[0].oldPlan)
	assert.Equal(t, "max", fx.notifier.calls[0].newPlan)
}

func TestChangeTier_Success_CancelsActiveSub(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)

	fx.targetUser.PlanID = "pro"
	fx.subsRepo.subsByUser[2] = &models.Subscription{ID: 100, UserID: 2, PlanID: "pro", Status: models.SubStatusActive}

	err := fx.svc.ChangeTier(ctx, 2, "max", "")
	require.NoError(t, err)
	assert.Equal(t, []uint{100}, fx.subsRepo.markedExpired, "active sub помечена expired")
	assert.Equal(t, "max", fx.users.users[2].PlanID)
}

func TestChangeTier_Success_PausedSub(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)

	fx.targetUser.PlanID = "free" // paused юзер уже на free пока пауза
	fx.subsRepo.subsByUser[2] = &models.Subscription{ID: 200, UserID: 2, PlanID: "pro", Status: models.SubStatusPaused}

	err := fx.svc.ChangeTier(ctx, 2, "max", "")
	require.NoError(t, err)
	assert.Equal(t, []uint{200}, fx.subsRepo.markedExpired, "paused sub тоже завершается")
	assert.Equal(t, "max", fx.users.users[2].PlanID)
}

func TestChangeTier_SameTier_Idempotent(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	fx.targetUser.PlanID = "max"

	err := fx.svc.ChangeTier(ctx, 2, "max", "")
	require.NoError(t, err)
	assert.Empty(t, fx.auditRepo.entries, "no audit для no-op")
	assert.Empty(t, fx.notifier.calls, "no email для no-op")
	assert.Empty(t, fx.subsRepo.markedExpired)
}

func TestChangeTier_InvalidTier(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	err := fx.svc.ChangeTier(ctx, 2, "enterprise", "")
	assert.ErrorIs(t, err, ErrInvalidTier)
}

func TestChangeTier_UserNotFound(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	err := fx.svc.ChangeTier(ctx, 999, "max", "")
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestChangeTier_AuditPayload(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)

	require.NoError(t, fx.svc.ChangeTier(ctx, 2, "pro", "support ticket"))
	require.Len(t, fx.auditRepo.entries, 1)

	var before map[string]any
	require.NoError(t, json.Unmarshal(fx.auditRepo.entries[0].BeforeState, &before))
	assert.Equal(t, "free", before["plan_id"])

	var after map[string]any
	require.NoError(t, json.Unmarshal(fx.auditRepo.entries[0].AfterState, &after))
	assert.Equal(t, "pro", after["plan_id"])
	assert.Equal(t, "support ticket", after["reason"])
	assert.Equal(t, "admin_override", after["source"])
}

func TestChangeTier_NotifierFailure_DoesNotBlock(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	fx.notifier.sendErr = errors.New("smtp unreachable")

	err := fx.svc.ChangeTier(ctx, 2, "pro", "")
	require.NoError(t, err, "notifier failure must not block ChangeTier")
	assert.Equal(t, "pro", fx.users.users[2].PlanID, "plan всё равно обновлён")
	require.Len(t, fx.auditRepo.entries, 1, "audit всё равно записан")
}

func TestChangeTier_MissingContext(t *testing.T) {
	fx := newFixture(t)
	err := fx.svc.ChangeTier(context.Background(), 2, "max", "")
	assert.ErrorIs(t, err, auditsvc.ErrMissingRequestInfo)
}

func TestChangeTier_GetCurrentByUserID_DBError_Aborts(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	fx.subsRepo.getCurrentErr = errors.New("db down")

	err := fx.svc.ChangeTier(ctx, 2, "max", "")
	require.Error(t, err, "ошибка lookup'а подписки должна прерывать ChangeTier")
	assert.Equal(t, "free", fx.users.users[2].PlanID, "plan_id НЕ обновлён при сбое lookup")
	assert.Empty(t, fx.auditRepo.entries, "audit НЕ пишется при сбое")
}

func TestChangeTier_MarkExpiredFails_Aborts(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	// Active sub есть, но MarkExpired падает.
	fx.subsRepo.subsByUser[2] = &models.Subscription{ID: 100, UserID: 2, PlanID: "pro", Status: models.SubStatusActive}
	fx.subsRepo.markExpiredErr = errors.New("transaction failed")

	err := fx.svc.ChangeTier(ctx, 2, "max", "")
	require.Error(t, err, "ошибка MarkExpired должна прерывать ChangeTier")
	assert.Equal(t, "free", fx.users.users[2].PlanID, "plan_id НЕ обновлён — иначе drift со старой sub")
	assert.Empty(t, fx.auditRepo.entries)
}

// ========== helpers tests ==========

func TestUserStateSnapshot_ExcludesSensitiveFields(t *testing.T) {
	user := &models.User{
		ID:           42,
		Email:        "x@y.z",
		PasswordHash: "secret",
		TokenNonce:   "nonce",
		Role:         models.RoleAdmin,
		Status:       models.StatusActive,
	}
	snap := userStateSnapshot(user)
	assert.Equal(t, uint(42), snap["id"])
	assert.Equal(t, "x@y.z", snap["email"])
	assert.NotContains(t, snap, "password_hash")
	assert.NotContains(t, snap, "token_nonce")
}

// Sanity check: helper compile-check for errors.Is in fake ForgotPassword flow.
var _ = errors.Is
