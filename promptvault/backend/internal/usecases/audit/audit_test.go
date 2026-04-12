package audit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- fakeAuditRepo ---

type fakeAuditRepo struct {
	entries []models.AuditLog
	logErr  error
}

func (f *fakeAuditRepo) Log(_ context.Context, entry *models.AuditLog) error {
	if f.logErr != nil {
		return f.logErr
	}
	entry.ID = uint(len(f.entries) + 1)
	f.entries = append(f.entries, *entry)
	return nil
}

func (f *fakeAuditRepo) List(_ context.Context, filter repo.AuditLogFilter) ([]models.AuditLog, int64, error) {
	return f.entries, int64(len(f.entries)), nil
}

// --- tests ---

func TestLog_Success(t *testing.T) {
	fake := &fakeAuditRepo{}
	svc := NewService(fake)

	ctx := WithContext(context.Background(), AdminRequestInfo{
		AdminID:   1,
		IP:        "127.0.0.1",
		UserAgent: "test",
	})

	targetID := uint(42)
	err := svc.Log(ctx, LogInput{
		Action:     ActionGrantBadge,
		TargetType: TargetUser,
		TargetID:   &targetID,
		BeforeState: nil,
		AfterState: map[string]any{
			"badge_id": "first_prompt",
		},
	})
	require.NoError(t, err)
	require.Len(t, fake.entries, 1)

	e := fake.entries[0]
	assert.Equal(t, uint(1), e.AdminID)
	assert.Equal(t, "grant_badge", e.Action)
	assert.Equal(t, "user", e.TargetType)
	require.NotNil(t, e.TargetID)
	assert.Equal(t, uint(42), *e.TargetID)
	assert.Equal(t, "127.0.0.1", e.IP)
	assert.Equal(t, "test", e.UserAgent)
	assert.Nil(t, e.BeforeState)

	var after map[string]any
	require.NoError(t, json.Unmarshal(e.AfterState, &after))
	assert.Equal(t, "first_prompt", after["badge_id"])
}

func TestLog_MissingContextReturnsError(t *testing.T) {
	fake := &fakeAuditRepo{}
	svc := NewService(fake)

	// Context БЕЗ AdminRequestInfo — typical bug.
	err := svc.Log(context.Background(), LogInput{
		Action:     ActionGrantBadge,
		TargetType: TargetUser,
	})
	assert.ErrorIs(t, err, ErrMissingRequestInfo)
	assert.Empty(t, fake.entries)
}

func TestLog_RepoErrorPropagates(t *testing.T) {
	fake := &fakeAuditRepo{logErr: errors.New("db down")}
	svc := NewService(fake)

	ctx := WithContext(context.Background(), AdminRequestInfo{AdminID: 1, IP: "127.0.0.1"})
	err := svc.Log(ctx, LogInput{
		Action:     ActionGrantBadge,
		TargetType: TargetUser,
	})
	assert.Error(t, err)
}

func TestLog_NilStatesAreNotMarshaled(t *testing.T) {
	fake := &fakeAuditRepo{}
	svc := NewService(fake)
	ctx := WithContext(context.Background(), AdminRequestInfo{AdminID: 1, IP: "127.0.0.1"})

	require.NoError(t, svc.Log(ctx, LogInput{
		Action:     ActionGrantBadge,
		TargetType: TargetUser,
	}))
	require.Len(t, fake.entries, 1)

	e := fake.entries[0]
	assert.Nil(t, e.BeforeState, "nil state не должен быть marshal'ен в JSON null")
	assert.Nil(t, e.AfterState)
}

func TestWithContextFromContext_RoundTrip(t *testing.T) {
	info := AdminRequestInfo{AdminID: 42, IP: "1.2.3.4", UserAgent: "curl"}
	ctx := WithContext(context.Background(), info)
	got, ok := FromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, info, got)
}

func TestFromContext_MissingReturnsFalse(t *testing.T) {
	_, ok := FromContext(context.Background())
	assert.False(t, ok)
}
