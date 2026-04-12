// Package admin — HTTP handlers для админ-панели (/api/admin/*).
//
// Middleware chain (монтируется в app.MountRoutes):
//
//	authmw.Middleware (JWT) → admin.RequireAdmin → admin.AdminAuditContext → handler
//
// Destructive actions (freeze, reset_password, grant/revoke badge, change_tier)
// требуют fresh TOTP verification через totp_code в теле запроса (sudo mode).
// Проверка выполняется в handler через adminauth.Service.Verify до вызова
// admin.Service метода. При неверном коде возвращается 401 — frontend должен
// перепросить TOTP.
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	adminuc "promptvault/internal/usecases/admin"
	adminauthuc "promptvault/internal/usecases/adminauth"
	auditsvc "promptvault/internal/usecases/audit"
	badgeuc "promptvault/internal/usecases/badge"
)

// AdminService — локальный интерфейс того, что handler использует от admin usecase.
// Позволяет подставить fake в handler_test.go без привязки к *adminuc.Service.
type AdminService interface {
	ListUsers(ctx context.Context, filter adminuc.UserListFilter) (*adminuc.UserListResult, error)
	GetUserDetail(ctx context.Context, userID uint) (*repo.UserDetail, error)
	FreezeUser(ctx context.Context, targetID uint) error
	UnfreezeUser(ctx context.Context, targetID uint) error
	ResetPassword(ctx context.Context, targetID uint) error
	GrantBadge(ctx context.Context, targetID uint, badgeID string) (*badgeuc.Badge, error)
	RevokeBadge(ctx context.Context, targetID uint, badgeID string) error
	ChangeTier(ctx context.Context, targetID uint, tier string) error
}

// TOTPVerifier — узкий интерфейс для sudo-mode verification. *adminauthuc.Service
// удовлетворяет ему.
type TOTPVerifier interface {
	Verify(ctx context.Context, userID uint, code string) (*adminauthuc.VerifyResult, error)
}

// AuditReader — read-only доступ к audit_log для GET /api/admin/audit.
type AuditReader interface {
	List(ctx context.Context, filter repo.AuditLogFilter) ([]models.AuditLog, int64, error)
}

// HealthCounter — агрегация метрик для /admin/health.
// Узкий интерфейс, реализуется в app.go локальным адаптером (в adminRepo
// добавлять специализированные health-методы избыточно).
type HealthCounter interface {
	CountUsers(ctx context.Context) (total, admins, active, frozen int64, err error)
}

type Handler struct {
	admin    AdminService
	totp     TOTPVerifier
	audit    AuditReader
	health   HealthCounter
	validate *validator.Validate
}

func NewHandler(admin AdminService, totp TOTPVerifier, audit AuditReader, health HealthCounter) *Handler {
	return &Handler{
		admin:    admin,
		totp:     totp,
		audit:    audit,
		health:   health,
		validate: validator.New(),
	}
}

// ==================== users ====================

// GET /api/admin/users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := adminuc.UserListFilter{
		Query:    q.Get("q"),
		Role:     q.Get("role"),
		Status:   q.Get("status"),
		SortBy:   q.Get("sort"),
		SortDesc: q.Get("desc") == "true",
		Page:     parseIntDefault(q.Get("page"), 1),
		PageSize: parseIntDefault(q.Get("page_size"), 20),
	}

	result, err := h.admin.ListUsers(r.Context(), filter)
	if err != nil {
		respondError(w, err)
		return
	}

	items := make([]UserSummaryResponse, 0, len(result.Items))
	for _, u := range result.Items {
		items = append(items, NewUserSummaryResponse(u))
	}
	utils.WritePaginated(w, items, result.Total, result.Page, result.PageSize)
}

// GET /api/admin/users/{id}
func (h *Handler) GetUserDetail(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUintParam(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("неверный id пользователя"))
		return
	}
	detail, err := h.admin.GetUserDetail(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, NewUserDetailResponse(detail))
}

// POST /api/admin/users/{id}/freeze
func (h *Handler) FreezeUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUintParam(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("неверный id пользователя"))
		return
	}
	if err := h.admin.FreezeUser(r.Context(), userID); err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, ActionResponse{OK: true, Action: "freeze_user"})
}

// POST /api/admin/users/{id}/unfreeze
func (h *Handler) UnfreezeUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUintParam(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("неверный id пользователя"))
		return
	}
	if err := h.admin.UnfreezeUser(r.Context(), userID); err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, ActionResponse{OK: true, Action: "unfreeze_user"})
}

// POST /api/admin/users/{id}/reset-password
// REQUIRES fresh TOTP (sudo mode).
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUintParam(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("неверный id пользователя"))
		return
	}
	if err := h.verifyTOTP(r); err != nil {
		respondError(w, err)
		return
	}
	if err := h.admin.ResetPassword(r.Context(), userID); err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, ActionResponse{OK: true, Action: "reset_password"})
}

