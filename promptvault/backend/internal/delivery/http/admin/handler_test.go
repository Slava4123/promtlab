package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	adminuc "promptvault/internal/usecases/admin"
	adminauthuc "promptvault/internal/usecases/adminauth"
	auditsvc "promptvault/internal/usecases/audit"
	badgeuc "promptvault/internal/usecases/badge"
)

// ==================== fakes ====================

type fakeAdminSvc struct {
	listResult  *adminuc.UserListResult
	detail      *repo.UserDetail
	freezeErr   error
	unfreezeErr error
	resetErr    error
	grantBadge  *badgeuc.Badge
	grantErr    error
	revokeErr   error
	tierErr     error

	freezeCalls  int
	resetCalls   int
	grantCalls   int
	revokeCalls  int
}

func (f *fakeAdminSvc) ListUsers(_ context.Context, _ adminuc.UserListFilter) (*adminuc.UserListResult, error) {
	if f.listResult != nil {
		return f.listResult, nil
	}
	return &adminuc.UserListResult{Items: []repo.UserSummary{}, Page: 1, PageSize: 20}, nil
}
func (f *fakeAdminSvc) GetUserDetail(_ context.Context, _ uint) (*repo.UserDetail, error) {
	if f.detail == nil {
		return nil, adminuc.ErrUserNotFound
	}
	return f.detail, nil
}
func (f *fakeAdminSvc) FreezeUser(_ context.Context, _ uint) error {
	f.freezeCalls++
	return f.freezeErr
}
func (f *fakeAdminSvc) UnfreezeUser(_ context.Context, _ uint) error {
	return f.unfreezeErr
}
func (f *fakeAdminSvc) ResetPassword(_ context.Context, _ uint) error {
	f.resetCalls++
	return f.resetErr
}
func (f *fakeAdminSvc) GrantBadge(_ context.Context, _ uint, _ string) (*badgeuc.Badge, error) {
	f.grantCalls++
	if f.grantErr != nil {
		return nil, f.grantErr
	}
	return f.grantBadge, nil
}
func (f *fakeAdminSvc) RevokeBadge(_ context.Context, _ uint, _ string) error {
	f.revokeCalls++
	return f.revokeErr
}
func (f *fakeAdminSvc) ChangeTier(_ context.Context, _ uint, _ string) error {
	return f.tierErr
}

type fakeTOTPVerifier struct {
	valid bool
}

func (f *fakeTOTPVerifier) Verify(_ context.Context, _ uint, _ string) (*adminauthuc.VerifyResult, error) {
	if !f.valid {
		return nil, adminauthuc.ErrInvalidCode
	}
	return &adminauthuc.VerifyResult{UsedBackupCode: false, RemainingBackupCodes: 10}, nil
}

type fakeAuditReader struct {
	entries []models.AuditLog
}

func (f *fakeAuditReader) List(_ context.Context, _ repo.AuditLogFilter) ([]models.AuditLog, int64, error) {
	return f.entries, int64(len(f.entries)), nil
}

type fakeHealth struct {
	total, admins, active, frozen int64
}

func (f *fakeHealth) CountUsers(_ context.Context) (int64, int64, int64, int64, error) {
	return f.total, f.admins, f.active, f.frozen, nil
}

// ==================== helpers ====================

func setupHandler(svc AdminService, totp TOTPVerifier, audit AuditReader, health HealthCounter) *Handler {
	return NewHandler(svc, totp, audit, health)
}

