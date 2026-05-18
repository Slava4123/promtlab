package prompt_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"promptvault/internal/delivery/http/prompt"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/usecases/prompt_insights"
)

// fakeInsightsService реализует prompt.InsightsService с настраиваемыми return values.
type fakeInsightsService struct {
	unused     []prompt_insights.PromptInsightRow
	duplicates []prompt_insights.DuplicatePair
	trending   []prompt_insights.PromptInsightRow
	declining  []prompt_insights.PromptInsightRow
	mostEdited []prompt_insights.PromptInsightRow
	err        error
	mergeErr   error
}

func (f *fakeInsightsService) ListUnused(_ context.Context, _ uint, _ *uint, _ int) ([]prompt_insights.PromptInsightRow, error) {
	return f.unused, f.err
}
func (f *fakeInsightsService) ListDuplicates(_ context.Context, _ uint, _ *uint, _ int) ([]prompt_insights.DuplicatePair, error) {
	return f.duplicates, f.err
}
func (f *fakeInsightsService) ListTrending(_ context.Context, _ uint, _ *uint, _ int) ([]prompt_insights.PromptInsightRow, error) {
	return f.trending, f.err
}
func (f *fakeInsightsService) ListDeclining(_ context.Context, _ uint, _ *uint, _ int) ([]prompt_insights.PromptInsightRow, error) {
	return f.declining, f.err
}
func (f *fakeInsightsService) ListMostEdited(_ context.Context, _ uint, _ *uint, _ int) ([]prompt_insights.PromptInsightRow, error) {
	return f.mostEdited, f.err
}
func (f *fakeInsightsService) MergePrompts(_ context.Context, _, _, _ uint) error {
	return f.mergeErr
}

// injectUserID — middleware-stub для тестов handlers, требующих authmw.GetUserID.
// Использует context.WithValue с тем же ключом, что и middleware/auth/auth.go.
func injectUserID(uid uint) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), authmw.UserIDKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TestInsightsHandlerUnused200(t *testing.T) {
	svc := &fakeInsightsService{unused: []prompt_insights.PromptInsightRow{{PromptID: 1, Title: "X", Uses: 0}}}
	h := prompt.NewInsightsHandler(svc)

	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/prompts/insights/unused", h.Unused)

	req := httptest.NewRequest("GET", "/api/prompts/insights/unused?limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items []prompt_insights.PromptInsightRow `json:"items"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Len(t, body.Items, 1)
	require.Equal(t, uint(1), body.Items[0].PromptID)
}

func TestInsightsHandlerUnused402WhenProRequired(t *testing.T) {
	svc := &fakeInsightsService{err: prompt_insights.ErrProRequired}
	h := prompt.NewInsightsHandler(svc)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/prompts/insights/unused", h.Unused)
	req := httptest.NewRequest("GET", "/api/prompts/insights/unused", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusPaymentRequired, w.Code)
}

func TestInsightsHandlerDuplicates200(t *testing.T) {
	svc := &fakeInsightsService{
		duplicates: []prompt_insights.DuplicatePair{
			{PromptA: prompt_insights.PromptInsightRow{PromptID: 1, Title: "A"}, PromptB: prompt_insights.PromptInsightRow{PromptID: 2, Title: "B"}, Similarity: 0.9},
		},
	}
	h := prompt.NewInsightsHandler(svc)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Get("/api/prompts/insights/duplicates", h.Duplicates)
	req := httptest.NewRequest("GET", "/api/prompts/insights/duplicates", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Items []prompt_insights.DuplicatePair `json:"items"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Len(t, body.Items, 1)
	require.Equal(t, 0.9, body.Items[0].Similarity)
}

func TestInsightsHandlerMergeHappy(t *testing.T) {
	svc := &fakeInsightsService{} // mergeErr nil by default
	h := prompt.NewInsightsHandler(svc)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Post("/api/prompts/{id}/merge-with/{other_id}", h.Merge)

	req := httptest.NewRequest("POST", "/api/prompts/1/merge-with/2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body struct {
		KeptID   uint `json:"kept_id"`
		MergedID uint `json:"merged_id"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Equal(t, uint(1), body.KeptID)
	require.Equal(t, uint(2), body.MergedID)
}

func TestInsightsHandlerMergeNotOwned404(t *testing.T) {
	svc := &fakeInsightsService{mergeErr: prompt_insights.ErrPromptsNotOwned}
	h := prompt.NewInsightsHandler(svc)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Post("/api/prompts/{id}/merge-with/{other_id}", h.Merge)
	req := httptest.NewRequest("POST", "/api/prompts/1/merge-with/2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestInsightsHandlerMergeSameID400(t *testing.T) {
	svc := &fakeInsightsService{mergeErr: prompt_insights.ErrSamePrompt}
	h := prompt.NewInsightsHandler(svc)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Post("/api/prompts/{id}/merge-with/{other_id}", h.Merge)
	req := httptest.NewRequest("POST", "/api/prompts/5/merge-with/5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInsightsHandlerMergeBadID(t *testing.T) {
	svc := &fakeInsightsService{}
	h := prompt.NewInsightsHandler(svc)
	r := chi.NewRouter()
	r.With(injectUserID(42)).Post("/api/prompts/{id}/merge-with/{other_id}", h.Merge)
	req := httptest.NewRequest("POST", "/api/prompts/abc/merge-with/2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}
