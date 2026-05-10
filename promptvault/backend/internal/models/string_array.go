// MJ-38: drop-in replacement для lib/pq.StringArray. lib/pq — maintenance-mode,
// pgx (через gorm.io/driver/postgres ≥ v1.5) — основной driver. Этот тип реализует
// sql.Scanner / driver.Valuer для PostgreSQL text[] формата без зависимости от pq.
//
// Совместим с существующим on-disk форматом (тот же text[] литерал — `{"a","b"}`),
// поэтому миграция данных не требуется.
package models

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
)

// StringArray — массив строк, совместимый с PostgreSQL text[].
//
// Use:
//
//	type APIKey struct {
//	    AllowedTools models.StringArray `gorm:"type:text[]"`
//	}
type StringArray []string

// Value сериализует массив в PostgreSQL array literal: `{"a","b","c\"d"}`.
// Спецсимволы экранируются: `"` → `\"`, `\` → `\\`. NULL-значения не поддерживаются —
// вызывайте `Value()` на не-nil receiver.
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	if len(a) == 0 {
		return "{}", nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, s := range a {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('"')
		// Эскейпим обратные слэши и кавычки.
		for _, r := range s {
			if r == '\\' || r == '"' {
				buf.WriteByte('\\')
			}
			buf.WriteRune(r)
		}
		buf.WriteByte('"')
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

// Scan десериализует PostgreSQL array literal в []string.
// Поддерживает quoted и unquoted элементы, escaped quotes/backslashes.
func (a *StringArray) Scan(src any) error {
	if src == nil {
		*a = nil
		return nil
	}
	var raw string
	switch v := src.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		return fmt.Errorf("models.StringArray: unsupported Scan type %T", src)
	}
	if raw == "" || raw == "{}" {
		*a = []string{}
		return nil
	}
	if len(raw) < 2 || raw[0] != '{' || raw[len(raw)-1] != '}' {
		return errors.New("models.StringArray: invalid array literal (missing braces)")
	}
	body := raw[1 : len(raw)-1]
	parsed, err := parseArrayBody(body)
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}

// parseArrayBody — разбор тела `"a","b","c"` или `a,b,c`.
// Поддерживает quoted strings с escape-sequence \\ и \".
func parseArrayBody(body string) ([]string, error) {
	out := []string{}
	i := 0
	for i < len(body) {
		// Пропускаем пробелы между элементами.
		for i < len(body) && body[i] == ' ' {
			i++
		}
		if i >= len(body) {
			break
		}
		if body[i] == '"' {
			// Quoted element — читаем до закрывающей кавычки, обрабатывая escape.
			i++
			var elem bytes.Buffer
			for i < len(body) {
				c := body[i]
				if c == '\\' && i+1 < len(body) {
					elem.WriteByte(body[i+1])
					i += 2
					continue
				}
				if c == '"' {
					i++
					break
				}
				elem.WriteByte(c)
				i++
			}
			out = append(out, elem.String())
		} else {
			// Unquoted element — до запятой или конца.
			start := i
			for i < len(body) && body[i] != ',' {
				i++
			}
			elem := body[start:i]
			// PostgreSQL передаёт NULL без кавычек; мы трактуем как пустую строку
			// для простоты (text[] обычно создан с NOT NULL elements).
			if elem == "NULL" {
				out = append(out, "")
			} else {
				out = append(out, elem)
			}
		}
		// Пропускаем разделитель.
		if i < len(body) && body[i] == ',' {
			i++
		}
	}
	return out, nil
}
