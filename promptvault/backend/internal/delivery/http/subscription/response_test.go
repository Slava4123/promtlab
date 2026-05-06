package subscription

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"promptvault/internal/models"
	quotauc "promptvault/internal/usecases/quota"
)

// Регрессия на QA-баг: все тарифы показывали "undefined share-ссылок/день"
// на /pricing — backend DTO не копировал MaxDailyShares и IsActive из модели.
// Тест гарантирует что JSON-контракт больше не теряет эти поля.

func TestNewPlanResponse_CopiesAllFields(t *testing.T) {
	plan := models.SubscriptionPlan{
		ID:                 "pro",
		Name:               "Pro",
		PriceKop:           59900,
		PeriodDays:         30,
		MaxPrompts:         500,
		MaxCollections:     100,
		MaxTeams:           5,
		MaxTeamMembers:     10,
		MaxExtUsesDaily:    100,
		MaxMCPUsesDaily:    100,
		MaxChains:          5,
		MaxStepsPerChain:   10,
		MaxSavedExecutions: 10,
		Features:           json.RawMessage(`["priority_support"]`),
		SortOrder:          1,
		IsActive:           true,
	}

	resp := NewPlanResponse(plan)

	assert.Equal(t, "pro", resp.ID)
	assert.Equal(t, "Pro", resp.Name)
	assert.Equal(t, 59900, resp.PriceKop)
	assert.Equal(t, 30, resp.PeriodDays)
	assert.Equal(t, 500, resp.MaxPrompts)
	assert.Equal(t, 100, resp.MaxCollections)
	assert.Equal(t, 5, resp.MaxTeams)
	assert.Equal(t, 10, resp.MaxTeamMembers)
	assert.Equal(t, 100, resp.MaxExtUsesDaily)
	assert.Equal(t, 100, resp.MaxMCPUsesDaily)
	assert.Equal(t, 5, resp.MaxChains, "MaxChains обязателен — Phase 16 dark launch")
	assert.Equal(t, 10, resp.MaxStepsPerChain)
	assert.Equal(t, 10, resp.MaxSavedExecutions)
	assert.Equal(t, 1, resp.SortOrder)
	assert.True(t, resp.IsActive, "IsActive обязателен для фильтрации на фронте — баг /pricing #1")
}

func TestNewPlanResponse_JSONContainsNewFields(t *testing.T) {
	plan := models.SubscriptionPlan{
		ID:                 "max",
		Name:               "Max",
		MaxChains:          100,
		MaxStepsPerChain:   50,
		MaxSavedExecutions: 1000,
		IsActive:           true,
		Features:           json.RawMessage(`[]`),
	}

	resp := NewPlanResponse(plan)
	raw, err := json.Marshal(resp)
	assert.NoError(t, err)

	payload := string(raw)
	assert.Contains(t, payload, `"is_active":true`)
	// Phase 16: chains-лимиты для /pricing под VITE_CHAINS_ENABLED.
	assert.Contains(t, payload, `"max_chains":100`)
	assert.Contains(t, payload, `"max_steps_per_chain":50`)
	assert.Contains(t, payload, `"max_saved_executions":1000`)
	// Phase 16-Y: max_share_links и max_daily_shares УДАЛЕНЫ — не должно быть
	// в payload (фронт больше не показывает счётчики share-ссылок).
	assert.NotContains(t, payload, `"max_share_links"`)
	assert.NotContains(t, payload, `"max_daily_shares"`)
}

// Phase 16-Y: UsageResponse больше не содержит share_links и daily_shares_today.
func TestNewUsageResponse_NoShareCounters(t *testing.T) {
	summary := &quotauc.UsageSummary{
		PlanID:       "pro",
		Prompts:      quotauc.QuotaInfo{Used: 10, Limit: 500},
		Collections:  quotauc.QuotaInfo{Used: 2, Limit: 100},
		Teams:        quotauc.QuotaInfo{Used: 1, Limit: 5},
		ExtUsesToday: quotauc.QuotaInfo{Used: 12, Limit: 100},
		MCPUsesToday: quotauc.QuotaInfo{Used: 4, Limit: 100},
		Chains:       quotauc.QuotaInfo{Used: 2, Limit: 5},
	}

	resp := NewUsageResponse(summary)
	raw, err := json.Marshal(resp)
	assert.NoError(t, err)

	payload := string(raw)
	assert.Contains(t, payload, `"chains":{"used":2,"limit":5}`)
	assert.NotContains(t, payload, `"share_links"`)
	assert.NotContains(t, payload, `"daily_shares_today"`)
}

func TestNewPlansResponse_ReturnsAllPlans(t *testing.T) {
	plans := []models.SubscriptionPlan{
		{ID: "free", Features: json.RawMessage(`[]`)},
		{ID: "pro", Features: json.RawMessage(`[]`)},
		{ID: "max", Features: json.RawMessage(`[]`)},
	}
	resp := NewPlansResponse(plans)
	assert.Len(t, resp, 3)
	assert.Equal(t, "free", resp[0].ID)
	assert.Equal(t, "pro", resp[1].ID)
	assert.Equal(t, "max", resp[2].ID)
}
