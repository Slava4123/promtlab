package subscription

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	subscriptionuc "promptvault/internal/usecases/subscription"
)

// respondError маппит доменные ошибки подписок в HTTP-статусы.
// Для 5xx (Internal, BadGateway при провале T-Bank) используется RespondWithRequest,
// чтобы Sentry захватывал ошибки с user context — сбои биллинга должны алертить.
func respondError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, subscriptionuc.ErrAlreadySubscribed):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	case errors.Is(err, subscriptionuc.ErrNoActiveSubscription):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, subscriptionuc.ErrPlanNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, subscriptionuc.ErrPaymentNotConfigured):
		httperr.Respond(w, httperr.New(http.StatusNotImplemented, err.Error(), nil))
	case errors.Is(err, subscriptionuc.ErrPaymentFailed):
		// 502 при провале T-Bank Init — критично, нужен алерт.
		httperr.RespondWithRequest(w, r, httperr.New(http.StatusBadGateway, err.Error(), err))
	case errors.Is(err, subscriptionuc.ErrInvalidWebhookSignature):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, subscriptionuc.ErrSubscriptionNotPausable),
		errors.Is(err, subscriptionuc.ErrSubscriptionPaused),
		errors.Is(err, subscriptionuc.ErrSubscriptionNotPaused),
		errors.Is(err, subscriptionuc.ErrInvalidPauseMonths),
		errors.Is(err, subscriptionuc.ErrInvalidCancelReason):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	default:
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
	}
}
