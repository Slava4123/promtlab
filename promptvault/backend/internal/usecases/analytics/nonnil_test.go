package analytics

import (
	"testing"

	"github.com/stretchr/testify/assert"

	repo "promptvault/internal/interface/repository"
)

// Регрессия-тест на BUG #1 (QA-сессия 2026-04-22):
// GORM Scan возвращал nil-slice для пустого результата → JSON маршалил `null`
// → frontend AnalyticsPage падал на `.reduce` of null.
// Теперь Service.* вызывает ensure*NonNil перед возвратом — нормализуем
// API-контракт: массивы всегда есть, пусть даже пустые.

func TestEnsurePersonalNonNil_AllNilToEmpty(t *testing.T) {
	d := &PersonalDashboard{}
	ensurePersonalNonNil(d)
	assert.NotNil(t, d.UsagePerDay, "UsagePerDay must not be nil")
	assert.NotNil(t, d.TopPrompts)
	assert.NotNil(t, d.PromptsCreated)
	assert.NotNil(t, d.PromptsUpdated)
	assert.NotNil(t, d.ShareViews)
	assert.NotNil(t, d.TopShared)
	assert.Len(t, d.UsagePerDay, 0, "empty, not nil")
	assert.Len(t, d.TopPrompts, 0)
	assert.Len(t, d.PromptsCreated, 0)
	assert.Len(t, d.PromptsUpdated, 0)
	assert.Len(t, d.ShareViews, 0)
	assert.Len(t, d.TopShared, 0)
}

func TestEnsurePersonalNonNil_PreservesExistingData(t *testing.T) {
	d := &PersonalDashboard{
		UsagePerDay: []repo.UsagePoint{{Count: 5}},
		TopPrompts:  []repo.PromptUsageRow{{PromptID: 42, Uses: 3}},
		// ShareViews nil — должен стать []
	}
	ensurePersonalNonNil(d)
	assert.Len(t, d.UsagePerDay, 1, "existing data preserved")
	assert.Equal(t, int64(5), d.UsagePerDay[0].Count)
	assert.Len(t, d.TopPrompts, 1)
	assert.Equal(t, uint(42), d.TopPrompts[0].PromptID)
	assert.NotNil(t, d.ShareViews, "nil field normalized")
	assert.Len(t, d.ShareViews, 0)
}

func TestEnsureTeamNonNil_AllNilToEmpty(t *testing.T) {
	d := &TeamDashboard{}
	ensureTeamNonNil(d)
	assert.NotNil(t, d.UsagePerDay)
	assert.NotNil(t, d.TopPrompts)
	assert.NotNil(t, d.PromptsCreated)
	assert.NotNil(t, d.PromptsUpdated)
	assert.NotNil(t, d.Contributors)
}

func TestEnsurePromptNonNil_AllNilToEmpty(t *testing.T) {
	p := &PromptAnalytics{PromptID: 1}
	ensurePromptNonNil(p)
	assert.NotNil(t, p.UsagePerDay)
	assert.NotNil(t, p.ShareViewsPerDay)
}
