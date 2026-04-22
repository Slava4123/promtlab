package team

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"promptvault/internal/models"
)

// TestCanSeeActorEmail фиксирует матрицу доступов для GDPR-маскирования
// (Phase 14 Q1, вариант C). Owner/editor видят полный email,
// viewer/пустая роль — получают маску в toActivityItemResponses.
func TestCanSeeActorEmail(t *testing.T) {
	assert.True(t, canSeeActorEmail(models.RoleOwner), "owner видит полный email")
	assert.True(t, canSeeActorEmail(models.RoleEditor), "editor видит полный email")
	assert.False(t, canSeeActorEmail(models.RoleViewer), "viewer не видит raw email (получает маску)")
	assert.False(t, canSeeActorEmail(""), "пустая роль — по умолчанию без доступа")
}

// TestToActivityItemResponses_MaskingMatrix — конечный-к-концу контракт
// response DTO для viewer и owner на одной же входной выборке событий.
// Страхует регрессию маскирования при рефакторинге utils.MaskEmail или
// self.go преобразования.
func TestToActivityItemResponses_MaskingMatrix(t *testing.T) {
	events := []models.TeamActivityLog{
		{
			ID:         1,
			ActorID:    ptrUint(11),
			ActorName:  "Alice",
			ActorEmail: "alice@acme.com",
			EventType:  "prompt.created",
			TargetType: "prompt",
			TargetID:   ptrUint(100),
			TargetLabel: "hello",
		},
	}

	owner := toActivityItemResponses(events, models.RoleOwner)
	require := assert.New(t)
	require.Equal("alice@acme.com", owner[0].ActorEmail, "owner видит raw email")

	viewer := toActivityItemResponses(events, models.RoleViewer)
	require.Equal("a***@acme.com", viewer[0].ActorEmail, "viewer получает маску (вариант C)")
	require.Equal("Alice", viewer[0].ActorName, "name остаётся для context")

	// Пустая роль (edge case) — маска, не raw.
	empty := toActivityItemResponses(events, "")
	require.Equal("a***@acme.com", empty[0].ActorEmail)
}

// TestToActivityItemResponses_EmptyEmail — nil-safe masking при пустом email.
func TestToActivityItemResponses_EmptyEmail(t *testing.T) {
	events := []models.TeamActivityLog{{ID: 2, ActorName: "Bot"}}
	items := toActivityItemResponses(events, models.RoleViewer)
	assert.Equal(t, "", items[0].ActorEmail, "пустой email → пустая маска")
}

func ptrUint(v uint) *uint { return &v }
