package mcpserver

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	repo "promptvault/internal/interface/repository"
)

// cursorData — непрозрачный курсор для keyset pagination.
// JSON-кодируется и base64-URL-энкодится.
//
// Поля:
//   - Lid: ID последнего промпта на предыдущей странице.
//   - Lts: updated_at того же промпта (RFC3339Nano для точности).
//   - Fh:  truncated SHA-256 фильтра — защита от смены фильтра между страницами.
type cursorData struct {
	Lid uint      `json:"lid"`
	Lts time.Time `json:"lts"`
	Fh  string    `json:"fh"`
}

// ErrCursorFilterMismatch — фильтр изменился между страницами, клиент должен
// начать пагинацию с первой страницы (cursor=nil).
var ErrCursorFilterMismatch = errors.New("cursor_filter_mismatch: filter changed between pages, restart pagination from first page")

// encodeCursor сериализует cursorData в opaque base64-URL строку.
func encodeCursor(c cursorData) (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("cursor marshal: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

// decodeCursor разбирает opaque-строку. Пустая строка возвращает zero-value + nil
// (означает "первая страница").
func decodeCursor(raw string) (cursorData, error) {
	var c cursorData
	if raw == "" {
		return c, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return c, fmt.Errorf("cursor base64 decode: %w", err)
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("cursor json decode: %w", err)
	}
	if c.Lid == 0 {
		return c, errors.New("cursor: lid is required")
	}
	return c, nil
}

// filterHash вычисляет стабильный 16-hex-символьный хеш (8 байт) от значимых
// фильтров. Используется для детекции изменения фильтра между страницами.
// Порядок полей — стабильный; PageSize/Page/AfterID/AfterUpdatedAt намеренно
// НЕ включены — они меняются между страницами без смены "логического" запроса.
func filterHash(f repo.PromptListFilter) string {
	payload := struct {
		UserID       uint   `json:"u"`
		TeamIDs      []uint `json:"t,omitempty"`
		CollectionID *uint  `json:"c,omitempty"`
		TagIDs       []uint `json:"tg,omitempty"`
		FavoriteOnly bool   `json:"f,omitempty"`
		Query        string `json:"q,omitempty"`
	}{
		UserID:       f.UserID,
		TeamIDs:      f.TeamIDs,
		CollectionID: f.CollectionID,
		TagIDs:       f.TagIDs,
		FavoriteOnly: f.FavoriteOnly,
		Query:        f.Query,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8])
}
