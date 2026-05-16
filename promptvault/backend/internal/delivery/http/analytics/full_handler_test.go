package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	analyticsuc "promptvault/internal/usecases/analytics"
)

// --- mockAnalyticsSvc ---

type mockAnalyticsSvc struct{ mock.Mock }

func (m *mockAnalyticsSvc) GetPersonalDashboard(ctx context.Context, userID uint, rng analyticsuc.RangeID) (*analyticsuc.PersonalDashboard, error) {
	args := m.Called(ctx, userID, rng)
	v, _ := args.Get(0).(*analyticsuc.PersonalDashboard)
	return v, args.Error(1)
}
func (m *mockAnalyticsSvc) GetPersonalDashboardFiltered(ctx context.Context, userID uint, rng analyticsuc.RangeID, tagID, collectionID *uint) (*analyticsuc.PersonalDashboard, error) {
	args := m.Called(ctx, userID, rng, tagID, collectionID)
	v, _ := args.Get(0).(*analyticsuc.PersonalDashboard)
	return v, args.Error(1)
}
func (m *mockAnalyticsSvc) GetTeamDashboard(ctx context.Context, userID, teamID uint, rng analyticsuc.RangeID) (*analyticsuc.TeamDashboard, error) {
	args := m.Called(ctx, userID, teamID, rng)
	v, _ := args.Get(0).(*analyticsuc.TeamDashboard)
	return v, args.Error(1)
}
func (m *mockAnalyticsSvc) GetTeamDashboardFiltered(ctx context.Context, userID, teamID uint, rng analyticsuc.RangeID, tagID, collectionID *uint) (*analyticsuc.TeamDashboard, error) {
	args := m.Called(ctx, userID, teamID, rng, tagID, collectionID)
	v, _ := args.Get(0).(*analyticsuc.TeamDashboard)
	return v, args.Error(1)
}
func (m *mockAnalyticsSvc) GetPromptAnalytics(ctx context.Context, promptID, userID uint) (*analyticsuc.PromptAnalytics, error) {
	args := m.Called(ctx, promptID, userID)
	v, _ := args.Get(0).(*analyticsuc.PromptAnalytics)
	return v, args.Error(1)
}
func (m *mockAnalyticsSvc) GetInsightsGated(ctx context.Context, userID uint, teamID *uint) ([]models.SmartInsight, error) {
	args := m.Called(ctx, userID, teamID)
	v, _ := args.Get(0).([]models.SmartInsight)
	return v, args.Error(1)
}
func (m *mockAnalyticsSvc) RefreshInsightsGated(ctx context.Context, userID uint, teamID *uint) ([]models.SmartInsight, error) {
	args := m.Called(ctx, userID, teamID)
	v, _ := args.Get(0).([]models.SmartInsight)
	return v, args.Error(1)
}
func (m *mockAnalyticsSvc) ExportGate(ctx context.Context, userID uint) error {
	return m.Called(ctx, userID).Error(0)
}

// --- helpers ---

func authedReq(t *testing.T, method, url string, userID uint) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, url, nil)
	ctx := context.WithValue(req.Context(), authmw.UserIDKey, userID)
	return req.WithContext(ctx)
}

