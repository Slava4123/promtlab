package tag_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"promptvault/internal/delivery/http/tag"
	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
)

// fakeOrphanRepo — узкий fake для OrphanAnalyticsRepo интерфейса.
type fakeOrphanRepo struct {
	tags []repo.TagRow
	err  error
}

func (f *fakeOrphanRepo) OrphanTags(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.TagRow, error) {
	return f.tags, f.err
}

func injectUserID(uid uint) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), authmw.UserIDKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TestOrphanHandlerHappy(t *testing.T) {
	ar := &fakeOrphanRepo{tags: []repo.TagRow{{TagID: 1, Name: "deprecated"}}}
	h := tag.NewOrphanHandler(ar)

	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/tags/orphan", h.List)

	req := httptest.NewRequest("GET", "/api/tags/orphan", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items []struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Len(t, body.Items, 1)
	require.Equal(t, "deprecated", body.Items[0].Name)
	require.Equal(t, uint(1), body.Items[0].ID)
}

func TestOrphanHandlerEmpty(t *testing.T) {
	ar := &fakeOrphanRepo{tags: []repo.TagRow{}}
	h := tag.NewOrphanHandler(ar)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/tags/orphan", h.List)
	req := httptest.NewRequest("GET", "/api/tags/orphan", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items []any `json:"items"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Len(t, body.Items, 0)
}
