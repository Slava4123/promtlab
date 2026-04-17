package seo

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"promptvault/internal/models"
)

type fakeLister struct {
	prompts  []models.Prompt
	calls    atomic.Int32
	err      error
	bySlug   map[string]*models.Prompt
	slugErr  error
}

func (f *fakeLister) ListPublic(_ context.Context, _ int) ([]models.Prompt, error) {
	f.calls.Add(1)
	if f.err != nil {
		return nil, f.err
	}
	return f.prompts, nil
}

func (f *fakeLister) GetPublicBySlug(_ context.Context, slug string) (*models.Prompt, error) {
	if f.slugErr != nil {
		return nil, f.slugErr
	}
	if p, ok := f.bySlug[slug]; ok {
		return p, nil
	}
	return nil, errors.New("not found")
}

func newReq() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
}

func TestSitemap_Empty(t *testing.T) {
	h := NewHandler(&fakeLister{}, "https://promtlabs.ru")
	rec := httptest.NewRecorder()
	h.Sitemap(rec, newReq())

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/xml") {
		t.Errorf("content-type: got %q", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"`) {
		t.Errorf("missing namespace: %s", body)
	}
	if strings.Contains(body, "<url>") {
		t.Errorf("expected empty urlset, got: %s", body)
	}
}

func TestSitemap_WithPrompts(t *testing.T) {
	now := time.Date(2026, 4, 17, 15, 30, 0, 0, time.UTC)
	lister := &fakeLister{prompts: []models.Prompt{
		{ID: 1, Slug: "first-a", UpdatedAt: now},
		{ID: 2, Slug: "second-b", UpdatedAt: now.Add(-1 * time.Hour)},
	}}
	h := NewHandler(lister, "https://promtlabs.ru/")  // trailing slash должен отрезаться

	rec := httptest.NewRecorder()
	h.Sitemap(rec, newReq())

	body := rec.Body.String()
	for _, want := range []string{
		"<loc>https://promtlabs.ru/p/first-a</loc>",
		"<loc>https://promtlabs.ru/p/second-b</loc>",
		"<lastmod>2026-04-17T15:30:00Z</lastmod>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q\n---\n%s", want, body)
		}
	}
}

func TestSitemap_SkipsEmptySlug(t *testing.T) {
	lister := &fakeLister{prompts: []models.Prompt{
		{ID: 1, Slug: "", UpdatedAt: time.Now()},   // должен пропуститься
		{ID: 2, Slug: "valid-x", UpdatedAt: time.Now()},
	}}
	h := NewHandler(lister, "https://promtlabs.ru")
	rec := httptest.NewRecorder()
	h.Sitemap(rec, newReq())

	body := rec.Body.String()
	if strings.Count(body, "<url>") != 1 {
		t.Errorf("expected 1 <url>, got: %s", body)
	}
	if !strings.Contains(body, "valid-x") {
		t.Errorf("missing valid slug")
	}
}

func TestSitemap_Cache(t *testing.T) {
	lister := &fakeLister{prompts: []models.Prompt{
		{ID: 1, Slug: "x", UpdatedAt: time.Now()},
	}}
	h := NewHandler(lister, "https://promtlabs.ru")

	for range 5 {
		rec := httptest.NewRecorder()
		h.Sitemap(rec, newReq())
		if rec.Code != 200 {
			t.Fatalf("status %d", rec.Code)
		}
		_, _ = io.Copy(io.Discard, rec.Body)
	}
	if got := lister.calls.Load(); got != 1 {
		t.Errorf("expected 1 SQL call (cache), got %d", got)
	}
}

func TestSitemap_ErrorPropagates(t *testing.T) {
	lister := &fakeLister{err: errors.New("db down")}
	h := NewHandler(lister, "https://promtlabs.ru")
	rec := httptest.NewRecorder()
	h.Sitemap(rec, newReq())

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}
