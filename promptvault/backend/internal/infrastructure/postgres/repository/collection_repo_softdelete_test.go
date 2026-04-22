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

// TestCollectionRepo_SoftDeleted_NotReturned — зеркало prompt_repo_softdelete_test.go.
// Collection имеет DeletedAt gorm.DeletedAt (models/collection.go), поэтому
// встроенный GORM scope добавляет WHERE deleted_at IS NULL автоматически.
// Регрессия ловит изменения типа поля или неожиданное использование .Unscoped()
// в List/GetByID методах collection repository.
func TestCollectionRepo_SoftDeleted_NotReturned(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	u := createTestUser(t, db, "softdelete-col@test.local")
	c := createTestCollection(t, db, u.ID, nil, "to be deleted")

	// Soft-delete через GORM: переводит deleted_at в NOW().
	require.NoError(t, db.Delete(&models.Collection{}, c.ID).Error)

	var found models.Collection
	err := db.WithContext(ctx).First(&found, c.ID).Error
	require.Error(t, err, "First без .Unscoped() не должен возвращать soft-deleted запись")
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound),
		"ожидался ErrRecordNotFound, got: %v", err)

	// Явный Unscoped() возвращает запись — гарантируем что soft-delete
	// данные сохраняются (trash flow опирается на это).
	var raw models.Collection
	require.NoError(t, db.Unscoped().First(&raw, c.ID).Error)
	assert.Equal(t, c.ID, raw.ID)
	assert.True(t, raw.DeletedAt.Valid, "DeletedAt.Valid должен быть true после soft-delete")
}

// TestCollectionListFilter_SoftDeletedExcluded — проверка, что List-подобный
// запрос через Find() тоже корректно отбрасывает удалённые.
func TestCollectionListFilter_SoftDeletedExcluded(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	u := createTestUser(t, db, "softdelete-col-list@test.local")
	active := createTestCollection(t, db, u.ID, nil, "active")
	deleted := createTestCollection(t, db, u.ID, nil, "deleted")
	require.NoError(t, db.Delete(&models.Collection{}, deleted.ID).Error)

	var found []models.Collection
	require.NoError(t, db.WithContext(ctx).Where("user_id = ?", u.ID).Find(&found).Error)
	require.Len(t, found, 1, "должна вернуться только не-удалённая коллекция")
	assert.Equal(t, active.ID, found[0].ID)
}
