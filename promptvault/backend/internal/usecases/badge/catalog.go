package badge

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed catalog.json
var catalogBytes []byte

// catalogFile — корневой узел catalog.json.
type catalogFile struct {
	Badges []Badge `json:"badges"`
}

// LoadCatalog парсит embedded catalog.json и валидирует каждый бейдж.
// Вызывается один раз при старте приложения из app.go → NewService.
// Любая ошибка здесь — fatal для приложения (как и в starter/changelog).
func LoadCatalog() ([]Badge, error) {
	var f catalogFile
	if err := json.Unmarshal(catalogBytes, &f); err != nil {
		return nil, fmt.Errorf("%w: parse json: %v", ErrCatalogLoad, err)
	}
	if len(f.Badges) == 0 {
		return nil, fmt.Errorf("%w: empty badges list", ErrCatalogLoad)
	}

	// Валидация: уникальность ID, непустые поля, известный ConditionType,
	// непустые Triggers, Threshold > 0, MinVersions согласован с Type.
	seen := make(map[string]struct{}, len(f.Badges))
	for i, b := range f.Badges {
		if b.ID == "" {
			return nil, fmt.Errorf("%w: badge[%d] has empty id", ErrCatalogLoad, i)
		}
		if _, dup := seen[b.ID]; dup {
			return nil, fmt.Errorf("%w: duplicate badge id %q", ErrCatalogLoad, b.ID)
		}
		seen[b.ID] = struct{}{}

		if b.Title == "" || b.Description == "" || b.Icon == "" {
			return nil, fmt.Errorf("%w: badge %q missing title/description/icon", ErrCatalogLoad, b.ID)
		}
		if b.Category == "" {
			return nil, fmt.Errorf("%w: badge %q missing category", ErrCatalogLoad, b.ID)
		}
		if len(b.Triggers) == 0 {
			return nil, fmt.Errorf("%w: badge %q has no triggers", ErrCatalogLoad, b.ID)
		}
		if b.Condition.Threshold <= 0 {
			return nil, fmt.Errorf("%w: badge %q has non-positive threshold", ErrCatalogLoad, b.ID)
		}
		if !isKnownCondition(b.Condition.Type) {
			return nil, fmt.Errorf("%w: badge %q: %w: %q", ErrCatalogLoad, b.ID, ErrUnknownCondition, b.Condition.Type)
		}
		// MinVersions применим только к versioned_prompt_count.
		if b.Condition.Type == CondVersionedPromptCount {
			if b.Condition.MinVersions <= 0 {
				return nil, fmt.Errorf("%w: badge %q requires positive min_versions", ErrCatalogLoad, b.ID)
			}
		} else if b.Condition.MinVersions != 0 {
			return nil, fmt.Errorf("%w: badge %q: min_versions only allowed with versioned_prompt_count", ErrCatalogLoad, b.ID)
		}
	}

	return f.Badges, nil
}

func isKnownCondition(t ConditionType) bool {
	switch t {
	case CondSoloPromptCount, CondTeamPromptCount, CondTotalPromptCount,
		CondSoloCollectionCount, CondTeamCollectionCount,
		CondTotalUsage, CondVersionedPromptCount, CondCurrentStreak:
		return true
	}
	return false
}
