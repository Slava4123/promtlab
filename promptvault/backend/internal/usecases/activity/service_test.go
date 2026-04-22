package activity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ===================== Mocks =====================

type mockActivityRepo struct{ mock.Mock }

func (m *mockActivityRepo) Log(ctx context.Context, event *models.TeamActivityLog) error {
	return m.Called(ctx, event).Error(0)
}
func (m *mockActivityRepo) List(ctx context.Context, filter repo.TeamActivityFilter) ([]models.TeamActivityLog, *time.Time, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]models.TeamActivityLog), args.Get(1).(*time.Time), args.Error(2)
}
func (m *mockActivityRepo) ListByTarget(ctx context.Context, targetType string, targetID uint, limit int) ([]models.TeamActivityLog, error) {
	args := m.Called(ctx, targetType, targetID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamActivityLog), args.Error(1)
}
func (m *mockActivityRepo) AnonymizeActor(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockActivityRepo) DeleteOlderThan(ctx context.Context, teamID uint, before time.Time) (int64, error) {
	args := m.Called(ctx, teamID, before)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockActivityRepo) CleanupByRetention(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

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
	return nil, nil
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) SearchUsers(ctx context.Context, query string, limit int) ([]models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error { return nil }
func (m *mockUserRepo) SetQuotaWarningSentOn(ctx context.Context, userID uint, date time.Time) error {
	return nil
}
func (m *mockUserRepo) TouchLastLogin(ctx context.Context, userID uint) error { return nil }
func (m *mockUserRepo) ListInactiveForReengagement(ctx context.Context, inactiveBefore, sentBefore time.Time, limit int) ([]models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) MarkReengagementSent(ctx context.Context, userID uint) error { return nil }
func (m *mockUserRepo) CountReferredBy(ctx context.Context, code string) (int64, error) {
	return 0, nil
}
func (m *mockUserRepo) GetByReferralCode(ctx context.Context, code string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) MarkReferralRewarded(ctx context.Context, userID uint) (bool, error) {
	return false, nil
}
func (m *mockUserRepo) ListMaxUsers(_ context.Context) ([]uint, error) { return nil, nil }

func newTestActivityService() (*Service, *mockActivityRepo, *mockUserRepo) {
	ar := new(mockActivityRepo)
	ur := new(mockUserRepo)
	return NewService(ar, ur), ar, ur
}

// ===================== Tests =====================

// TestLog_ValidatesRequiredFields: пустой TeamID, EventType/TargetType, отсутствие
// актора — каждое условие должно падать на своём домен-ошибке из errors.go.
func TestLog_ValidatesRequiredFields(t *testing.T) {
	cases := []struct {
		name    string
		event   Event
		wantErr error
	}{
		{
			name:    "missing team",
			event:   Event{EventType: "x", TargetType: "y", ActorEmail: "a@b"},
			wantErr: ErrMissingTeam,
		},
		{
			name:    "missing event_type",
			event:   Event{TeamID: 1, TargetType: "y", ActorEmail: "a@b"},
			wantErr: ErrMissingEventType,
		},
		{
			name:    "missing target_type",
			event:   Event{TeamID: 1, EventType: "x", ActorEmail: "a@b"},
			wantErr: ErrMissingEventType,
		},
		{
			name:    "missing actor (no ID, no email)",
			event:   Event{TeamID: 1, EventType: "x", TargetType: "y"},
			wantErr: ErrMissingActor,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, _ := newTestActivityService()
			err := svc.Log(context.Background(), tc.event)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// TestLog_ResolvesActorFromUserRepo: ActorID есть, ActorEmail пуст → сервис
// достаёт email/name из UserRepository и записывает в repo.
func TestLog_ResolvesActorFromUserRepo(t *testing.T) {
	svc, ar, ur := newTestActivityService()
	ctx := context.Background()

	ur.On("GetByID", ctx, uint(42)).Return(&models.User{ID: 42, Email: "jane@example.com", Name: "Jane"}, nil)
	ar.On("Log", ctx, mock.MatchedBy(func(ev *models.TeamActivityLog) bool {
		return ev.ActorEmail == "jane@example.com" &&
			ev.ActorName == "Jane" &&
			ev.ActorID != nil && *ev.ActorID == 42 &&
			ev.TeamID == 1 && ev.EventType == "prompt.created"
	})).Return(nil)

	err := svc.Log(ctx, Event{
		TeamID:     1,
		ActorID:    42,
		EventType:  "prompt.created",
		TargetType: models.TargetPrompt,
	})
	assert.NoError(t, err)
	ar.AssertExpectations(t)
}

// TestLog_UserLookupFailurePropagates: если GetByID возвращает ошибку и
// ActorEmail пуст — Log возвращает обёрнутую ошибку (не молчаливый success).
func TestLog_UserLookupFailurePropagates(t *testing.T) {
	svc, _, ur := newTestActivityService()
	ctx := context.Background()
	ur.On("GetByID", ctx, uint(42)).Return(nil, errors.New("db down"))

	err := svc.Log(ctx, Event{
		TeamID:     1,
		ActorID:    42,
		EventType:  "x",
		TargetType: "y",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

// TestLog_UsesExplicitActorEmailWithoutUserLookup: если ActorEmail явный,
// users.GetByID не должен вызываться (system-events, robots).
func TestLog_UsesExplicitActorEmailWithoutUserLookup(t *testing.T) {
	svc, ar, ur := newTestActivityService()
	ctx := context.Background()
	ar.On("Log", ctx, mock.AnythingOfType("*models.TeamActivityLog")).Return(nil)

	err := svc.Log(ctx, Event{
		TeamID:     1,
		ActorEmail: "system@promptvault",
		EventType:  "system.event",
		TargetType: models.TargetMember,
	})
	assert.NoError(t, err)
	ur.AssertNumberOfCalls(t, "GetByID", 0)
}

// TestLogSafe_NilServiceIsNoOp: service заявлен как nil-safe в хуках
// из других usecases — вызов на nil-ресивере не должен паниковать.
func TestLogSafe_NilServiceIsNoOp(t *testing.T) {
	var svc *Service
	assert.NotPanics(t, func() {
		svc.LogSafe(context.Background(), Event{TeamID: 1, EventType: "x", TargetType: "y", ActorEmail: "a@b"})
	})
}

// TestLogSafe_SwallowsErrors: Log fail'ится (валидация / repo) — LogSafe
// не должен пробрасывать ошибку (это explicit design: log miss ≠ broken flow).
func TestLogSafe_SwallowsErrors(t *testing.T) {
	svc, _, _ := newTestActivityService()
	assert.NotPanics(t, func() {
		// Event без TeamID — Log вернёт ErrMissingTeam, LogSafe проглатывает.
		svc.LogSafe(context.Background(), Event{EventType: "x", TargetType: "y", ActorEmail: "a@b"})
	})
}

// TestLogSafe_HappyPathForwardsToRepo: валидный Event → repo.Log вызван.
func TestLogSafe_HappyPathForwardsToRepo(t *testing.T) {
	svc, ar, _ := newTestActivityService()
	ctx := context.Background()
	ar.On("Log", ctx, mock.AnythingOfType("*models.TeamActivityLog")).Return(nil)

	svc.LogSafe(ctx, Event{
		TeamID:     1,
		ActorEmail: "a@b",
		EventType:  "x",
		TargetType: "y",
	})
	ar.AssertNumberOfCalls(t, "Log", 1)
}

// TestAnonymizeActor_DelegatesToRepo: service просто пробрасывает вызов
// в repo.AnonymizeActor — GDPR flow замены actor_*/actor_id=NULL.
func TestAnonymizeActor_DelegatesToRepo(t *testing.T) {
	svc, ar, _ := newTestActivityService()
	ctx := context.Background()
	ar.On("AnonymizeActor", ctx, uint(42)).Return(int64(7), nil)

	count, err := svc.AnonymizeActor(ctx, 42)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), count)
	ar.AssertExpectations(t)
}

// TestGetPromptHistory_DelegatesToListByTarget: склейка с версиями — на
// стороне handler, сервис просто вызывает repo.ListByTarget.
func TestGetPromptHistory_DelegatesToListByTarget(t *testing.T) {
	svc, ar, _ := newTestActivityService()
	ctx := context.Background()
	expected := []models.TeamActivityLog{{ID: 1}, {ID: 2}}
	ar.On("ListByTarget", ctx, models.TargetPrompt, uint(99), 100).Return(expected, nil)

	got, err := svc.GetPromptHistory(ctx, 99, 100)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}
