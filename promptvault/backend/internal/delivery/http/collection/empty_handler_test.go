package collection_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"promptvault/internal/delivery/http/collection"
	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
)

// fakeEmptyRepo — узкий fake для EmptyAnalyticsRepo интерфейса.
type fakeEmptyRepo struct {
	collections []repo.CollectionRow
	err         error
}

func (f *fakeEmptyRepo) EmptyCollections(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.CollectionRow, error) {
	return f.collections, f.err
}

func injectUserID(uid uint) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), authmw.UserIDKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TestEmptyHandlerHappy(t *testing.T) {
	ar := &fakeEmptyRepo{collections: []repo.CollectionRow{{CollectionID: 9, Name: "Old"}}}
	h := collection.NewEmptyHandler(ar)

	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/collections/empty", h.List)

	req := httptest.NewRequest("GET", "/api/collections/empty", nil)
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
	require.Equal(t, "Old", body.Items[0].Name)
	require.Equal(t, uint(9), body.Items[0].ID)
}

func TestEmptyHandlerEmpty(t *testing.T) {
	ar := &fakeEmptyRepo{collections: []repo.CollectionRow{}}
	h := collection.NewEmptyHandler(ar)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/collections/empty", h.List)
	req := httptest.NewRequest("GET", "/api/collections/empty", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items []any `json:"items"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Len(t, body.Items, 0)
}
