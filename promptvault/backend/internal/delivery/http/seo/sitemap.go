package seo

import (
	"bytes"
	"context"
	"encoding/xml"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// Лимит sitemap.xml по протоколу — 50K URL / 50MB. Текущий ListPublic
// зашивает 1000-10000. При близости к лимиту → log.Warn + миграция на sitemap-index.
const (
	sitemapMaxURLs    = 50_000
	sitemapWarnURLs   = 45_000 // алерт за 5K до лимита
	sitemapListLimit  = 10_000 // совпадает с PromptService.ListPublic max
	sitemapCacheTTL   = time.Hour
)

// sitemapURL — один <url> в выводе. Только canonical-fields:
// changefreq и priority Google игнорирует с 2024+ (Respona, 2026 best practice).
type sitemapURL struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
	LastMod string   `xml:"lastmod,omitempty"`
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// sitemapCache — простой in-memory TTL-cache. Один writer (mu.Lock на rebuild),
// много читателей (RLock на hot path). Достаточно для current scale.
type sitemapCache struct {
	mu        sync.RWMutex
	body      []byte
	expiresAt time.Time
}

func newSitemapCache() *sitemapCache {
	return &sitemapCache{}
}

// Get возвращает рендеренный sitemap.xml, перестраивая кэш по TTL.
func (c *sitemapCache) Get(ctx context.Context, lister PromptService, frontendURL string) ([]byte, error) {
	c.mu.RLock()
	if c.body != nil && time.Now().Before(c.expiresAt) {
		body := c.body
		c.mu.RUnlock()
		return body, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	// double-check после получения write-lock — другая горутина могла обновить.
	if c.body != nil && time.Now().Before(c.expiresAt) {
		return c.body, nil
	}

	body, err := buildSitemap(ctx, lister, frontendURL)
	if err != nil {
		return nil, err
	}
	c.body = body
	c.expiresAt = time.Now().Add(sitemapCacheTTL)
	return body, nil
}

// buildSitemap делает SQL-запрос + сериализует XML.
func buildSitemap(ctx context.Context, lister PromptService, frontendURL string) ([]byte, error) {
	prompts, err := lister.ListPublic(ctx, sitemapListLimit)
	if err != nil {
		return nil, err
	}

	if len(prompts) >= sitemapWarnURLs {
		slog.Warn("seo.sitemap.size_warning", "count", len(prompts), "max", sitemapMaxURLs)
	}

	base := strings.TrimRight(frontendURL, "/")
	urls := make([]sitemapURL, 0, len(prompts))
	for _, p := range prompts {
		if p.Slug == "" {
			continue // защита от promпт без slug (не должно случиться при is_public=TRUE)
		}
		urls = append(urls, sitemapURL{
			Loc:     base + "/p/" + p.Slug,
			LastMod: p.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	urlset := sitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(urlset); err != nil {
		return nil, err
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
