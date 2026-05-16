// Регрессионный тест: Init передаёт FailURL и в signature, и в body.
//
// Раньше FailURL намеренно опускался (с комментарием «T-Bank покажет свой
// экран ошибки») — но это плохой UX retention: юзер видит generic
// «Повторить попытку», провал тот же → отвал. После фикса
// /settings/subscription?payment=failure показывает наш экран с кнопкой
// «Обновить способ оплаты».

package tbank

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"promptvault/internal/infrastructure/payment"
)

func TestInit_PassesFailURLInBodyAndToken(t *testing.T) {
	var capturedBody initRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/Init") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(raw, &capturedBody); err != nil {
			t.Fatalf("unmarshal: %v\nraw=%s", err, string(raw))
		}
		// Минимальный валидный ответ T-Bank.
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Success":true,"ErrorCode":"0","PaymentId":"42","PaymentURL":"https://pay.test/x","Status":"NEW"}`))
	}))
	defer srv.Close()

	p := NewProvider(Config{
		TerminalKey: "TEST_TERM",
		Password:    "TEST_PASS",
		BaseURL:     srv.URL,
	})

	const failURL = "https://app.example.com/settings/subscription?payment=failure"
	_, err := p.Init(context.Background(), payment.InitRequest{
		OrderID:     "order-1",
		Amount:      59900,
		Description: "Test sub",
		SuccessURL:  "https://app.example.com/settings/subscription?payment=success",
		FailURL:     failURL,
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// FailURL должен попасть в JSON body — без него T-Bank редиректит на свой экран.
	if capturedBody.FailURL != failURL {
		t.Errorf("body.FailURL = %q, want %q", capturedBody.FailURL, failURL)
	}

	// Token должен учитывать FailURL: пересоберём ожидаемый Token и сравним.
	// Если FailURL пропущен в tokenParams (как было до фикса), хеш будет другим
	// и T-Bank вернул бы код 9 (invalid token). Эта проверка регрессит
	// именно signing, а не только тело запроса.
	wantToken := p.generateToken(map[string]string{
		"TerminalKey": "TEST_TERM",
		"Amount":      "59900",
		"OrderId":     "order-1",
		"Description": "Test sub",
		"SuccessURL":  "https://app.example.com/settings/subscription?payment=success",
		"FailURL":     failURL,
	})
	if capturedBody.Token != wantToken {
		t.Errorf("body.Token mismatch:\n got  %s\n want %s\n(если расходится — FailURL не участвует в Token signing)", capturedBody.Token, wantToken)
	}
}

func TestInit_OmitsFailURLWhenEmpty(t *testing.T) {
	// Defensive: пустой FailURL не должен влиять на Token (omitempty в body,
	// и не должен попадать в tokenParams). Гарантирует обратную совместимость
	// с тестами/инстансами где FailURL не сконфигурирован.
	var capturedBody initRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &capturedBody)
		_, _ = w.Write([]byte(`{"Success":true,"ErrorCode":"0","PaymentId":"42","PaymentURL":"x","Status":"NEW"}`))
	}))
	defer srv.Close()

	p := NewProvider(Config{TerminalKey: "T", Password: "P", BaseURL: srv.URL})
	_, err := p.Init(context.Background(), payment.InitRequest{
		OrderID:     "order-2",
		Amount:      100,
		Description: "X",
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if capturedBody.FailURL != "" {
		t.Errorf("body.FailURL must be empty when not provided, got %q", capturedBody.FailURL)
	}

	wantToken := p.generateToken(map[string]string{
		"TerminalKey": "T",
		"Amount":      "100",
		"OrderId":     "order-2",
		"Description": "X",
	})
	if capturedBody.Token != wantToken {
		t.Errorf("body.Token mismatch (FailURL пустой не должен влиять):\n got  %s\n want %s", capturedBody.Token, wantToken)
	}
}
