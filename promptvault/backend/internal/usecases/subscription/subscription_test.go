package subscription

import (
	"encoding/json"
	"strings"
	"testing"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/models"
)

// TestExtractPlanID покрывает ключевой A4-фикс: активация подписки идёт строго
// по plan_id из ProviderData, а не по сумме платежа (чтобы совпадение цен
// у разных планов — например, промо-Max за цену Pro — не приводило к выдаче
// не того плана).
func TestExtractPlanID(t *testing.T) {
	cases := []struct {
		name       string
		pay        *models.Payment
		wantPlanID string
		wantErr    string // substring в error message; "" = ошибки быть не должно
	}{
		{
			name:       "ok — plan_id в provider_data",
			pay:        &models.Payment{ID: 1, ProviderData: mustJSON(map[string]string{"plan_id": "pro"})},
			wantPlanID: "pro",
		},
		{
			name:       "ok — max план",
			pay:        &models.Payment{ID: 2, ProviderData: mustJSON(map[string]string{"plan_id": "max"})},
			wantPlanID: "max",
		},
		{
			name:    "пустой ProviderData — ошибка",
			pay:     &models.Payment{ID: 3, ProviderData: nil},
			wantErr: "provider_data пуст",
		},
		{
			name:    "невалидный JSON — ошибка",
			pay:     &models.Payment{ID: 4, ProviderData: json.RawMessage("not-json")},
			wantErr: "unmarshal",
		},
		{
			name:    "ProviderData без plan_id — ошибка",
			pay:     &models.Payment{ID: 5, ProviderData: mustJSON(map[string]string{"other": "value"})},
			wantErr: "plan_id отсутствует",
		},
		{
			name:    "plan_id пустая строка — ошибка",
			pay:     &models.Payment{ID: 6, ProviderData: mustJSON(map[string]string{"plan_id": ""})},
			wantErr: "plan_id отсутствует",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractPlanID(tc.pay)

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("ожидалась ошибка содержащая %q, получено nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %q, want substring %q", err.Error(), tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if got != tc.wantPlanID {
				t.Fatalf("plan_id = %q, want %q", got, tc.wantPlanID)
			}
		})
	}
}

func TestGenerateIdempotencyKey(t *testing.T) {
	// Ключ должен быть уникальным между вызовами и правильной длины (32 hex).
	seen := make(map[string]struct{}, 100)
	for i := range 100 {
		k, err := generateIdempotencyKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(k) != 32 {
			t.Fatalf("длина ключа = %d, want 32 (16 байт в hex)", len(k))
		}
		if _, dup := seen[k]; dup {
			t.Fatalf("коллизия ключей после %d итераций: %s", i, k)
		}
		seen[k] = struct{}{}
	}
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// TestBuildReceipt проверяет формирование фискального чека 54-ФЗ для T-Bank.
// Чек должен содержать Email (54-ФЗ требует email или телефон), Taxation
// (система налогообложения из конфига) и единственную позицию с полной
// суммой подписки без НДС (для УСН).
func TestBuildReceipt(t *testing.T) {
	plan := &models.SubscriptionPlan{
		ID:       "pro",
		Name:     "Pro",
		PriceKop: 59900, // 599 ₽
	}

	t.Run("nil config — возвращает nil", func(t *testing.T) {
		if r := buildReceipt(nil, "user@example.com", plan); r != nil {
			t.Fatalf("ожидался nil при nil config, получено: %+v", r)
		}
	})

	t.Run("ReceiptEnabled=false — возвращает nil", func(t *testing.T) {
		cfg := &config.PaymentConfig{ReceiptEnabled: false, Taxation: "usn_income"}
		if r := buildReceipt(cfg, "user@example.com", plan); r != nil {
			t.Fatalf("ожидался nil при ReceiptEnabled=false, получено: %+v", r)
		}
	})

	t.Run("ReceiptEnabled=true — формирует Receipt для УСН 6%", func(t *testing.T) {
		cfg := &config.PaymentConfig{ReceiptEnabled: true, Taxation: "usn_income"}
		email := "user@example.com"
		r := buildReceipt(cfg, email, plan)

		if r == nil {
			t.Fatal("ожидался Receipt, получено nil")
		}
		if r.Email != email {
			t.Errorf("Email = %q, want %q", r.Email, email)
		}
		if r.Taxation != "usn_income" {
			t.Errorf("Taxation = %q, want %q", r.Taxation, "usn_income")
		}
		if len(r.Items) != 1 {
			t.Fatalf("len(Items) = %d, want 1", len(r.Items))
		}

		item := r.Items[0]
		if item.PriceKop != plan.PriceKop {
			t.Errorf("Price = %d, want %d", item.PriceKop, plan.PriceKop)
		}
		if item.AmountKop != plan.PriceKop {
			t.Errorf("Amount = %d, want %d (Price * Quantity для 1 шт)", item.AmountKop, plan.PriceKop)
		}
		if item.Quantity != 1 {
			t.Errorf("Quantity = %d, want 1", item.Quantity)
		}
		if item.Tax != "none" {
			t.Errorf("Tax = %q, want %q (УСН без НДС)", item.Tax, "none")
		}
		if item.PaymentMethod != "full_payment" {
			t.Errorf("PaymentMethod = %q, want %q", item.PaymentMethod, "full_payment")
		}
		if item.PaymentObject != "service" {
			t.Errorf("PaymentObject = %q, want %q (подписка = услуга)", item.PaymentObject, "service")
		}
		if !strings.Contains(item.Name, plan.Name) {
			t.Errorf("Name = %q, должен содержать имя плана %q", item.Name, plan.Name)
		}
	})

	t.Run("Max план с другой ценой", func(t *testing.T) {
		cfg := &config.PaymentConfig{ReceiptEnabled: true, Taxation: "usn_income"}
		maxPlan := &models.SubscriptionPlan{ID: "max", Name: "Max", PriceKop: 129900}
		r := buildReceipt(cfg, "user@example.com", maxPlan)

		if r == nil {
			t.Fatal("ожидался Receipt")
		}
		if r.Items[0].AmountKop != 129900 {
			t.Errorf("Amount = %d, want 129900 для Max", r.Items[0].AmountKop)
		}
	})

	t.Run("разные СНО передаются как есть", func(t *testing.T) {
		taxations := []string{"usn_income", "usn_income_outcome", "osn", "patent", "esn"}
		for _, tax := range taxations {
			cfg := &config.PaymentConfig{ReceiptEnabled: true, Taxation: tax}
			r := buildReceipt(cfg, "user@example.com", plan)
			if r.Taxation != tax {
				t.Errorf("Taxation для %q = %q, want %q", tax, r.Taxation, tax)
			}
		}
	})

	t.Run("пустой email — всё равно формирует Receipt (нужно email ИЛИ phone)", func(t *testing.T) {
		// buildReceipt не валидирует наличие email — это ответственность caller'а
		// (Checkout читает user.Email из БД и подставляет). Если email пустой,
		// T-Bank может либо использовать phone (если передан), либо отклонить.
		cfg := &config.PaymentConfig{ReceiptEnabled: true, Taxation: "usn_income"}
		r := buildReceipt(cfg, "", plan)
		if r == nil {
			t.Fatal("buildReceipt должен возвращать Receipt даже с пустым email")
		}
		if r.Email != "" {
			t.Errorf("Email = %q, want empty", r.Email)
		}
	})
}
