// MJ-26 / MN-26 — typed HexColor с validation, защита от XSS через MCP/import.
//
// Раньше Color был просто `string` с size:20 (collection) или size:7 (tag).
// HTTP-валидатор `hexcolor` стоял только в delivery/http/tag/request.go,
// MCP-tool `create_tag` ходил в Service.Create мимо HTTP-validation.
// Размер 20 в Collection.Color позволял сохранять `red; }</style><script>...`.
//
// Теперь HexColor валидируется через BeforeSave (GORM hook) и UnmarshalJSON,
// независимо от точки входа (HTTP / MCP / import). size:7 везде.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// HexColor — 6-значный hex-цвет вида `#RRGGBB`. Empty string допустим
// (означает «использовать default цвет на frontend").
type HexColor string

// hexColorRe — `#` + ровно 6 hex цифр. Не покрываем 3-значные сокращения
// (`#abc`) и 8-значные с alpha — в проекте мы их не используем.
var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// NewHexColor валидирует и нормализует цвет (lowercase). Empty input
// разрешён — используется default на frontend.
func NewHexColor(raw string) (HexColor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if !hexColorRe.MatchString(raw) {
		return "", fmt.Errorf("invalid hex color %q (expected #RRGGBB)", raw)
	}
	return HexColor(strings.ToLower(raw)), nil
}

// Validate возвращает nil если значение пустое или матчит #RRGGBB.
func (c HexColor) Validate() error {
	if c == "" {
		return nil
	}
	if !hexColorRe.MatchString(string(c)) {
		return fmt.Errorf("invalid hex color %q (expected #RRGGBB)", string(c))
	}
	return nil
}

// String для удобства логов и JSON.
func (c HexColor) String() string { return string(c) }

// UnmarshalJSON — парсит string в HexColor с validation.
// Если поле в JSON отсутствует — оставит zero-value (empty string), что OK.
func (c *HexColor) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := NewHexColor(s)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// Value — для GORM/database/sql. Сохраняем как string.
func (c HexColor) Value() (driver.Value, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if c == "" {
		return nil, nil
	}
	return string(c), nil
}

// Scan — из БД. БД может вернуть NULL — оставляем empty.
func (c *HexColor) Scan(src any) error {
	if src == nil {
		*c = ""
		return nil
	}
	switch v := src.(type) {
	case string:
		*c = HexColor(v)
	case []byte:
		*c = HexColor(string(v))
	default:
		return errors.New("HexColor.Scan: unsupported type")
	}
	return nil
}
