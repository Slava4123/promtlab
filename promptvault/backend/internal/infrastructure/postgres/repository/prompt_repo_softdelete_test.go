package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
)

// TestPromptRepo_GetByID_SoftDeletedNotReturned фиксирует встроенную GORM
// soft-delete защиту через тип models.Prompt.DeletedAt = gorm.DeletedAt.
// First(&prompt, id) автоматически добавляет WHERE deleted_at IS NULL,
// поэтому после SoftDelete возвращается ErrNotFound (а не архивная запись).
// Регрессия ловит случайный переход типа поля на *time.Time / разбивку
// этого поведения через Unscoped() где не должно быть.
func TestPromptRepo_GetByID_SoftDeletedNotReturned(t *testing.T) {
	db := setupTestDB(t)
	r := NewPromptRepository(db)
	ctx := context.Background()

	u := createTestUser(t, db, "softdelete-prompt@test.local")
	p := createTestPrompt(t, db, u.ID, nil, "to be deleted", "content", 0)

	// Убедимся что запись доступна до удаления.
	got, err := r.GetByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, p.ID, got.ID)

	require.NoError(t, r.SoftDelete(ctx, p.ID))

	// После soft-delete GetByID должен вернуть ErrNotFound.
	got2, err := r.GetByID(ctx, p.ID)
	assert.Nil(t, got2)
	require.Error(t, err)
	assert.True(t, errors.Is(err, repo.ErrNotFound),
		"soft-deleted prompt не должен возвращаться через GetByID, got: %v", err)
}
