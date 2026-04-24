// Package seo предоставляет server-side endpoints для SEO/discovery:
// sitemap.xml, server-rendered HTML для bot-UA, OG-image generation.
//
// Это отдельный слой над usecases/prompt — никакой бизнес-логики, только
// presentation для поисковиков и соц-парсеров. См. plan: /plans/playful-whistling-snail.md.
package seo

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// PromptService — узкий интерфейс для DI. Только то что нужно SEO-слою.
// Используется существующий *prompt.Service.
type PromptService interface {
	ListPublic(ctx context.Context, limit int) ([]models.Prompt, error)
	GetPublicBySlug(ctx context.Context, slug string) (*models.Prompt, error)
}

// Handler собирает все SEO-endpoints. Тонкий слой презентации.
type Handler struct {
	prompts     PromptService
	frontendURL string // канонический base-URL без trailing slash, напр. https://promtlabs.ru
	sitemap     *sitemapCache
}

func NewHandler(prompts PromptService, frontendURL string) *Handler {
	return &Handler{
		prompts:     prompts,
		frontendURL: frontendURL,
		sitemap:     newSitemapCache(),
	}
}

// PromptHTML — GET /p/{slug}. Server-rendered HTML для bot-UA (nginx роутит сюда
// только Yandexbot/Googlebot/Telegram/VK и т.д.). Обычный юзер получает SPA.
//
// Errors:
//   404 Not Found — slug не существует или промпт не публичный
//   500 Internal Server Error — render fail (Sentry middleware capturit)
func (h *Handler) PromptHTML(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	p, err := h.prompts.GetPublicBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			slog.Warn("seo.html.not_found", "slug", slug)
		} else {
			slog.Error("seo.html.lookup_failed", "slug", slug, "error", err)
		}
		http.NotFound(w, r)
		return
	}

	body, err := renderPromptHTML(p, h.frontendURL)
	if err != nil {
		slog.Error("seo.html.render_failed", "slug", slug, "error", err)
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("X-Robots-Tag", "index, follow")
	_, _ = w.Write(body)
	slog.Info("seo.html.served", "slug", slug, "ua", r.Header.Get("User-Agent"), "duration_ms", time.Since(start).Milliseconds())
}

// OGImage — GET /api/og/prompts/{slug}.png. Динамический OG-PNG с заголовком.
// ETag по updated_at — клиенты получают 304 без повторного рендера.
func (h *Handler) OGImage(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSuffix(chi.URLParam(r, "slug"), ".png")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	p, err := h.prompts.GetPublicBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	etag := ogETag(p.Slug, p.UpdatedAt)
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
		w.WriteHeader(http.StatusNotModified)
		return
	}

	body, err := renderOGImage(p.Title)
	if err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	w.Header().Set("ETag", etag)
	_, _ = w.Write(body)
}

// Sitemap — GET /sitemap.xml. Список публичных промптов в формате sitemaps.org/0.9.
// In-memory cache 1ч (изменения публичных промптов редкие, full-rebuild дёшев).
func (h *Handler) Sitemap(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	body, err := h.sitemap.Get(r.Context(), h.prompts, h.frontendURL)
	if err != nil {
		slog.Error("seo.sitemap.failed", "error", err)
		http.Error(w, "sitemap unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write(body)
	slog.Info("seo.sitemap.served", "bytes", len(body), "duration_ms", time.Since(start).Milliseconds())
}
