package prompt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeSlug_TransliterateCyrillic(t *testing.T) {
	got := makeSlug(42, "Мой первый промпт")
	// "Мой первый промпт" → unidecode → "Moi pervyi prompt" → "moi-pervyi-prompt"
	// id 42 → base36 = "16"
	assert.True(t, strings.HasSuffix(got, "-16"), "ожидается id-суффикс, got=%q", got)
	assert.NotContains(t, got, "p-", "не должно быть fallback на p-<id>")
	assert.Regexp(t, `^[a-z0-9-]+-16$`, got, "только [a-z0-9-]")
}

func TestMakeSlug_MixedRussianEnglish(t *testing.T) {
	got := makeSlug(100, "GPT-4 ответ на русском")
	assert.Contains(t, got, "gpt-4")
	assert.Regexp(t, `^[a-z0-9-]+-2s$`, got)
}

func TestMakeSlug_EnglishOnly(t *testing.T) {
	got := makeSlug(1, "How to write CV")
	assert.Equal(t, "how-to-write-cv-1", got)
}

func TestMakeSlug_EmptyAfterSanitize(t *testing.T) {
	// Только спецсимволы → fallback p-<id>.
	got := makeSlug(1, "!!!@@@")
	assert.Equal(t, "p-1", got)
}

func TestMakeSlug_TruncatesLongTitle(t *testing.T) {
	long := strings.Repeat("very long title ", 20)
	got := makeSlug(1, long)
	assert.LessOrEqual(t, len(got), 120)
	assert.True(t, strings.HasSuffix(got, "-1"))
}

func TestMakeSlug_YoSymbol(t *testing.T) {
	got := makeSlug(7, "Ёжик в тумане")
	// "Ё" → "Io" / "ezh" в зависимости от unidecode-таблицы; точное slug
	// не фиксируем, проверяем что не fallback и что есть кириллический корень.
	assert.NotEqual(t, "p-7", got)
	assert.Regexp(t, `^[a-z0-9-]+-7$`, got)
}

func TestMakeSlug_Apostrophe(t *testing.T) {
	got := makeSlug(2, "Don't repeat")
	// "don't" → "dont" (апостроф — не в whitelist sanitize)
	assert.Contains(t, got, "dont")
}
