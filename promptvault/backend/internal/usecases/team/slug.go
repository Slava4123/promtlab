package team

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode"
)

var cyrillic = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d",
	'е': "e", 'ё': "yo", 'ж': "zh", 'з': "z", 'и': "i",
	'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
	'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
	'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch",
	'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "",
	'э': "e", 'ю': "yu", 'я': "ya",
}

func generateSlug(name string) string {
	name = strings.ToLower(name)

	var b strings.Builder
	for _, r := range name {
		if lat, ok := cyrillic[r]; ok {
			b.WriteString(lat)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}

	slug := b.String()

	// Убираем множественные дефисы и trim
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = "team"
	}

	// Обрезаем ДО добавления суффикса чтобы суффикс не пострадал
	if len(slug) > 93 {
		slug = slug[:93]
	}

	// Добавляем 6-char hex суффикс для уникальности (16M комбинаций)
	suffix := make([]byte, 3)
	if _, err := rand.Read(suffix); err != nil {
		suffix = []byte{0x42, 0xff, 0xab} // fallback
	}
	slug = slug + "-" + hex.EncodeToString(suffix)

	return slug
}