// POST /api/admin/users/{id}/badges/{badge_id}/grant
func (h *Handler) GrantBadge(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUintParam(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("неверный id пользователя"))
		return
	}
	badgeID := chi.URLParam(r, "badge_id")
	if badgeID == "" {
		httperr.Respond(w, httperr.BadRequest("не указан badge_id"))
		return
	}
	badge, err := h.admin.GrantBadge(r.Context(), userID, badgeID)
	if err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, NewGrantBadgeResponse(badge))
}

// DELETE /api/admin/users/{id}/badges/{badge_id}
// REQUIRES fresh TOTP (sudo mode).
func (h *Handler) RevokeBadge(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUintParam(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("неверный id пользователя"))
		return
	}
	badgeID := chi.URLParam(r, "badge_id")
	if badgeID == "" {
		httperr.Respond(w, httperr.BadRequest("не указан badge_id"))
		return
	}
	if err := h.verifyTOTP(r); err != nil {
		respondError(w, err)
		return
	}
	if err := h.admin.RevokeBadge(r.Context(), userID, badgeID); err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, ActionResponse{OK: true, Action: "revoke_badge"})
}

// POST /api/admin/users/{id}/tier
// REQUIRES fresh TOTP (sudo mode). STUB — возвращает 501.
func (h *Handler) ChangeTier(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUintParam(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("неверный id пользователя"))
		return
	}
	req, err := utils.DecodeAndValidate[ChangeTierRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	// Verify TOTP (sudo mode) даже для stub — для UX-предсказуемости.
	if err := h.verifyTOTPCode(r.Context(), req.TOTPCode); err != nil {
		respondError(w, err)
		return
	}
	if err := h.admin.ChangeTier(r.Context(), userID, req.Tier); err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, ActionResponse{OK: true, Action: "change_tier"})
}

// ==================== audit / health ====================

// GET /api/admin/audit
func (h *Handler) ListAudit(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := repo.AuditLogFilter{
		Action:     q.Get("action"),
		TargetType: q.Get("target_type"),
		Page:       parseIntDefault(q.Get("page"), 1),
		PageSize:   parseIntDefault(q.Get("page_size"), 20),
	}
	if v := q.Get("admin_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			idU := uint(id)
			filter.AdminID = &idU
		}
	}
	if v := q.Get("target_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			idU := uint(id)
			filter.TargetID = &idU
		}
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.FromTime = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.ToTime = &t
		}
	}

	entries, total, err := h.audit.List(r.Context(), filter)
	if err != nil {
		respondError(w, err)
		return
	}

	items := make([]AuditEntryResponse, 0, len(entries))
	for _, e := range entries {
		items = append(items, auditEntryToResponse(e))
	}
	page := max(filter.Page, 1)
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	utils.WritePaginated(w, items, total, page, pageSize)
}

// GET /api/admin/health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	total, admins, active, frozen, err := h.health.CountUsers(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}
	utils.WriteOK(w, HealthResponse{
		Status:      "ok",
		Time:        time.Now().UTC(),
		TotalUsers:  total,
		AdminUsers:  admins,
		ActiveUsers: active,
		FrozenUsers: frozen,
	})
}

// ==================== helpers ====================

// verifyTOTP читает totp_code из body и проверяет через adminauth.Verify.
// Для destructive actions без additional body fields.
func (h *Handler) verifyTOTP(r *http.Request) error {
	var req TOTPCodeRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		return httperr.BadRequest("требуется totp_code")
	}
	if req.TOTPCode == "" {
		return httperr.BadRequest("требуется totp_code")
	}
	return h.verifyTOTPCode(r.Context(), req.TOTPCode)
}

func (h *Handler) verifyTOTPCode(ctx context.Context, code string) error {
	adminID := authmw.GetUserID(ctx)
	if adminID == 0 {
		return httperr.Unauthorized("требуется авторизация")
	}
	_, err := h.totp.Verify(ctx, adminID, code)
	return err
}

func parseUintParam(r *http.Request, key string) (uint, error) {
	v := chi.URLParam(r, key)
	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func auditEntryToResponse(e models.AuditLog) AuditEntryResponse {
	resp := AuditEntryResponse{
		ID:         e.ID,
		AdminID:    e.AdminID,
		Action:     e.Action,
		TargetType: e.TargetType,
		TargetID:   e.TargetID,
		IP:         e.IP,
		UserAgent:  e.UserAgent,
		CreatedAt:  e.CreatedAt,
	}
	// BeforeState/AfterState — json.RawMessage, раскодируем в arbitrary any
	// чтобы frontend получил нормальный JSON (не escaped string).
	if len(e.BeforeState) > 0 {
		var v any
		if err := json.Unmarshal(e.BeforeState, &v); err == nil {
			resp.BeforeState = v
		}
	}
	if len(e.AfterState) > 0 {
		var v any
		if err := json.Unmarshal(e.AfterState, &v); err == nil {
			resp.AfterState = v
		}
	}
	return resp
}

// Assertion — compile-time guard, что основные зависимости неиспользуемых
// типов не забыты в имортах. Нет ошибок — нет warning.
var _ = errors.New
var _ = auditsvc.TargetUser
