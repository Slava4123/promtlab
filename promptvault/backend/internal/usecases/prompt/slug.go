package prompt

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/mozillazg/go-unidecode"
)

// makeSlug — детерминированный slug из title + id.
// title → unidecode (cyrillic/прочая non-latin → латинская транслитерация),
// затем lower и оставляем только [a-z0-9] и тире. Добавляем id в base36
// для гарантированной уникальности: title_part + "-" + id36.
// Если title пустой после sanitize — fallback "p-<id>".
//
// Длина slug ≤ 120 (column limit). Если slug из title > 100 — обрезаем.
func makeSlug(id uint, title string) string {
	id36 := strconv.FormatUint(uint64(id), 36)
	clean := sanitizeTitle(title)
	if clean == "" {
		return "p-" + id36
	}
	if len(clean) > 100 {
		clean = strings.TrimRight(clean[:100], "-")
	}
	return clean + "-" + id36
}

func sanitizeTitle(s string) string {
	// Транслитерация: "Мой первый промпт" → "Moi pervyi prompt"
	// (мягкий знак становится "'", тире — обрабатываем ниже).
	s = unidecode.Unidecode(s)
	s = strings.ToLower(s)
	var b strings.Builder
	var lastDash bool
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		case unicode.IsSpace(r) || r == '-' || r == '_':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
			// Любые другие символы (апостроф, пунктуация) пропускаем —
			// слова склеиваются: "don't" → "dont", "ёж" → "ezh".
		}
	}
	return strings.Trim(b.String(), "-")
}
