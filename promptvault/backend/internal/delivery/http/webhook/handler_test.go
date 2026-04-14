package webhook

import (
	"encoding/json"
	"testing"
)

// TestRawToSigValue покрывает критичный парсинг разнотипных полей T-Bank
// webhook'а в строковое представление, совместимое с расчётом подписи.
// Объекты и массивы ДОЛЖНЫ исключаться (keep=false), иначе подпись не сойдётся.
func TestRawToSigValue(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantVal  string
		wantKeep bool
	}{
		{"string", `"hello"`, "hello", true},
		{"empty string", `""`, "", true},
		{"bool true", `true`, "true", true},
		{"bool false", `false`, "false", true},
		{"positive int", `12345`, "12345", true},
		{"negative int", `-42`, "-42", true},
		{"zero", `0`, "0", true},
		{"decimal (Amount в рублях)", `19.99`, "19.99", true},
		{"null — пропускаем", `null`, "", false},
		{"object (Receipt) — пропускаем", `{"Email":"a@b.ru"}`, "", false},
		{"array — пропускаем", `[1,2,3]`, "", false},
		{"empty raw", ``, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, keep := rawToSigValue(json.RawMessage(tc.input))
			if keep != tc.wantKeep {
				t.Fatalf("keep = %v, want %v (input=%s)", keep, tc.wantKeep, tc.input)
			}
			if got != tc.wantVal {
				t.Fatalf("val = %q, want %q", got, tc.wantVal)
			}
		})
	}
}