// chiCtx подставляет URL params в chi.RouteContext (для теста Prompt/Team
// handler'ов, которым нужен chi.URLParam).
func chiCtx(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// --- Personal ---

func TestHandler_Personal_OK(t *testing.T) {
	svc := new(mockAnalyticsSvc)
	svc.On("GetPersonalDashboardFiltered", mock.Anything, uint(42), analyticsuc.Range7d, (*uint)(nil), (*uint)(nil)).
		Return(&analyticsuc.PersonalDashboard{Range: analyticsuc.Range7d}, nil)

	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.Personal(rr, authedReq(t, http.MethodGet, "/api/analytics/personal", 42))

	require.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Personal_DrillDown_PassesFilters(t *testing.T) {
	svc := new(mockAnalyticsSvc)
	tagID := uint(5)
	collectionID := uint(7)
	svc.On("GetPersonalDashboardFiltered", mock.Anything, uint(1), analyticsuc.Range30d, &tagID, &collectionID).
		Return(&analyticsuc.PersonalDashboard{Range: analyticsuc.Range30d}, nil)

	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.Personal(rr, authedReq(t, http.MethodGet, "/api/analytics/personal?range=30d&tag_id=5&collection_id=7", 1))

	require.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Personal_ServiceError_500(t *testing.T) {
	svc := new(mockAnalyticsSvc)
	svc.On("GetPersonalDashboardFiltered", mock.Anything, uint(1), analyticsuc.Range7d, (*uint)(nil), (*uint)(nil)).
		Return((*analyticsuc.PersonalDashboard)(nil), errors.New("db down"))

	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.Personal(rr, authedReq(t, http.MethodGet, "/api/analytics/personal", 1))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- Team ---

func TestHandler_Team_OK(t *testing.T) {
	svc := new(mockAnalyticsSvc)
	svc.On("GetTeamDashboardFiltered", mock.Anything, uint(1), uint(10), analyticsuc.Range7d, (*uint)(nil), (*uint)(nil)).
		Return(&analyticsuc.TeamDashboard{Range: analyticsuc.Range7d}, nil)

	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	req := authedReq(t, http.MethodGet, "/api/analytics/teams/10", 1)
	req = chiCtx(req, map[string]string{"id": "10"})
	h.Team(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestHandler_Team_InvalidID_400(t *testing.T) {
	h := NewHandler(new(mockAnalyticsSvc))
	rr := httptest.NewRecorder()
	req := authedReq(t, http.MethodGet, "/api/analytics/teams/abc", 1)
	req = chiCtx(req, map[string]string{"id": "abc"})
	h.Team(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- Prompt ---

func TestHandler_Prompt_NotFound_404(t *testing.T) {
	svc := new(mockAnalyticsSvc)
	svc.On("GetPromptAnalytics", mock.Anything, uint(99), uint(1)).
		Return((*analyticsuc.PromptAnalytics)(nil), analyticsuc.ErrNotFound)

	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	req := authedReq(t, http.MethodGet, "/api/analytics/prompts/99", 1)
	req = chiCtx(req, map[string]string{"id": "99"})
	h.Prompt(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- Insights (402 on non-paid) ---

// TestHandler_Insights_NonPaid_402_Pro — Pricing iteration v3 (Task 6+8):
// после рефактора GetInsightsGated больше не Max-only; Free → ErrProRequired,
// Pro/Max → данные (с фильтром по plan). Handler должен маппить ErrProRequired
// в 402 с plan="Pro" — для UI upgrade prompt.
//
// Раньше тест мокал ErrMaxRequired (insights были Max-only) — это semantic
// stale после Task 6.
func TestHandler_Insights_NonPaid_402_Pro(t *testing.T) {
	svc := new(mockAnalyticsSvc)
	svc.On("GetInsightsGated", mock.Anything, uint(1), (*uint)(nil)).
		Return(([]models.SmartInsight)(nil), analyticsuc.ErrProRequired)

	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.Insights(rr, authedReq(t, http.MethodGet, "/api/analytics/insights", 1))

	require.Equal(t, http.StatusPaymentRequired, rr.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	assert.Equal(t, "Pro", body["plan"], "ErrProRequired → plan='Pro' для upgrade prompt")
	assert.Equal(t, "/pricing", body["upgrade_url"])
	assert.Equal(t, "premium_feature", body["quota_type"])
}

// TestRespondError_InsightsProRequired_Returns402WithProPlan — unit-тест
// низкоуровневого маппинга respondError(ErrProRequired) → 402 с правильным
// body. Дублирует cover'едж handler-теста, но без mock'а service —
// гарантирует контракт самого маппинга на случай рефактора handler'ов.
func TestRespondError_InsightsProRequired_Returns402WithProPlan(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/analytics/insights", nil)
	respondError(rec, req, analyticsuc.ErrProRequired)

	require.Equal(t, http.StatusPaymentRequired, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "Pro", body["plan"])
	assert.Equal(t, "/pricing", body["upgrade_url"])
}

// TestRespondError_RefreshMaxRequired_Returns402WithMaxPlan — guard'ит
// что refresh endpoint (Max-only) остался на ErrMaxRequired → plan="Max".
// После Task 8 mapping ErrMaxRequired не трогали, но проверяем явно.
func TestRespondError_RefreshMaxRequired_Returns402WithMaxPlan(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/analytics/insights/refresh", nil)
	respondError(rec, req, analyticsuc.ErrMaxRequired)

	require.Equal(t, http.StatusPaymentRequired, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "Max", body["plan"])
	assert.Equal(t, "insights", body["quota_type"])
}

func TestHandler_Insights_Max_OK(t *testing.T) {
	svc := new(mockAnalyticsSvc)
	svc.On("GetInsightsGated", mock.Anything, uint(1), (*uint)(nil)).
		Return([]models.SmartInsight{{InsightType: "trending"}}, nil)

	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.Insights(rr, authedReq(t, http.MethodGet, "/api/analytics/insights", 1))

	require.Equal(t, http.StatusOK, rr.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	assert.Contains(t, body, "items")
}

// --- Export ---

func TestHandler_Export_FormatInvalid_400(t *testing.T) {
	h := NewHandler(new(mockAnalyticsSvc))
	rr := httptest.NewRecorder()
	h.Export(rr, authedReq(t, http.MethodGet, "/api/analytics/export?format=pdf", 1))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Export_FreeUser_402(t *testing.T) {
	svc := new(mockAnalyticsSvc)
	svc.On("ExportGate", mock.Anything, uint(1)).Return(analyticsuc.ErrProRequired)

	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.Export(rr, authedReq(t, http.MethodGet, "/api/analytics/export?format=csv", 1))

	assert.Equal(t, http.StatusPaymentRequired, rr.Code)
}

func TestHandler_Export_TeamScope_NoTeamID_400(t *testing.T) {
	svc := new(mockAnalyticsSvc)
	svc.On("ExportGate", mock.Anything, uint(1)).Return(nil)

	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.Export(rr, authedReq(t, http.MethodGet, "/api/analytics/export?format=csv&scope=team", 1))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Export_CSV_OK(t *testing.T) {
	svc := new(mockAnalyticsSvc)
	svc.On("ExportGate", mock.Anything, uint(1)).Return(nil)
	svc.On("GetPersonalDashboard", mock.Anything, uint(1), analyticsuc.Range7d).
		Return(&analyticsuc.PersonalDashboard{Range: analyticsuc.Range7d}, nil)

	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.Export(rr, authedReq(t, http.MethodGet, "/api/analytics/export?format=csv&scope=personal", 1))

	assert.Equal(t, http.StatusOK, rr.Code)
	ct := rr.Header().Get("Content-Type")
	assert.True(t, strings.Contains(ct, "csv") || strings.Contains(ct, "text/"),
		"ожидался CSV Content-Type, got: %s", ct)
}
