package http_test

// Q-8: lint-test проверяющий что все доменные ошибки (Err* из usecases)
// упомянуты хотя бы в одном файле errors.go в соответствующем delivery/http пакете.
//
// Не заменяет полноценный handler-test, но ловит наиболее частый баг:
// добавили новую sentinel-ошибку в usecase, забыли пометить HTTP-статус —
// респондер проваливается в default (500) вместо правильного 4xx.
//
// Пара <usecase, delivery> задана явно в usecaseToDelivery ниже. Если будет
// несовпадение (ошибка определена, но нигде не замаплена), тест скажет где.

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// usecaseToDelivery — map usecase package → delivery http package. Ключ
// относительно backend/internal/usecases/, значение — delivery/http/.
// Пары добавляются только для usecases у которых есть own HTTP-handler;
// чисто-внутренние сервисы (email, payment, streak internal etc) не входят.
var usecaseToDelivery = map[string]string{
	"auth":         "auth",
	"admin":        "admin",
	"adminauth":    "adminauth",
	"ai":           "ai",
	"apikey":       "apikey",
	"badge":        "badge",
	"changelog":    "changelog",
	"collection":   "collection",
	"feedback":     "feedback",
	"prompt":       "prompt",
	"search":       "search",
	"share":        "share",
	"starter":      "starter",
	"streak":       "streak",
	"subscription": "subscription",
	"tag":          "tag",
	"team":         "team",
	"trash":        "trash",
	"user":         "user",
}

// allowUnmapped — ошибки, которые умышленно НЕ возвращаются из handler'ов
// (внутренние sentinel'ы для flow-control между слоями). Добавлять сюда
// только если честно проверили: ошибка никогда не bubble'ится до HTTP.
var allowUnmapped = map[string]bool{
	"ErrInvalidWebhookSignature": true, // subscription: используется как sentinel в webhook, маппится отдельно
}

// errDecl — "var ErrFoo = errors.New(...)" или "ErrBar = errors.New(...)" в var(...).
var errDecl = regexp.MustCompile(`(?m)^\s*(Err[A-Z][A-Za-z0-9]*)\s*=\s*`)

func TestErrorMappingCoverage(t *testing.T) {
	backendRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs backend root: %v", err)
	}
	ucRoot := filepath.Join(backendRoot, "internal", "usecases")
	httpRoot := filepath.Join(backendRoot, "internal", "delivery", "http")

	var missing []string
	for ucPkg, httpPkg := range usecaseToDelivery {
		errsFile := filepath.Join(ucRoot, ucPkg, "errors.go")
		ucErrs, err := collectErrDecls(errsFile)
		if err != nil {
			// Нет errors.go — пакет может не иметь доменных ошибок, это ок.
			continue
		}

		mapped, err := collectMappedErrors(filepath.Join(httpRoot, httpPkg))
		if err != nil {
			t.Fatalf("read %s: %v", httpPkg, err)
		}

		for _, name := range ucErrs {
			if allowUnmapped[name] {
				continue
			}
			if !mapped[name] {
				missing = append(missing, "usecases/"+ucPkg+"."+name+" not referenced in delivery/http/"+httpPkg)
			}
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf("unmapped domain errors (%d):\n  %s", len(missing), strings.Join(missing, "\n  "))
	}
}

func collectErrDecls(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	matches := errDecl.FindAllStringSubmatch(string(data), -1)
	out := make([]string, 0, len(matches))
	seen := map[string]bool{}
	for _, m := range matches {
		if seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		out = append(out, m[1])
	}
	return out, nil
}

func collectMappedErrors(httpPkgDir string) (map[string]bool, error) {
	out := map[string]bool{}
	entries, err := os.ReadDir(httpPkgDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(httpPkgDir, e.Name()))
		if err != nil {
			return nil, err
		}
		// Отслеживаем любые упоминания "ErrFoo" в коде delivery — в errors.Is,
		// switch-case, comments. Достаточно для coverage: если имя встречается,
		// программист уже задумался о mapping'е.
		for _, match := range errRef.FindAllStringSubmatch(string(data), -1) {
			out[match[1]] = true
		}
	}
	return out, nil
}

var errRef = regexp.MustCompile(`\b(Err[A-Z][A-Za-z0-9]*)\b`)
