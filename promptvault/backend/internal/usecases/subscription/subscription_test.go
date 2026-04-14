package subscription

import (
	"encoding/json"
	"strings"
	"testing"

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
	for i := 0; i < 100; i++ {
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
