package starter

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// catalogBytes содержит контент catalog.json, встроенный в бинарник.
// Источник правды — git, обновления через PR + редеплой.
//
//go:embed catalog.json
var catalogBytes []byte

// loadEmbeddedCatalog парсит встроенный JSON в типизированный Catalog.
// Вызывается из NewService один раз при старте.
func loadEmbeddedCatalog() (*Catalog, error) {
	var c Catalog
	if err := json.Unmarshal(catalogBytes, &c); err != nil {
		return nil, fmt.Errorf("starter: parse embedded catalog.json: %w", err)
	}
	if len(c.Categories) == 0 {
		return nil, fmt.Errorf("starter: catalog.json contains no categories")
	}
	if len(c.Templates) == 0 {
		return nil, fmt.Errorf("starter: catalog.json contains no templates")
	}
	return &c, nil
}
