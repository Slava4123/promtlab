package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"promptvault/internal/models"
)

// TestPromptMergeWith_HappyPath — оба промпта принадлежат юзеру:
// MergeWith возвращает nil; keep остаётся активным (видим через GetByID),
// merge soft-deleted (deleted_at NOT NULL, GetByID возвращает ErrNotFound).
func TestPromptMergeWith_HappyPath(t *testing.T) {
	db := setupTestDB(t)
	r := NewPromptRepository(db)
	ctx := context.Background()

	u := createTestUser(t, db, "merge-happy@test.local")
	keep := createTestPrompt(t, db, u.ID, nil, "keeper", "keep content", 0)
	merge := createTestPrompt(t, db, u.ID, nil, "duplicate", "merge content", 0)

	require.NoError(t, r.MergeWith(ctx, keep.ID, merge.ID, u.ID))

	// Keep — активен.
	gotKeep, err := r.GetByID(ctx, keep.ID)
	require.NoError(t, err)
	assert.Equal(t, keep.ID, gotKeep.ID)

	// Merge — soft-deleted: GetByID отдаёт ErrNotFound (GORM авто-фильтр).
	_, err = r.GetByID(ctx, merge.ID)
	require.Error(t, err)

	// Дополнительная проверка через Unscoped: запись существует и deleted_at не null.
	var raw models.Prompt
	require.NoError(t, db.Unscoped().First(&raw, merge.ID).Error)
	assert.True(t, raw.DeletedAt.Valid, "merge prompt должен иметь deleted_at != NULL")
}

// TestPromptMergeWith_OwnershipError — merge принадлежит другому юзеру.
// MergeWith должен вернуть gorm.ErrRecordNotFound и НЕ удалить запись.
func TestPromptMergeWith_OwnershipError(t *testing.T) {
	db := setupTestDB(t)
	r := NewPromptRepository(db)
	ctx := context.Background()

	owner := createTestUser(t, db, "merge-owner@test.local")
	other := createTestUser(t, db, "merge-stranger@test.local")

	keep := createTestPrompt(t, db, owner.ID, nil, "owned by user A", "k", 0)
	mergeOther := createTestPrompt(t, db, other.ID, nil, "owned by user B", "m", 0)

	err := r.MergeWith(ctx, keep.ID, mergeOther.ID, owner.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound),
		"ownership mismatch должен вернуть gorm.ErrRecordNotFound, got: %v", err)

	// Merge prompt не должен быть удалён (TOCTOU защита через транзакцию).
	got, err := r.GetByID(ctx, mergeOther.ID)
	require.NoError(t, err)
	assert.Equal(t, mergeOther.ID, got.ID)
}

// TestPromptMergeWith_SelfMerge — keepID == mergeID должен вернуть ошибку.
func TestPromptMergeWith_SelfMerge(t *testing.T) {
	db := setupTestDB(t)
	r := NewPromptRepository(db)
	ctx := context.Background()

	u := createTestUser(t, db, "merge-self@test.local")
	p := createTestPrompt(t, db, u.ID, nil, "alone", "c", 0)

	err := r.MergeWith(ctx, p.ID, p.ID, u.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "itself")

	// Промпт всё ещё активен.
	got, err := r.GetByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, p.ID, got.ID)
}
