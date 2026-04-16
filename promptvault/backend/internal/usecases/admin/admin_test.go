package admin

import (
	"context"
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

// ========== Test helpers ==========

type fixture struct {
	svc          *Service
	users        *fakeUserRepo
	adminRepo    *fakeAdminRepo
	auditRepo    *fakeAuditRepo
	badgeRepo    *fakeBadgeRepo
	badgeSvc     *badgeuc.Service
	adminUser    *models.User
	targetUser   *models.User
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	users := newFakeUserRepo()
	adminUser := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin, Status: models.StatusActive}
	target := &models.User{ID: 2, Email: "user@example.com", Role: models.RoleUser, Status: models.StatusActive}
	users.users[1] = adminUser
	users.users[2] = target

	adminRepo := newFakeAdminRepo()
	adminRepo.detailByID[2] = &repo.UserDetail{User: target, PromptCount: 10}

	auditRepo := &fakeAuditRepo{}
	auditSvc := auditsvc.NewService(auditRepo)

	badgeRepo := newFakeBadgeRepo()
	badgeSvc, err := badgeuc.NewService(badgeRepo, nil)
	require.NoError(t, err)

	svc := NewService(adminRepo, users, auditSvc, nil, badgeSvc, nil, nil)

	return &fixture{
		svc:        svc,
		users:      users,
		adminRepo:  adminRepo,
		auditRepo:  auditRepo,
		badgeRepo:  badgeRepo,
		badgeSvc:   badgeSvc,
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
	assert.Equal(t, "freeze_user", fx.auditRepo.entries[0].Action)
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
	assert.Equal(t, "unfreeze_user", fx.auditRepo.entries[0].Action)
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
	assert.Equal(t, "grant_badge", fx.auditRepo.entries[0].Action)
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
	assert.Equal(t, "revoke_badge", fx.auditRepo.entries[1].Action)
}

func TestRevokeBadge_UnknownBadge(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	err := fx.svc.RevokeBadge(ctx, 2, "nonexistent")
	assert.ErrorIs(t, err, ErrBadgeNotFound)
}

// ========== ChangeTier ==========

func TestChangeTier_InvalidPlan(t *testing.T) {
	fx := newFixture(t)
	ctx := ctxWithAdmin(1)
	// plans repo is nil → returns ErrInvalidTier
	err := fx.svc.ChangeTier(ctx, 2, "pro")
	assert.ErrorIs(t, err, ErrInvalidTier)
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
