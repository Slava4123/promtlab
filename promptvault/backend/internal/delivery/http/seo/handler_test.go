package seo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"promptvault/internal/models"
)

func newSlugReq(slug string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/p/"+slug, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", slug)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestPromptHTML_Success(t *testing.T) {
	now := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	p := &models.Prompt{
		ID: 11, Slug: "hello-world-b", Title: "Hello World", Content: "Это содержимое промпта",
		IsPublic: true, CreatedAt: now, UpdatedAt: now,
	}
	h := NewHandler(&fakeLister{bySlug: map[string]*models.Prompt{"hello-world-b": p}}, "https://promtlabs.ru")
	rec := httptest.NewRecorder()
	h.PromptHTML(rec, newSlugReq("hello-world-b"))

	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("content-type: %q", ct)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"<title>Hello World — ПромтЛаб</title>",
		`<meta property="og:title" content="Hello World">`,
		`<meta property="og:type" content="article">`,
		`<link rel="canonical" href="https://promtlabs.ru/p/hello-world-b">`,
		`<meta property="og:image" content="https://promtlabs.ru/api/og/prompts/hello-world-b.png">`,
		`<script type="application/ld+json">`,
		`"@type":"Article"`,
		`"headline":"Hello World"`,
		`"datePublished":"2026-04-17T12:00:00Z"`,
		"Это содержимое промпта",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in body", want)
		}
	}
}

func TestPromptHTML_NotFound(t *testing.T) {
	h := NewHandler(&fakeLister{slugErr: errors.New("not found")}, "https://promtlabs.ru")
	rec := httptest.NewRecorder()
	h.PromptHTML(rec, newSlugReq("ghost"))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status %d, want 404", rec.Code)
	}
}

func TestPromptHTML_XSSEscape(t *testing.T) {
	p := &models.Prompt{
		ID: 1, Slug: "x", Title: `<script>alert(1)</script>`,
		Content:  `<img src=x onerror="alert(1)">`,
		IsPublic: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	h := NewHandler(&fakeLister{bySlug: map[string]*models.Prompt{"x": p}}, "https://promtlabs.ru")
	rec := httptest.NewRecorder()
	h.PromptHTML(rec, newSlugReq("x"))

	body := rec.Body.String()
	// title должен быть escaped в обоих местах <title> и <meta>
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Errorf("XSS unescaped script in body")
	}
	if !strings.Contains(body, "&lt;script&gt;") && !strings.Contains(body, "&#34;") {
		// html/template должен экранировать в каком-то виде
		t.Errorf("expected html-escaping, got: %s", body[:300])
	}
}
