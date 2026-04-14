package webhook

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	subscriptionuc "promptvault/internal/usecases/subscription"
)

// Handler — HTTP-транспорт для обработки webhook-уведомлений от платёжных провайдеров.
type Handler struct {
	svc *subscriptionuc.Service
}

// NewHandler создаёт handler webhook'ов.
func NewHandler(svc *subscriptionuc.Service) *Handler {
	return &Handler{svc: svc}
}

// TBank — POST /api/webhooks/tbank.
// Обрабатывает webhook-уведомление от T-Bank. Тело запроса — JSON с полями
// OrderId, Status, Amount, PaymentId, Token и др. T-Bank ожидает ответ 200 OK.
func (h *Handler) TBank(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	// T-Bank отправляет JSON с полями разных типов: строки (OrderId, Status),
	// числа (Amount, ErrorCode), bool (Success), а также вложенные объекты
	// (Receipt, DATA). Для корректной проверки подписи нужно представление,
	// совпадающее с тем, что T-Bank использует при расчёте: bool → "true"/"false",
	// числа → десятичное представление. Вложенные объекты/массивы исключаются —
	// T-Bank не включает их в расчёт подписи.
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		slog.Error("webhook.tbank.decode_failed", "error", err)
		httperr.Respond(w, httperr.BadRequest("невалидный JSON"))
		return
	}

	params := make(map[string]string, len(raw))
	for k, v := range raw {
		if s, ok := rawToSigValue(v); ok {
			params[k] = s
		}
		// Объекты/массивы пропускаем — они не участвуют в подписи.
	}

	if err := h.svc.HandleWebhook(r.Context(), "tbank", params); err != nil {
		slog.Warn("webhook.tbank.handle_failed", "error", err)
		// T-Bank будет ретраить non-200 ответы, поэтому для невалидной подписи
		// возвращаем 400 — T-Bank не должен повторять заведомо неверный webhook.
		if errors.Is(err, subscriptionuc.ErrInvalidWebhookSignature) {
			httperr.Respond(w, httperr.BadRequest(err.Error()))
			return
		}
		// RespondWithRequest захватывает 5xx в Sentry — критично для монетизации:
		// без этого сбои активации подписки проходят незамеченными.
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
		return
	}

	// T-Bank ожидает 200 OK при успешной обработке
	utils.WriteOK(w, map[string]string{"status": "OK"})
}

// rawToSigValue конвертирует RawMessage в строковое представление, совместимое
// с алгоритмом подписи T-Bank. Возвращает (value, true) для скаляров;
// (``, false) для объектов/массивов/null — такие поля исключаются из подписи.
func rawToSigValue(v json.RawMessage) (string, bool) {
	if len(v) == 0 {
		return "", false
	}
	// Строка: "..." → убираем кавычки через json.Unmarshal.
	if v[0] == '"' {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			return s, true
		}
		return "", false
	}
	// Bool: true/false → "true"/"false".
	if v[0] == 't' || v[0] == 'f' {
		var b bool
		if err := json.Unmarshal(v, &b); err == nil {
			return strconv.FormatBool(b), true
		}
		return "", false
	}
	// Число: десятичная запись без форматирования.
	if v[0] == '-' || (v[0] >= '0' && v[0] <= '9') {
		return string(v), true
	}
	// null / object / array — пропускаем.
	return "", false
}
