package ipallowlist

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// newRequest строит запрос с заданным RemoteAddr и опциональным XFF.
func newRequest(t *testing.T, remoteAddr, xff string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/api/webhooks/tbank", nil)
	r.RemoteAddr = remoteAddr
	if xff != "" {
		r.Header.Set("X-Forwarded-For", xff)
	}
	return r
}

// ok — простой handler, возвращающий 200. Оборачиваем им middleware для
// проверки что запрос дошёл до внутреннего обработчика.
func ok() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

func TestIPAllowlist_EmptyList_PassesAll(t *testing.T) {
	// Пустой список — middleware no-op. Используется в dev/pre-prod когда
	// IP провайдера ещё не получены — сервис работает без ошибок, защиты нет.
	mw := New(nil, false)

	rec := httptest.NewRecorder()
	mw(ok()).ServeHTTP(rec, newRequest(t, "1.2.3.4:8080", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("пустой allowlist должен пропускать: got %d", rec.Code)
	}
}

func TestIPAllowlist_ExactIP(t *testing.T) {
	mw := New([]string{"212.233.80.7"}, false)

	cases := []struct {
		name       string
		remoteAddr string
		want       int
	}{
		{"разрешённый IP", "212.233.80.7:5000", http.StatusOK},
		{"другой IP", "1.2.3.4:5000", http.StatusForbidden},
		{"локальный IP", "127.0.0.1:5000", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mw(ok()).ServeHTTP(rec, newRequest(t, tc.remoteAddr, ""))
			if rec.Code != tc.want {
				t.Errorf("для %q got %d, want %d", tc.remoteAddr, rec.Code, tc.want)
			}
		})
	}
}

func TestIPAllowlist_CIDR(t *testing.T) {
	// Диапазон T-Bank 91.194.226.0/23 — 510 хостов от .226.1 до .227.254.
	mw := New([]string{"91.194.226.0/23"}, false)

	cases := []struct {
		name string
		ip   string
		want int
	}{
		{"начало диапазона", "91.194.226.1:0", http.StatusOK},
		{"середина диапазона", "91.194.226.128:0", http.StatusOK},
		{"второй подсети /23", "91.194.227.100:0", http.StatusOK},
		{"последний хост", "91.194.227.254:0", http.StatusOK},
		{"за пределами диапазона", "91.194.228.1:0", http.StatusForbidden},
		{"другая сеть", "10.0.0.1:0", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mw(ok()).ServeHTTP(rec, newRequest(t, tc.ip, ""))
			if rec.Code != tc.want {
				t.Errorf("для %q got %d, want %d", tc.ip, rec.Code, tc.want)
			}
		})
	}
}

func TestIPAllowlist_MixedCIDRAndIP(t *testing.T) {
	// Реальный набор T-Bank: два одиночных IP + один CIDR.
	mw := New([]string{"212.233.80.7", "91.218.132.2", "91.194.226.0/23"}, false)

	pass := []string{"212.233.80.7:1", "91.218.132.2:1", "91.194.226.50:1", "91.194.227.200:1"}
	deny := []string{"1.1.1.1:1", "91.218.132.3:1", "91.194.225.255:1"}

	for _, addr := range pass {
		rec := httptest.NewRecorder()
		mw(ok()).ServeHTTP(rec, newRequest(t, addr, ""))
		if rec.Code != http.StatusOK {
			t.Errorf("ожидался 200 для %q, got %d", addr, rec.Code)
		}
	}
	for _, addr := range deny {
		rec := httptest.NewRecorder()
		mw(ok()).ServeHTTP(rec, newRequest(t, addr, ""))
		if rec.Code != http.StatusForbidden {
			t.Errorf("ожидался 403 для %q, got %d", addr, rec.Code)
		}
	}
}

func TestIPAllowlist_TrustForwarded(t *testing.T) {
	// За nginx: r.RemoteAddr = IP контейнера nginx, реальный клиент в X-Forwarded-For.
	mw := New([]string{"212.233.80.7"}, true)

	// Запрос от nginx (172.x), реальный клиент T-Bank через XFF.
	rec := httptest.NewRecorder()
	mw(ok()).ServeHTTP(rec, newRequest(t, "172.20.0.2:5000", "212.233.80.7"))
	if rec.Code != http.StatusOK {
		t.Errorf("XFF с разрешённым IP должен пропускать, got %d", rec.Code)
	}

	// XFF с несколькими прокси — берётся первый.
	rec = httptest.NewRecorder()
	mw(ok()).ServeHTTP(rec, newRequest(t, "172.20.0.2:5000", "212.233.80.7, 10.0.0.1, 10.0.0.2"))
	if rec.Code != http.StatusOK {
		t.Errorf("XFF с несколькими IP должен брать первый, got %d", rec.Code)
	}

	// XFF с недопустимым клиентским IP — отказ, даже если nginx IP валидный.
	rec = httptest.NewRecorder()
	mw(ok()).ServeHTTP(rec, newRequest(t, "212.233.80.7:5000", "1.2.3.4"))
	if rec.Code != http.StatusForbidden {
		t.Errorf("XFF с чужим IP должен отказывать (nginx IP не считается), got %d", rec.Code)
	}
}

func TestIPAllowlist_NoTrustForwarded_IgnoresXFF(t *testing.T) {
	// TrustForwarded=false — XFF не читается, защита от spoofing'а XFF клиентом.
	mw := New([]string{"212.233.80.7"}, false)

	rec := httptest.NewRecorder()
	mw(ok()).ServeHTTP(rec, newRequest(t, "1.2.3.4:5000", "212.233.80.7"))
	if rec.Code != http.StatusForbidden {
		t.Errorf("XFF должен игнорироваться при trustForwarded=false, got %d", rec.Code)
	}
}

func TestIPAllowlist_InvalidEntries_LoggedButOthersWork(t *testing.T) {
	// Часть записей невалидна — валидные всё равно работают.
	mw := New([]string{"not-an-ip", "999.999.999.999", "bogus/cidr", "212.233.80.7"}, false)

	rec := httptest.NewRecorder()
	mw(ok()).ServeHTTP(rec, newRequest(t, "212.233.80.7:5000", ""))
	if rec.Code != http.StatusOK {
		t.Errorf("валидный IP должен проходить даже если другие записи битые, got %d", rec.Code)
	}
}

func TestIPAllowlist_UnparsableRemoteAddr_Denied(t *testing.T) {
	mw := New([]string{"212.233.80.7"}, false)

	rec := httptest.NewRecorder()
	mw(ok()).ServeHTTP(rec, newRequest(t, "garbage", ""))
	if rec.Code != http.StatusForbidden {
		t.Errorf("битый RemoteAddr должен давать 403, got %d", rec.Code)
	}
}
