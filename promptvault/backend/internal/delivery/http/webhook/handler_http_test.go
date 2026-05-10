// MN-2 — HTTP-level webhook handler tests: invalid signature → 400,
// malformed body → 400, valid → 200. Дополнение к существующему RawToSigValue.
package webhook

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	subscriptionuc "promptvault/internal/usecases/subscription"
)

// fakeService записывает call'ы и возвращает заданную ошибку.
type fakeService struct {
	called   bool
	provider string
	params   map[string]string
	err      error
}

func (f *fakeService) HandleWebhook(_ context.Context, provider string, params map[string]string) error {
	f.called = true
	f.provider = provider
	f.params = params
	return f.err
}

func newHandler(svc *fakeService) *Handler {
	return &Handler{svc: svc}
}

func TestTBank_MalformedJSON_400(t *testing.T) {
	svc := &fakeService{}
	h := newHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/tbank", bytes.NewReader([]byte("not-json")))
	rec := httptest.NewRecorder()
	h.TBank(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 на malformed JSON, got %d", rec.Code)
	}
	if svc.called {
		t.Error("svc.HandleWebhook не должен быть вызван при malformed body")
	}
}

func TestTBank_InvalidSignature_400_NoRetry(t *testing.T) {
	// MN-2: T-Bank ретраит non-200 ответы, но invalid signature — намеренно
	// 400, чтобы T-Bank НЕ ретраил (заведомо невалидный webhook).
	svc := &fakeService{err: subscriptionuc.ErrInvalidWebhookSignature}
	h := newHandler(svc)

	body := []byte(`{"OrderId":"abc","Status":"CONFIRMED","Token":"bad-sig"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/tbank", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.TBank(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 для invalid signature, got %d", rec.Code)
	}
	if !svc.called {
		t.Error("svc.HandleWebhook должен быть вызван (для проверки signature)")
	}
}

func TestTBank_InternalError_500_RetryFromTBank(t *testing.T) {
	// Любая non-signature ошибка → 500, T-Bank ретраит (e.g. DB connection drop).
	svc := &fakeService{err: errors.New("db down")}
	h := newHandler(svc)

	body := []byte(`{"OrderId":"abc","Status":"CONFIRMED"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/tbank", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.TBank(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 для transient error, got %d", rec.Code)
	}
}

func TestTBank_HappyPath_200(t *testing.T) {
	svc := &fakeService{}
	h := newHandler(svc)

	body := []byte(`{"OrderId":"abc","Status":"CONFIRMED","Amount":59900,"Token":"sig"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/tbank", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.TBank(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if svc.provider != "tbank" {
		t.Errorf("provider = %q, want tbank", svc.provider)
	}
	if svc.params["OrderId"] != "abc" {
		t.Errorf("OrderId не передан в params: %v", svc.params)
	}
	if svc.params["Status"] != "CONFIRMED" {
		t.Errorf("Status не передан: %v", svc.params)
	}
	// Amount — number → "59900"
	if svc.params["Amount"] != "59900" {
		t.Errorf("Amount expected '59900', got %q", svc.params["Amount"])
	}
}

func TestTBank_NestedReceipt_ExcludedFromSig(t *testing.T) {
	// Receipt — объект, не должен попасть в params (excluded from sig).
	svc := &fakeService{}
	h := newHandler(svc)

	body := []byte(`{"OrderId":"abc","Status":"CONFIRMED","Receipt":{"Email":"a@b.ru"},"Token":"x"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/tbank", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.TBank(rec, req)

	if _, ok := svc.params["Receipt"]; ok {
		t.Errorf("Receipt должен быть excluded из params (объекты пропускаются), got %v", svc.params)
	}
}