// buildRequest — вспомогательная обёртка для chi URL params + admin context.
func buildRequest(method, url string, body any, params map[string]string) *http.Request {
	var r *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		r = httptest.NewRequest(method, url, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	// Chi URL params.
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	// Auth ctx (как authmw.Middleware).
	ctx = context.WithValue(ctx, authmw.UserIDKey, uint(1))
	// Admin audit ctx.
	ctx = auditsvc.WithContext(ctx, auditsvc.AdminRequestInfo{AdminID: 1, IP: "127.0.0.1", UserAgent: "test"})
	return r.WithContext(ctx)
}

// ==================== tests ====================

func TestListUsers_Empty(t *testing.T) {
	svc := &fakeAdminSvc{}
	h := setupHandler(svc, nil, nil, nil)

	r := buildRequest(http.MethodGet, "/api/admin/users", nil, nil)
	w := httptest.NewRecorder()
	h.ListUsers(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, float64(0), resp["total"])
	items, ok := resp["items"].([]any)
	require.True(t, ok)
	assert.Empty(t, items)
}

func TestGetUserDetail_NotFound(t *testing.T) {
	svc := &fakeAdminSvc{detail: nil}
	h := setupHandler(svc, nil, nil, nil)

	r := buildRequest(http.MethodGet, "/api/admin/users/99", nil, map[string]string{"id": "99"})
	w := httptest.NewRecorder()
	h.GetUserDetail(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserDetail_Success(t *testing.T) {
	u := &models.User{ID: 2, Email: "x@y.z", Role: models.RoleUser, Status: models.StatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	svc := &fakeAdminSvc{detail: &repo.UserDetail{User: u, PromptCount: 5, BadgeCount: 2}}
	h := setupHandler(svc, nil, nil, nil)

	r := buildRequest(http.MethodGet, "/api/admin/users/2", nil, map[string]string{"id": "2"})
	w := httptest.NewRecorder()
	h.GetUserDetail(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp UserDetailResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, uint(2), resp.ID)
	assert.Equal(t, "x@y.z", resp.Email)
	assert.Equal(t, int64(5), resp.PromptCount)
	assert.Equal(t, "free", resp.Tier)
}

func TestFreezeUser_Success(t *testing.T) {
	svc := &fakeAdminSvc{}
	h := setupHandler(svc, nil, nil, nil)

	r := buildRequest(http.MethodPost, "/api/admin/users/2/freeze", nil, map[string]string{"id": "2"})
	w := httptest.NewRecorder()
	h.FreezeUser(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, svc.freezeCalls)
}

func TestFreezeUser_DomainError(t *testing.T) {
	svc := &fakeAdminSvc{freezeErr: adminuc.ErrCannotFreezeSelf}
	h := setupHandler(svc, nil, nil, nil)

	r := buildRequest(http.MethodPost, "/api/admin/users/1/freeze", nil, map[string]string{"id": "1"})
	w := httptest.NewRecorder()
	h.FreezeUser(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== sudo-mode (TOTP) tests ====================

func TestResetPassword_MissingTOTPCode(t *testing.T) {
	svc := &fakeAdminSvc{}
	totp := &fakeTOTPVerifier{valid: true}
	h := setupHandler(svc, totp, nil, nil)

	r := buildRequest(http.MethodPost, "/api/admin/users/2/reset-password", nil, map[string]string{"id": "2"})
	w := httptest.NewRecorder()
	h.ResetPassword(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, svc.resetCalls, "service must not be called without TOTP")
}

func TestResetPassword_InvalidTOTPCode(t *testing.T) {
	svc := &fakeAdminSvc{}
	totp := &fakeTOTPVerifier{valid: false}
	h := setupHandler(svc, totp, nil, nil)

	body := map[string]string{"totp_code": "000000"}
	r := buildRequest(http.MethodPost, "/api/admin/users/2/reset-password", body, map[string]string{"id": "2"})
	w := httptest.NewRecorder()
	h.ResetPassword(w, r)

	// 422 Unprocessable Entity — бизнес-валидация TOTP, НЕ auth failure (401).
	// client.ts не будет ретритить 422, избегая двойного запроса (BUG #1).
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, 0, svc.resetCalls)
}

func TestResetPassword_ValidTOTP_CallsService(t *testing.T) {
	svc := &fakeAdminSvc{}
	totp := &fakeTOTPVerifier{valid: true}
	h := setupHandler(svc, totp, nil, nil)

	body := map[string]string{"totp_code": "123456"}
	r := buildRequest(http.MethodPost, "/api/admin/users/2/reset-password", body, map[string]string{"id": "2"})
	w := httptest.NewRecorder()
	h.ResetPassword(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, svc.resetCalls)
}

func TestGrantBadge_NoTOTPRequired(t *testing.T) {
	svc := &fakeAdminSvc{grantBadge: &badgeuc.Badge{ID: "first_prompt", Title: "Первопроходец", Icon: "🎯"}}
	h := setupHandler(svc, nil, nil, nil)

	r := buildRequest(http.MethodPost,
		"/api/admin/users/2/badges/first_prompt/grant", nil,
		map[string]string{"id": "2", "badge_id": "first_prompt"})
	w := httptest.NewRecorder()
	h.GrantBadge(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, svc.grantCalls)
}

func TestRevokeBadge_RequiresTOTP(t *testing.T) {
	svc := &fakeAdminSvc{}
	totp := &fakeTOTPVerifier{valid: true}
	h := setupHandler(svc, totp, nil, nil)

	body := map[string]string{"totp_code": "123456"}
	r := buildRequest(http.MethodDelete,
		"/api/admin/users/2/badges/first_prompt", body,
		map[string]string{"id": "2", "badge_id": "first_prompt"})
	w := httptest.NewRecorder()
	h.RevokeBadge(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, svc.revokeCalls)
}

// ==================== audit / health ====================

func TestListAudit_Empty(t *testing.T) {
	svc := &fakeAdminSvc{}
	audit := &fakeAuditReader{}
	h := setupHandler(svc, nil, audit, nil)

	r := buildRequest(http.MethodGet, "/api/admin/audit", nil, nil)
	w := httptest.NewRecorder()
	h.ListAudit(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealth(t *testing.T) {
	svc := &fakeAdminSvc{}
	health := &fakeHealth{total: 10, admins: 1, active: 9, frozen: 1}
	h := setupHandler(svc, nil, nil, health)

	r := buildRequest(http.MethodGet, "/api/admin/health", nil, nil)
	w := httptest.NewRecorder()
	h.Health(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp HealthResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, int64(10), resp.TotalUsers)
	assert.Equal(t, int64(1), resp.AdminUsers)
	assert.Equal(t, int64(9), resp.ActiveUsers)
}

// compile-time: helpers mentioned in imports.
var _ = strings.TrimSpace
var _ = errors.New
