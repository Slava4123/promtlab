package share

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// Регрессия на L1: truncateString обрезал по байтам — если граница падала
// на середину кириллической руны, в PostgreSQL шёл invalid UTF-8 и запись
// в share_view_log терялась.

func TestTruncateString_ShortStays(t *testing.T) {
	assert.Equal(t, "hello", truncateString("hello", 10))
	assert.Equal(t, "", truncateString("", 10))
}

func TestTruncateString_ASCIITruncated(t *testing.T) {
	assert.Equal(t, "hello", truncateString("hello world", 5))
}

func TestTruncateString_RussianRuneBoundary(t *testing.T) {
	// "привет" — каждая буква = 2 байта. len = 12.
	// Обрезка по 5 байтам в старом варианте резала «при» + половина «в».
	s := "привет мир"
	out := truncateString(s, 5)
	assert.True(t, utf8.ValidString(out), "result must remain valid UTF-8")
	// 5 байт = 2 полные руны «пр» (4 байта), 3-я руна «и» не влазит — обрезается.
	assert.Equal(t, "пр", out)
}

func TestTruncateString_ExactUTF8Boundary(t *testing.T) {
	// «аб» — ровно 4 байта. maxLen=4 — помещается полностью.
	assert.Equal(t, "аб", truncateString("абвг", 4))
}

func TestTruncateString_Emoji(t *testing.T) {
	// Эмодзи = 4 байта каждый. maxLen=3 — ни одна не помещается.
	assert.Equal(t, "", truncateString("🚀🎉", 3))
	// maxLen=4 — одна ракета.
	assert.Equal(t, "🚀", truncateString("🚀🎉", 4))
}

func TestTruncateString_LongReferer(t *testing.T) {
	// Реальный сценарий: Referer с кириллицей > 500 байт.
	base := "https://example.com/страница/"
	long := base + strings.Repeat("документ", 100)
	out := truncateString(long, 500)
	assert.LessOrEqual(t, len(out), 500)
	assert.True(t, utf8.ValidString(out))
}
