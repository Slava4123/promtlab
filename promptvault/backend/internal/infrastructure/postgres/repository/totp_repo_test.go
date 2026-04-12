package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
)

func newTOTPRepoTest(t *testing.T) (repo.TOTPRepository, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	return NewTOTPRepository(db), db
}

func TestTOTPRepo_UpsertEnrollment_Insert(t *testing.T) {
	r, db := newTOTPRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "totp-insert@test.local")

	require.NoError(t, r.UpsertEnrollment(ctx, u.ID, "JBSWY3DPEHPK3PXP"))

	t1, err := r.GetByUserID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "JBSWY3DPEHPK3PXP", t1.Secret)
	assert.Nil(t, t1.ConfirmedAt, "свежая запись должна быть unconfirmed")
}

func TestTOTPRepo_UpsertEnrollment_OverwriteResetsConfirmation(t *testing.T) {
	r, db := newTOTPRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "totp-overwrite@test.local")

	require.NoError(t, r.UpsertEnrollment(ctx, u.ID, "FIRST_SECRET"))
	require.NoError(t, r.MarkConfirmed(ctx, u.ID))

	// Проверяем что после confirmed
	t1, err := r.GetByUserID(ctx, u.ID)
	require.NoError(t, err)
	require.NotNil(t, t1.ConfirmedAt)

	// Re-enroll должен сбросить confirmed_at в NULL и заменить secret
	require.NoError(t, r.UpsertEnrollment(ctx, u.ID, "SECOND_SECRET"))

	t2, err := r.GetByUserID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "SECOND_SECRET", t2.Secret)
	assert.Nil(t, t2.ConfirmedAt, "re-enroll должен сбросить confirmation")
}

func TestTOTPRepo_GetByUserID_NotFound(t *testing.T) {
	r, db := newTOTPRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "totp-notfound@test.local")

	_, err := r.GetByUserID(ctx, u.ID)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestTOTPRepo_MarkConfirmed(t *testing.T) {
	r, db := newTOTPRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "totp-confirm@test.local")

	require.NoError(t, r.UpsertEnrollment(ctx, u.ID, "SECRET"))
	require.NoError(t, r.MarkConfirmed(ctx, u.ID))

	t1, err := r.GetByUserID(ctx, u.ID)
	require.NoError(t, err)
	require.NotNil(t, t1.ConfirmedAt)
	assert.True(t, t1.IsConfirmed())
}

func TestTOTPRepo_Delete_RemovesTOTPAndBackupCodes(t *testing.T) {
	r, db := newTOTPRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "totp-delete@test.local")

	require.NoError(t, r.UpsertEnrollment(ctx, u.ID, "SECRET"))
	require.NoError(t, r.ReplaceBackupCodes(ctx, u.ID, []string{"hash1", "hash2", "hash3"}))

	codes, err := r.ListActiveBackupCodes(ctx, u.ID)
	require.NoError(t, err)
	require.Len(t, codes, 3)

	require.NoError(t, r.Delete(ctx, u.ID))

	_, err = r.GetByUserID(ctx, u.ID)
	assert.ErrorIs(t, err, repo.ErrNotFound)

	codes, err = r.ListActiveBackupCodes(ctx, u.ID)
	require.NoError(t, err)
	assert.Empty(t, codes, "backup codes должны быть удалены вместе с TOTP")
}

func TestTOTPRepo_ReplaceBackupCodes_OverwritesAll(t *testing.T) {
	r, db := newTOTPRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "totp-backup-replace@test.local")

	require.NoError(t, r.ReplaceBackupCodes(ctx, u.ID, []string{"h1", "h2", "h3"}))
	codes, err := r.ListActiveBackupCodes(ctx, u.ID)
	require.NoError(t, err)
	require.Len(t, codes, 3)

	// Regenerate: полная замена, старые не должны остаться.
	require.NoError(t, r.ReplaceBackupCodes(ctx, u.ID, []string{"new1", "new2"}))
	codes, err = r.ListActiveBackupCodes(ctx, u.ID)
	require.NoError(t, err)
	require.Len(t, codes, 2)
	assert.Equal(t, "new1", codes[0].CodeHash)
	assert.Equal(t, "new2", codes[1].CodeHash)
}

func TestTOTPRepo_MarkBackupCodeUsed(t *testing.T) {
	r, db := newTOTPRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "totp-backup-use@test.local")

	require.NoError(t, r.ReplaceBackupCodes(ctx, u.ID, []string{"h1", "h2", "h3"}))

	codes, err := r.ListActiveBackupCodes(ctx, u.ID)
	require.NoError(t, err)
	require.Len(t, codes, 3)

	// Использовать первый код.
	require.NoError(t, r.MarkBackupCodeUsed(ctx, codes[0].ID))

	// Теперь только 2 активных.
	active, err := r.ListActiveBackupCodes(ctx, u.ID)
	require.NoError(t, err)
	assert.Len(t, active, 2)

	// Повторный mark того же кода — идемпотентен (ничего не меняет).
	require.NoError(t, r.MarkBackupCodeUsed(ctx, codes[0].ID))
	active, err = r.ListActiveBackupCodes(ctx, u.ID)
	require.NoError(t, err)
	assert.Len(t, active, 2)
}

func TestTOTPRepo_ListActiveBackupCodes_EmptyForNoEnrollment(t *testing.T) {
	r, db := newTOTPRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "totp-backup-empty@test.local")

	codes, err := r.ListActiveBackupCodes(ctx, u.ID)
	require.NoError(t, err)
	assert.Empty(t, codes)
}
