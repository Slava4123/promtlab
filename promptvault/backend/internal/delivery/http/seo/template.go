package seo

import (
	"bytes"
	"embed"
	"encoding/json"
	"html/template"
	"strings"
	"time"

	"promptvault/internal/models"
)

//go:embed templates/prompt.html
var templatesFS embed.FS

// promptTpl парсится один раз при старте — Execute thread-safe.
// html/template auto-escape: title/content/description с user-input безопасны
// (Go экранирует <>&"'). JSONLD передаётся как template.JS — нам нужен raw JSON
// внутри <script>, без двойного escape.
var promptTpl = template.Must(template.ParseFS(templatesFS, "templates/prompt.html"))

// promptViewModel — данные для prompt.html. Все строки кроме JSONLD проходят
// auto-escape (XSS-safe). JSONLD — pre-marshalled JSON, оборачивается в template.JS.
type promptViewModel struct {
	Title              string
	Description        string
	Content            string
	CanonicalURL       string
	ImageURL           string
	DatePublished      string // RFC3339
	DateModified       string
	DatePublishedHuman string // "17 апреля 2026"
	JSONLD             template.JS
}

// renderPromptHTML — формирует view-model и рендерит шаблон.
// frontendURL без trailing slash, prompt.Slug гарантированно не пустой
// (caller проверил).
func renderPromptHTML(p *models.Prompt, frontendURL string) ([]byte, error) {
	base := strings.TrimRight(frontendURL, "/")
	canonical := base + "/p/" + p.Slug
	imageURL := base + "/api/og/prompts/" + p.Slug + ".png" // PR3 добавит endpoint
	desc := makeDescription(p.Content, 160)

	jsonld, err := buildJSONLD(p, canonical, imageURL)
	if err != nil {
		return nil, err
	}

	vm := promptViewModel{
		Title:              p.Title,
		Description:        desc,
		Content:            p.Content,
		CanonicalURL:       canonical,
		ImageURL:           imageURL,
		DatePublished:      p.CreatedAt.UTC().Format(time.RFC3339),
		DateModified:       p.UpdatedAt.UTC().Format(time.RFC3339),
		DatePublishedHuman: formatRussianDate(p.CreatedAt),
		JSONLD:             template.JS(jsonld),
	}

	var buf bytes.Buffer
	if err := promptTpl.Execute(&buf, vm); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// makeDescription — берёт первые maxLen символов content, обрезает по слову,
// добавляет «…». Удаляет переводы строк для meta-description (single-line).
func makeDescription(content string, maxLen int) string {
	cleaned := strings.Join(strings.Fields(content), " ")
	if len(cleaned) <= maxLen {
		return cleaned
	}
	// безопасная обрезка по unicode-границе
	runes := []rune(cleaned)
	if len(runes) <= maxLen {
		return cleaned
	}
	return string(runes[:maxLen-1]) + "…"
}

// buildJSONLD — Schema.org Article. Поля по Google Search Central рекомендациям.
// author намеренно опущен (нужен Preload User в репо — отдельный PR).
func buildJSONLD(p *models.Prompt, canonical, imageURL string) ([]byte, error) {
	headline := p.Title
	if r := []rune(headline); len(r) > 110 {
		headline = string(r[:110])
	}
	jsonld := map[string]any{
		"@context":         "https://schema.org",
		"@type":            "Article",
		"headline":         headline,
		"description":      makeDescription(p.Content, 160),
		"datePublished":    p.CreatedAt.UTC().Format(time.RFC3339),
		"dateModified":     p.UpdatedAt.UTC().Format(time.RFC3339),
		"image":            imageURL,
		"mainEntityOfPage": map[string]string{"@type": "WebPage", "@id": canonical},
		"publisher": map[string]any{
			"@type": "Organization",
			"name":  "ПромтЛаб",
			"url":   "https://promtlabs.ru",
		},
	}
	return json.Marshal(jsonld)
}

var russianMonths = [...]string{
	"января", "февраля", "марта", "апреля", "мая", "июня",
	"июля", "августа", "сентября", "октября", "ноября", "декабря",
}

func formatRussianDate(t time.Time) string {
	t = t.UTC()
	return formatDay(t.Day()) + " " + russianMonths[t.Month()-1] + " " + formatYear(t.Year())
}

func formatDay(d int) string {
	if d < 10 {
		return string(rune('0'+d))
	}
	return string(rune('0'+d/10)) + string(rune('0'+d%10))
}

func formatYear(y int) string {
	return string(rune('0'+y/1000)) +
		string(rune('0'+(y/100)%10)) +
		string(rune('0'+(y/10)%10)) +
		string(rune('0'+y%10))
}
