package auth

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/models"
)

// --- UserRepository mock ---

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) SearchUsers(ctx context.Context, query string, limit int) ([]models.User, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) SetQuotaWarningSentOn(ctx context.Context, userID uint, date time.Time) error {
	return m.Called(ctx, userID, date).Error(0)
}

// TouchLastLogin не проверяется в большинстве тестов (вызывается из background
// горутины после login — неважно для unit-теста). No-op по умолчанию.
func (m *mockUserRepo) TouchLastLogin(_ context.Context, _ uint) error {
	return nil
}
func (m *mockUserRepo) ListInactiveForReengagement(ctx context.Context, inactiveBefore, sentBefore time.Time, limit int) ([]models.User, error) {
	args := m.Called(ctx, inactiveBefore, sentBefore, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *mockUserRepo) MarkReengagementSent(ctx context.Context, userID uint) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockUserRepo) CountReferredBy(ctx context.Context, code string) (int64, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockUserRepo) GetByReferralCode(ctx context.Context, code string) (*models.User, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) MarkReferralRewarded(ctx context.Context, userID uint) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepo) ListMaxUsers(_ context.Context) ([]uint, error) { return nil, nil }
func (m *mockUserRepo) SetInsightEmailsEnabled(_ context.Context, _ uint, _ bool) error {
	return nil
}

// --- LinkedAccountRepository mock ---

type mockLinkedAccountRepo struct{ mock.Mock }

func (m *mockLinkedAccountRepo) Create(ctx context.Context, la *models.LinkedAccount) error {
	return m.Called(ctx, la).Error(0)
}
func (m *mockLinkedAccountRepo) GetByUserID(ctx context.Context, userID uint) ([]models.LinkedAccount, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.LinkedAccount), args.Error(1)
}
func (m *mockLinkedAccountRepo) GetByProviderID(ctx context.Context, provider, providerID string) (*models.LinkedAccount, error) {
	args := m.Called(ctx, provider, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LinkedAccount), args.Error(1)
}
func (m *mockLinkedAccountRepo) Delete(ctx context.Context, userID uint, provider string) error {
	return m.Called(ctx, userID, provider).Error(0)
}
func (m *mockLinkedAccountRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// --- VerificationRepository mock ---

type mockVerificationRepo struct{ mock.Mock }

func (m *mockVerificationRepo) Create(ctx context.Context, v *models.EmailVerification) error {
	return m.Called(ctx, v).Error(0)
}
func (m *mockVerificationRepo) GetByUserID(ctx context.Context, userID uint) (*models.EmailVerification, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EmailVerification), args.Error(1)
}
func (m *mockVerificationRepo) IncrementAttempts(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockVerificationRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	return m.Called(ctx, userID).Error(0)
}

// --- Test helpers ---

func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:          "test-secret-at-least-32-chars-long!!",
			AccessDuration:  "15m",
			RefreshDuration: "168h",
		},
	}
}

func newTestService(users *mockUserRepo, linked *mockLinkedAccountRepo, verif *mockVerificationRepo) *Service {
	return NewService(testConfig(), users, linked, verif, nil)
}
