package mcpserver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	repo "promptvault/internal/interface/repository"
)

func TestCursor_EncodeDecodeRoundtrip(t *testing.T) {
	original := cursorData{
		Lid: 42,
		Lts: time.Date(2026, 4, 18, 10, 30, 0, 123456789, time.UTC),
		Fh:  "a3f2b1",
	}
	encoded, err := encodeCursor(original)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decoded, err := decodeCursor(encoded)
	assert.NoError(t, err)
	assert.Equal(t, original.Lid, decoded.Lid)
	assert.True(t, original.Lts.Equal(decoded.Lts))
	assert.Equal(t, original.Fh, decoded.Fh)
}

func TestCursor_EmptyIsZeroValue(t *testing.T) {
	c, err := decodeCursor("")
	assert.NoError(t, err)
	assert.Equal(t, uint(0), c.Lid)
}

func TestCursor_InvalidBase64(t *testing.T) {
	_, err := decodeCursor("!!!not-base64!!!")
	assert.Error(t, err)
}

func TestCursor_InvalidJSON(t *testing.T) {
	// валидный base64 "garbage" — не JSON.
	_, err := decodeCursor("Z2FyYmFnZQ")
	assert.Error(t, err)
}

func TestCursor_MissingLID(t *testing.T) {
	encoded, _ := encodeCursor(cursorData{Lid: 0, Lts: time.Now(), Fh: "x"})
	_, err := decodeCursor(encoded)
	assert.Error(t, err)
}

func TestFilterHash_Stable(t *testing.T) {
	f := repo.PromptListFilter{UserID: 1, Query: "hi", FavoriteOnly: true}
	assert.Equal(t, filterHash(f), filterHash(f))
}

func TestFilterHash_DiffersOnChange(t *testing.T) {
	base := repo.PromptListFilter{UserID: 1, Query: "hi"}
	modified := base
	modified.Query = "bye"
	assert.NotEqual(t, filterHash(base), filterHash(modified))
}

func TestFilterHash_IgnoresPagination(t *testing.T) {
	a := repo.PromptListFilter{UserID: 1, Page: 1, PageSize: 10}
	b := repo.PromptListFilter{UserID: 1, Page: 5, PageSize: 50}
	assert.Equal(t, filterHash(a), filterHash(b))
}

func TestFilterHash_IgnoresCursorFields(t *testing.T) {
	// AfterID/AfterUpdatedAt меняются между страницами, но фильтр
	// семантически тот же — хеш обязан быть стабильным.
	id := uint(42)
	ts := time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC)
	a := repo.PromptListFilter{UserID: 1, Query: "go"}
	b := repo.PromptListFilter{UserID: 1, Query: "go", AfterID: &id, AfterUpdatedAt: &ts}
	assert.Equal(t, filterHash(a), filterHash(b))
}
