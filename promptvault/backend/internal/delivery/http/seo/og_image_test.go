package seo

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"promptvault/internal/models"
)

func newOGReq(slug string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/og/prompts/"+slug, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", slug)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestOGImage_RenderReturnsPNG(t *testing.T) {
	now := time.Now()
	p := &models.Prompt{Slug: "x", Title: "Тестовый промпт", IsPublic: true, UpdatedAt: now}
	h := NewHandler(&fakeLister{bySlug: map[string]*models.Prompt{"x": p}}, "https://promtlabs.ru")
	rec := httptest.NewRecorder()
	h.OGImage(rec, newOGReq("x"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("content-type: %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc == "" {
		t.Error("missing Cache-Control")
	}
	if et := rec.Header().Get("ETag"); et == "" {
		t.Error("missing ETag")
	}
	// PNG magic bytes
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte{0x89, 'P', 'N', 'G'}) {
		t.Errorf("body is not PNG (first bytes: %x)", rec.Body.Bytes()[:4])
	}
}

func TestOGImage_StripsExtension(t *testing.T) {
	p := &models.Prompt{Slug: "x", Title: "X", IsPublic: true, UpdatedAt: time.Now()}
	h := NewHandler(&fakeLister{bySlug: map[string]*models.Prompt{"x": p}}, "https://promtlabs.ru")
	rec := httptest.NewRecorder()
	h.OGImage(rec, newOGReq("x.png")) // chi даёт slug="x.png", handler должен strip
	if rec.Code != http.StatusOK {
		t.Errorf("status %d, want 200 for x.png slug", rec.Code)
	}
}

func TestOGImage_NotFound(t *testing.T) {
	h := NewHandler(&fakeLister{bySlug: map[string]*models.Prompt{}}, "https://promtlabs.ru")
	rec := httptest.NewRecorder()
	h.OGImage(rec, newOGReq("ghost"))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status %d, want 404", rec.Code)
	}
}

func TestOGImage_ETag304(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	p := &models.Prompt{Slug: "x", Title: "X", IsPublic: true, UpdatedAt: now}
	h := NewHandler(&fakeLister{bySlug: map[string]*models.Prompt{"x": p}}, "https://promtlabs.ru")

	// Первый запрос — получаем ETag
	rec1 := httptest.NewRecorder()
	h.OGImage(rec1, newOGReq("x"))
	etag := rec1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no etag on first request")
	}

	// Второй запрос с If-None-Match → 304
	req := newOGReq("x")
	req.Header.Set("If-None-Match", etag)
	rec2 := httptest.NewRecorder()
	h.OGImage(rec2, req)
	if rec2.Code != http.StatusNotModified {
		t.Errorf("status %d, want 304", rec2.Code)
	}
	if rec2.Body.Len() != 0 {
		t.Errorf("304 must have empty body, got %d bytes", rec2.Body.Len())
	}
}
