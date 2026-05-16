package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// minimal stub UserRepository — track GetByID calls.
type stubUserRepo struct {
	getByIDCalls int
	user         *models.User
	err          error
}

func (s *stubUserRepo) GetByID(_ context.Context, _ uint) (*models.User, error) {
	s.getByIDCalls++
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

// Stub other UserRepository methods (panic — must not be called).
func (s *stubUserRepo) Create(context.Context, *models.User) error          { panic("unused") }
func (s *stubUserRepo) GetByEmail(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (s *stubUserRepo) GetByUsername(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (s *stubUserRepo) SearchUsers(context.Context, string, int) ([]models.User, error) {
	panic("unused")
}
func (s *stubUserRepo) Update(context.Context, *models.User) error  { panic("unused") }
func (s *stubUserRepo) SetPlan(context.Context, uint, string) error { panic("unused") }
func (s *stubUserRepo) SetQuotaWarningSentOn(context.Context, uint, time.Time) error {
	panic("unused")
}
func (s *stubUserRepo) TouchLastLogin(context.Context, uint) error { panic("unused") }
func (s *stubUserRepo) ListInactiveForReengagement(context.Context, time.Time, time.Time, int) ([]models.User, error) {
	panic("unused")
}
func (s *stubUserRepo) MarkReengagementSent(context.Context, uint) error { panic("unused") }
func (s *stubUserRepo) CountReferredBy(context.Context, string) (int64, error) {
	panic("unused")
}
func (s *stubUserRepo) GetByReferralCode(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (s *stubUserRepo) MarkReferralRewarded(context.Context, uint) (bool, error) {
	panic("unused")
}
func (s *stubUserRepo) ListPaidUsers(context.Context) ([]uint, error) { panic("unused") }
func (s *stubUserRepo) SetInsightEmailsEnabled(context.Context, uint, bool) error {
	panic("unused")
}

var _ repo.UserRepository = (*stubUserRepo)(nil)

// dummyKey — копия middleware/auth.ClaimsKey для теста (избегаем cyclic import).
type dummyKey string

const testClaimsKey dummyKey = "claims"

func TestPlanIDOrFallback_FromClaims_NoDBHit(t *testing.T) {
	ctx := context.WithValue(context.Background(), testClaimsKey, &Claims{UserID: 1, PlanID: "max"})
	users := &stubUserRepo{}

	plan, err := PlanIDOrFallback(ctx, users, 1, testClaimsKey)
	require.NoError(t, err)
	assert.Equal(t, "max", plan)
	assert.Equal(t, 0, users.getByIDCalls, "claims hit — no DB call")
}

func TestPlanIDOrFallback_LegacyJWT_FallsBackToDB(t *testing.T) {
	// Legacy claims без PlanID → fallback.
	ctx := context.WithValue(context.Background(), testClaimsKey, &Claims{UserID: 1, PlanID: ""})
	users := &stubUserRepo{user: &models.User{ID: 1, PlanID: "pro"}}

	plan, err := PlanIDOrFallback(ctx, users, 1, testClaimsKey)
	require.NoError(t, err)
	assert.Equal(t, "pro", plan)
	assert.Equal(t, 1, users.getByIDCalls, "legacy claims — fallback на DB")
}

func TestPlanIDOrFallback_NoClaims_FallsBackToDB(t *testing.T) {
	ctx := context.Background() // нет claims вообще
	users := &stubUserRepo{user: &models.User{ID: 1, PlanID: "free"}}

	plan, err := PlanIDOrFallback(ctx, users, 1, testClaimsKey)
	require.NoError(t, err)
	assert.Equal(t, "free", plan)
	assert.Equal(t, 1, users.getByIDCalls)
}

func TestPlanIDOrFallback_DBError_PropagatesError(t *testing.T) {
	ctx := context.Background()
	users := &stubUserRepo{err: errors.New("db down")}

	_, err := PlanIDOrFallback(ctx, users, 1, testClaimsKey)
	assert.Error(t, err)
}

func TestFromContext_LocalKey(t *testing.T) {
	c := &Claims{UserID: 42, PlanID: "max"}
	ctx := WithContext(context.Background(), c)
	got, ok := FromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, c, got)
}

func TestFromContext_Empty(t *testing.T) {
	_, ok := FromContext(context.Background())
	assert.False(t, ok)
}
