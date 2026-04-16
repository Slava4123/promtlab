package subscription

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	quotauc "promptvault/internal/usecases/quota"
	subscriptionuc "promptvault/internal/usecases/subscription"
)

// Handler — HTTP-транспорт для подписок и тарифов.
type Handler struct {
	svc      *subscriptionuc.Service
	quotas   *quotauc.Service
	validate *validator.Validate
}

// NewHandler создаёт handler подписок.
func NewHandler(svc *subscriptionuc.Service, quotas *quotauc.Service) *Handler {
	return &Handler{
		svc:      svc,
		quotas:   quotas,
		validate: validator.New(),
	}
}

// ListPlans — GET /api/subscription/plans.
// Возвращает список активных тарифных планов. Публичный endpoint.
func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.svc.GetPlans(r.Context())
	if err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteOK(w, NewPlansResponse(plans))
}

// GetSubscription — GET /api/subscription.
// Возвращает текущую активную подписку пользователя.
func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	sub, err := h.svc.GetSubscription(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteOK(w, NewSubscriptionResponse(sub))
}

// GetUsage — GET /api/subscription/usage.
// Возвращает сводку использования vs лимитов для текущего юзера.
func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	summary, err := h.quotas.GetUsageSummary(r.Context(), userID)
	if err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteOK(w, NewUsageResponse(summary))
}

// Checkout — POST /api/subscription/checkout.
// Инициализирует платёж и возвращает URL для оплаты.
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[CheckoutRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.Checkout(r.Context(), subscriptionuc.CheckoutInput{
		UserID: userID,
		PlanID: req.PlanID,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteOK(w, CheckoutResponse{PaymentURL: result.PaymentURL})
}

// Cancel — POST /api/subscription/cancel.
// Помечает подписку для отмены в конце текущего периода. Принимает опциональную
// причину отмены для exit-survey (M-6b); reason и other_text игнорируются, если
// body пустой (backward compat со старыми клиентами).
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	// Body опционален: пустой запрос = cancel без причины. Игнорируем ошибку
	// парсинга, только если body действительно пустой (Content-Length==0).
	req := CancelRequest{}
	if r.ContentLength > 0 {
		parsed, err := utils.DecodeAndValidate[CancelRequest](r, h.validate)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest(err.Error()))
			return
		}
		req = parsed
	}

	if err := h.svc.Cancel(r.Context(), subscriptionuc.CancelInput{
		UserID: userID,
		Reason: req.Reason,
		Other:  req.OtherText,
	}); err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteNoContent(w)
}

// Pause — POST /api/subscription/pause. M-6. Body: {months: 1|2|3}.
func (h *Handler) Pause(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[PauseRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.svc.Pause(r.Context(), subscriptionuc.PauseInput{
		UserID: userID,
		Months: req.Months,
	}); err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteNoContent(w)
}

// Resume — POST /api/subscription/resume. M-6. Досрочное возобновление.
func (h *Handler) Resume(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	if err := h.svc.Resume(r.Context(), userID); err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteNoContent(w)
}

// Downgrade — POST /api/subscription/downgrade.
// Немедленно переводит на Free план, отменяя активную подписку.
func (h *Handler) Downgrade(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	if err := h.svc.Downgrade(r.Context(), userID); err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteNoContent(w)
}

// SetAutoRenew — POST /api/subscription/auto-renew.
// Включает/выключает автопродление подписки. При false подписка истечёт
// в конце текущего периода без попытки списания (renewLoop её пропустит).
func (h *Handler) SetAutoRenew(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[AutoRenewRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.svc.SetAutoRenew(r.Context(), userID, *req.AutoRenew); err != nil {
		respondError(w, r, err)
		return
	}

	utils.WriteNoContent(w)
}
