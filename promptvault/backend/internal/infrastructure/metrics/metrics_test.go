package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestShareQuotaIncrementFailed_Inc(t *testing.T) {
	before := testutil.ToFloat64(ShareQuotaIncrementFailed)
	ShareQuotaIncrementFailed.Inc()
	after := testutil.ToFloat64(ShareQuotaIncrementFailed)
	if after-before != 1 {
		t.Fatalf("expected +1 after Inc, got diff=%v", after-before)
	}
}

func TestInsightsRefresh_Labels(t *testing.T) {
	InsightsRefresh.WithLabelValues("success").Inc()
	InsightsRefresh.WithLabelValues("rate_limited").Inc()
	InsightsRefresh.WithLabelValues("error").Inc()
	if got := testutil.ToFloat64(InsightsRefresh.WithLabelValues("success")); got < 1 {
		t.Fatalf("expected success counter >= 1, got %v", got)
	}
}

func TestHandler_Enabled_ReturnsMetrics(t *testing.T) {
	h := Handler(true)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "share_quota_increment_failed_total") {
		t.Fatal("metrics endpoint must expose registered counters")
	}
}

func TestHandler_Disabled_Returns404(t *testing.T) {
	h := Handler(false)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when disabled, got %d", rr.Code)
	}
}
