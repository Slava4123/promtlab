package badge

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authmw "promptvault/internal/middleware/auth"
	badgeuc "promptvault/internal/usecases/badge"
)

// --- fake Service ---

type fakeBadgeService struct {
	items []badgeuc.BadgeWithState
	err   error
}

func (f *fakeBadgeService) List(_ context.Context, _ uint) ([]badgeuc.BadgeWithState, error) {
	return f.items, f.err
}

// withUserContext строит http.Request с userID в context как это делает
// authmw.Middleware. Использует экспортированный authmw.UserIDKey.
func withUserContext(userID uint) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/badges", nil)
	ctx := context.WithValue(req.Context(), authmw.UserIDKey, userID)
	return req.WithContext(ctx)
}

// --- tests ---

func TestHandler_List_EmptyState(t *testing.T) {
	svc := &fakeBadgeService{items: []badgeuc.BadgeWithState{}}
	h := NewHandler(svc)

	rec := httptest.NewRecorder()
	h.List(rec, withUserContext(1))

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp BadgeListResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, 0, resp.TotalCount)
	assert.Equal(t, 0, resp.TotalUnlocked)
	assert.Empty(t, resp.Items)
}

func TestHandler_List_MixedStates(t *testing.T) {
	unlockedTime := time.Now().Add(-1 * time.Hour)
	items := []badgeuc.BadgeWithState{
		{
			Badge: badgeuc.Badge{
				ID: "first_prompt", Title: "Первопроходец",
				Description: "Создай первый личный промпт", Icon: "🎯",
				Category: badgeuc.CategoryPersonal,
			},
			Unlocked:   true,
			UnlockedAt: &unlockedTime,
			Progress:   1,
			Target:     1,
		},
		{
			Badge: badgeuc.Badge{
				ID: "architect", Title: "Архитектор",
				Description: "Создай 10 личных промптов", Icon: "🏗️",
				Category: badgeuc.CategoryPersonal,
			},
			Unlocked: false,
			Progress: 3,
			Target:   10,
		},
		{
			Badge: badgeuc.Badge{
				ID: "on_fire", Title: "На огне",
				Description: "7 дней подряд", Icon: "🔥",
				Category: badgeuc.CategoryStreak,
			},
			Unlocked: false,
			Progress: 0,
			Target:   7,
		},
	}
	svc := &fakeBadgeService{items: items}
	h := NewHandler(svc)

	rec := httptest.NewRecorder()
	h.List(rec, withUserContext(42))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp BadgeListResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	assert.Equal(t, 3, resp.TotalCount)
	assert.Equal(t, 1, resp.TotalUnlocked)
	require.Len(t, resp.Items, 3)

	// first_prompt — unlocked
	assert.Equal(t, "first_prompt", resp.Items[0].ID)
	assert.Equal(t, "Первопроходец", resp.Items[0].Title)
	assert.Equal(t, "🎯", resp.Items[0].Icon)
	assert.Equal(t, "personal", resp.Items[0].Category)
	assert.True(t, resp.Items[0].Unlocked)
	require.NotNil(t, resp.Items[0].UnlockedAt)
	assert.WithinDuration(t, unlockedTime, *resp.Items[0].UnlockedAt, time.Second)
	assert.Equal(t, int64(1), resp.Items[0].Progress)
	assert.Equal(t, int64(1), resp.Items[0].Target)

	// architect — locked с прогрессом
	assert.Equal(t, "architect", resp.Items[1].ID)
	assert.False(t, resp.Items[1].Unlocked)
	assert.Nil(t, resp.Items[1].UnlockedAt, "unlocked_at должен отсутствовать для locked")
	assert.Equal(t, int64(3), resp.Items[1].Progress)
	assert.Equal(t, int64(10), resp.Items[1].Target)

	// on_fire — locked с прогрессом 0
	assert.Equal(t, "on_fire", resp.Items[2].ID)
	assert.Equal(t, "streak", resp.Items[2].Category)
	assert.False(t, resp.Items[2].Unlocked)
	assert.Equal(t, int64(0), resp.Items[2].Progress)
}

func TestHandler_List_ServiceError_Returns500(t *testing.T) {
	svc := &fakeBadgeService{err: errors.New("db connection lost")}
	h := NewHandler(svc)

	rec := httptest.NewRecorder()
	h.List(rec, withUserContext(1))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// --- NewBadgeSummaries helper tests ---

func TestNewBadgeSummaries_Empty(t *testing.T) {
	assert.Nil(t, NewBadgeSummaries(nil), "nil input → nil output (для omitempty)")
	assert.Nil(t, NewBadgeSummaries([]badgeuc.Badge{}), "empty slice → nil output")
}

func TestNewBadgeSummaries_Populated(t *testing.T) {
	input := []badgeuc.Badge{
		{ID: "first_prompt", Title: "Первопроходец", Description: "D1", Icon: "🎯"},
		{ID: "architect", Title: "Архитектор", Description: "D2", Icon: "🏗️"},
	}
	out := NewBadgeSummaries(input)
	require.Len(t, out, 2)
	assert.Equal(t, "first_prompt", out[0].ID)
	assert.Equal(t, "Первопроходец", out[0].Title)
	assert.Equal(t, "🎯", out[0].Icon)
	assert.Equal(t, "architect", out[1].ID)
}
