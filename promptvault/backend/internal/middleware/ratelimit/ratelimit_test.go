// MN-5 — sliding window logic + ByUserID/ByIP middleware tests.
package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// --- Limiter core ---

func TestLimiter_Allow_BelowLimit_True(t *testing.T) {
	l := NewLimiterWithWindow[uint](3, time.Minute, UintHash)
	defer l.Close()
	for i := 0; i < 3; i++ {
		if !l.Allow(42) {
			t.Fatalf("expected allow request %d (limit=3)", i+1)
		}
	}
}

func TestLimiter_Allow_AtLimit_Reject(t *testing.T) {
	l := NewLimiterWithWindow[uint](2, time.Minute, UintHash)
	defer l.Close()
	if !l.Allow(42) {
		t.Fatal("первый запрос должен проходить")
	}
	if !l.Allow(42) {
		t.Fatal("второй запрос должен проходить")
	}
	if l.Allow(42) {
		t.Fatal("3-й запрос при limit=2 должен быть отвергнут")
	}
}

func TestLimiter_Allow_DifferentKeys_Independent(t *testing.T) {
	l := NewLimiterWithWindow[uint](1, time.Minute, UintHash)
	defer l.Close()
	if !l.Allow(1) || !l.Allow(2) || !l.Allow(3) {
		t.Fatal("разные userID имеют независимые buckets")
	}
	if l.Allow(1) {
		t.Fatal("повторный для user 1 должен быть отвергнут (limit=1)")
	}
}

func TestLimiter_Allow_AfterWindow_Reset(t *testing.T) {
	// Короткое окно 50ms — после него bucket пустеет.
	l := NewLimiterWithWindow[uint](1, 50*time.Millisecond, UintHash)
	defer l.Close()
	if !l.Allow(42) {
		t.Fatal("первый запрос должен пройти")
	}
	if l.Allow(42) {
		t.Fatal("второй сразу — отвергнут")
	}
	time.Sleep(60 * time.Millisecond)
	if !l.Allow(42) {
		t.Fatal("после window — должно стать allow=true")
	}
}

func TestLimiter_Allow_ZeroLimit_AlwaysAllow(t *testing.T) {
	// Zero limit = "не ограничивать" (по контракту в Allow).
	l := NewLimiterWithWindow[uint](0, time.Minute, UintHash)
	defer l.Close()
	for i := 0; i < 100; i++ {
		if !l.Allow(42) {
			t.Fatalf("limit=0 должен пропускать всё (запрос %d)", i+1)
		}
	}
}

// --- ByUserID middleware ---

func TestByUserID_NoUser_401(t *testing.T) {
	mw := ByUserIDWithWindow(10, time.Minute, func(_ *http.Request) uint { return 0 })
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler не должен быть вызван при userID=0")
	}))

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestByUserID_OverLimit_429(t *testing.T) {
	uid := uint(42)
	mw := ByUserIDWithWindow(2, time.Minute, func(_ *http.Request) uint { return uid })
	called := 0
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		handler.ServeHTTP(rec, req)
		if i < 2 && rec.Code != http.StatusOK {
			t.Errorf("запрос %d: ожидался 200, got %d", i+1, rec.Code)
		}
		if i == 2 {
			if rec.Code != http.StatusTooManyRequests {
				t.Errorf("3-й запрос: ожидался 429, got %d", rec.Code)
			}
			if rec.Header().Get("Retry-After") == "" {
				t.Error("expected Retry-After header в 429")
			}
		}
	}
	if called != 2 {
		t.Errorf("handler вызван %d раз, ожидалось 2 (3-й заблокирован)", called)
	}
}

// --- ByIP middleware ---

func TestByIP_OverLimit_429(t *testing.T) {
	mw := ByIP(1, false)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req1.RemoteAddr = "1.2.3.4:5000"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("первый запрос: 200, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req2.RemoteAddr = "1.2.3.4:5001"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("второй: 429, got %d", rec2.Code)
	}
}

func TestByIP_DifferentIPs_Independent(t *testing.T) {
	mw := ByIP(1, false)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	for _, ip := range []string{"1.1.1.1:1", "2.2.2.2:1", "3.3.3.3:1"} {
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		req.RemoteAddr = ip
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("IP %q: первый запрос — 200, got %d", ip, rec.Code)
		}
	}
}

// --- clientIP / trustProxy ---

func TestClientIP_TrustProxy_PrefersXFF(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.RemoteAddr = "1.1.1.1:5000"
	req.Header.Set("X-Forwarded-For", "8.8.8.8, 9.9.9.9")
	if got := clientIP(req, true); got != "8.8.8.8" {
		t.Errorf("trustProxy=true: ожидался 8.8.8.8 (первый XFF), got %q", got)
	}
}

func TestClientIP_NoTrustProxy_IgnoresXFF(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.RemoteAddr = "1.1.1.1:5000"
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	if got := clientIP(req, false); got != "1.1.1.1" {
		t.Errorf("trustProxy=false: должны игнорить XFF; ожидался 1.1.1.1, got %q", got)
	}
}

func TestClientIP_TrustProxy_FallsBackToXRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.RemoteAddr = "1.1.1.1:5000"
	req.Header.Set("X-Real-IP", "8.8.4.4")
	if got := clientIP(req, true); got != "8.8.4.4" {
		t.Errorf("trustProxy=true + только XRealIP: ожидался 8.8.4.4, got %q", got)
	}
}

func TestClientIP_TrustProxy_MalformedXFF_Fallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.RemoteAddr = "1.1.1.1:5000"
	req.Header.Set("X-Forwarded-For", "  ,  ") // empty after trim
	got := clientIP(req, true)
	if got != "1.1.1.1" {
		t.Errorf("malformed XFF: должны fall back на RemoteAddr; got %q", got)
	}
}

// _ = context.Background — keep import.
var _ = context.Background
