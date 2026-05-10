// MN-6 — middleware/admin tests: RequireAdmin (RBAC + frozen guard) и
// AdminAuditContext (IP-spoof guard через trustProxy).
package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	"promptvault/internal/usecases/audit"
)

// fakeUserLookup — минимальный fake UserLookup.
type fakeUserLookup struct {
	user *models.User
	err  error
}

func (f *fakeUserLookup) GetByID(_ context.Context, _ uint) (*models.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.user, nil
}

// withUserID кладёт userID в ctx как authmw.Middleware (тот же ключ).
func reqWithUserID(uid uint) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/admin/x", nil)
	ctx := context.WithValue(r.Context(), authmw.UserIDKey, uid)
	return r.WithContext(ctx)
}

// --- RequireAdmin ---

func TestRequireAdmin_NoUserID_401(t *testing.T) {
	mw := RequireAdmin(&fakeUserLookup{user: &models.User{Role: models.RoleAdmin}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("not called")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/x", nil))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAdmin_UserNotFound_401(t *testing.T) {
	mw := RequireAdmin(&fakeUserLookup{err: repo.ErrNotFound})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("not called")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, reqWithUserID(42))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAdmin_NotAdmin_403(t *testing.T) {
	mw := RequireAdmin(&fakeUserLookup{user: &models.User{
		ID: 42, Role: models.RoleUser, Status: models.StatusActive,
	}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("not called")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, reqWithUserID(42))
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireAdmin_AdminFrozen_403(t *testing.T) {
	// Re-check guard: admin downgraded → frozen после выпуска JWT.
	mw := RequireAdmin(&fakeUserLookup{user: &models.User{
		ID: 42, Role: models.RoleAdmin, Status: models.StatusFrozen,
	}})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("frozen admin не должен пройти")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, reqWithUserID(42))
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireAdmin_AdminActive_PassesThrough(t *testing.T) {
	mw := RequireAdmin(&fakeUserLookup{user: &models.User{
		ID: 42, Role: models.RoleAdmin, Status: models.StatusActive,
	}})
	called := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, reqWithUserID(42))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, called)
}

// --- AdminAuditContext (extractIP / trustProxy) ---

func TestAdminAuditContext_TrustProxyOff_IgnoresXFF(t *testing.T) {
	mw := AdminAuditContext(false)
	var got audit.AdminRequestInfo
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = audit.FromContext(r.Context())
	}))

	r := reqWithUserID(42)
	r.RemoteAddr = "1.1.1.1:5000"
	r.Header.Set("X-Forwarded-For", "8.8.8.8")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)

	if got.IP != "1.1.1.1" {
		t.Errorf("trustProxy=false: ожидался RemoteAddr 1.1.1.1, got %q (XFF spoofing)", got.IP)
	}
	if got.AdminID != 42 {
		t.Errorf("AdminID = %d, want 42", got.AdminID)
	}
}

func TestAdminAuditContext_TrustProxyOn_HonorsXFF(t *testing.T) {
	mw := AdminAuditContext(true)
	var got audit.AdminRequestInfo
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = audit.FromContext(r.Context())
	}))

	r := reqWithUserID(42)
	r.RemoteAddr = "1.1.1.1:5000"
	r.Header.Set("X-Forwarded-For", "8.8.8.8, 9.9.9.9")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)

	if got.IP != "8.8.8.8" {
		t.Errorf("trustProxy=true: ожидался первый XFF 8.8.8.8, got %q", got.IP)
	}
}

func TestAdminAuditContext_CapturesUserAgent(t *testing.T) {
	mw := AdminAuditContext(false)
	var got audit.AdminRequestInfo
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = audit.FromContext(r.Context())
	}))

	r := reqWithUserID(42)
	r.Header.Set("User-Agent", "TestSuite/1.0")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	assert.Equal(t, "TestSuite/1.0", got.UserAgent)
}
