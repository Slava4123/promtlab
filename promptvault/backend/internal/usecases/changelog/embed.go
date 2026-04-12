package changelog

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// changelogBytes содержит контент changelog.json, встроенный в бинарник.
// Источник правды — git, обновления через PR + редеплой.
//
//go:embed changelog.json
var changelogBytes []byte

// loadEmbeddedChangelog парсит встроенный JSON в типизированный Changelog.
// Вызывается из NewService один раз при старте.
func loadEmbeddedChangelog() (*Changelog, error) {
	var c Changelog
	if err := json.Unmarshal(changelogBytes, &c); err != nil {
		return nil, fmt.Errorf("changelog: parse embedded changelog.json: %w", err)
	}
	if len(c.Entries) == 0 {
		return nil, fmt.Errorf("changelog: changelog.json contains no entries")
	}
	return &c, nil
}
