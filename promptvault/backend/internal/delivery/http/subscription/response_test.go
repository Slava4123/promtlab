package subscription

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"promptvault/internal/models"
)

// Регрессия на QA-баг: все тарифы показывали "undefined share-ссылок/день"
// на /pricing — backend DTO не копировал MaxDailyShares и IsActive из модели.
// Тест гарантирует что JSON-контракт больше не теряет эти поля.

func TestNewPlanResponse_CopiesAllFields(t *testing.T) {
	plan := models.SubscriptionPlan{
		ID:              "pro",
		Name:            "Pro",
		PriceKop:        59900,
		PeriodDays:      30,
		MaxPrompts:      500,
		MaxCollections:  100,
		MaxTeams:        5,
		MaxTeamMembers:  10,
		MaxShareLinks:   50,
		MaxDailyShares:  100,
		MaxExtUsesDaily: 100,
		MaxMCPUsesDaily: 100,
		Features:        json.RawMessage(`["priority_support"]`),
		SortOrder:       1,
		IsActive:        true,
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
	assert.Equal(t, 50, resp.MaxShareLinks)
	assert.Equal(t, 100, resp.MaxDailyShares, "MaxDailyShares обязателен — баг /pricing #1")
	assert.Equal(t, 100, resp.MaxExtUsesDaily)
	assert.Equal(t, 100, resp.MaxMCPUsesDaily)
	assert.Equal(t, 1, resp.SortOrder)
	assert.True(t, resp.IsActive, "IsActive обязателен для фильтрации на фронте — баг /pricing #1")
}

func TestNewPlanResponse_JSONContainsNewFields(t *testing.T) {
	plan := models.SubscriptionPlan{
		ID:              "max",
		Name:            "Max",
		MaxDailyShares:  1000,
		IsActive:        true,
		Features:        json.RawMessage(`[]`),
	}

	resp := NewPlanResponse(plan)
	raw, err := json.Marshal(resp)
	assert.NoError(t, err)

	payload := string(raw)
	// Frontend (pricing.tsx) ожидает эти ключи буквально. Если у DTO их не
	// будет — plan.max_daily_shares === undefined и .toLocaleString упадёт.
	assert.Contains(t, payload, `"max_daily_shares":1000`)
	assert.Contains(t, payload, `"is_active":true`)
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
