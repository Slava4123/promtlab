package analytics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAddBreadcrumb_NoHub — breadcrumb без Sentry hub не паникует.
// Pattern для всех analytics-HTTP-тестов: handler работает корректно
// даже если Sentry feature-flag выключен. Полноценные HTTP-тесты
// handler'ов потребуют выделения `analyticsService` интерфейса
// (сейчас *analyticsuc.Service — concrete struct) и отдельного PR.
func TestAddBreadcrumb_NoHub(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/analytics/personal", nil)
	// Нет sentry.Hub в контексте — addBreadcrumb должен быть no-op.
	addBreadcrumb(r, "analytics", "test.event", map[string]any{"user_id": 42})
	// Если не упало — успех.
}
