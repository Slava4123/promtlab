package repository

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

func newAuditRepoTest(t *testing.T) (repo.AuditRepository, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	return NewAuditRepository(db), db
}

// applyAppendOnlyGuard реплицирует триггер-защиту из миграции 000018 в
// AutoMigrate-окружении testcontainers. Без этого тесты append-only
// не имели бы смысла — AutoMigrate создаёт «голую» таблицу без триггеров.
// Trigger-based защита выбрана потому что REVOKE не работает для owner таблицы.
func applyAppendOnlyGuard(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`
		CREATE OR REPLACE FUNCTION prevent_audit_log_modification() RETURNS TRIGGER AS $$
		BEGIN
		    RAISE EXCEPTION 'audit_log is append-only: % operations are not allowed', TG_OP
		        USING ERRCODE = 'insufficient_privilege';
		END;
		$$ LANGUAGE plpgsql;
	`).Error)
	require.NoError(t, db.Exec(`DROP TRIGGER IF EXISTS audit_log_prevent_update ON audit_log`).Error)
	require.NoError(t, db.Exec(`DROP TRIGGER IF EXISTS audit_log_prevent_delete ON audit_log`).Error)
	require.NoError(t, db.Exec(`
		CREATE TRIGGER audit_log_prevent_update
		    BEFORE UPDATE ON audit_log
		    FOR EACH STATEMENT
		    EXECUTE FUNCTION prevent_audit_log_modification()
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TRIGGER audit_log_prevent_delete
		    BEFORE DELETE ON audit_log
		    FOR EACH STATEMENT
		    EXECUTE FUNCTION prevent_audit_log_modification()
	`).Error)
}

func makeEntry(adminID uint, action string) *models.AuditLog {
	payload, _ := json.Marshal(map[string]any{"ok": true})
	return &models.AuditLog{
		AdminID:     adminID,
		Action:      action,
		TargetType:  "user",
		TargetID:    uint_ptrAudit(42),
		BeforeState: nil,
		AfterState:  payload,
		IP:          "127.0.0.1",
		UserAgent:   "test-suite/1.0",
	}
}

func uint_ptrAudit(v uint) *uint { return &v }

func TestAuditRepo_Log_Insert(t *testing.T) {
	r, db := newAuditRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "audit-insert@test.local")

	entry := makeEntry(u.ID, "grant_badge")
	require.NoError(t, r.Log(ctx, entry))
	assert.NotZero(t, entry.ID, "ID должен быть заполнен после INSERT")
	assert.False(t, entry.CreatedAt.IsZero())
}

func TestAuditRepo_List_FiltersByAdmin(t *testing.T) {
	r, db := newAuditRepoTest(t)
	ctx := context.Background()
	admin1 := createTestUser(t, db, "audit-admin1@test.local")
	admin2 := createTestUser(t, db, "audit-admin2@test.local")

	require.NoError(t, r.Log(ctx, makeEntry(admin1.ID, "grant_badge")))
	require.NoError(t, r.Log(ctx, makeEntry(admin1.ID, "revoke_badge")))
	require.NoError(t, r.Log(ctx, makeEntry(admin2.ID, "freeze_user")))

	// Без фильтра — все 3.
	all, total, err := r.List(ctx, repo.AuditLogFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, all, 3)

	// По admin1 — 2.
	byAdmin, total, err := r.List(ctx, repo.AuditLogFilter{AdminID: &admin1.ID})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, byAdmin, 2)
}

func TestAuditRepo_List_FiltersByAction(t *testing.T) {
	r, db := newAuditRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "audit-action@test.local")

	require.NoError(t, r.Log(ctx, makeEntry(u.ID, "grant_badge")))
	require.NoError(t, r.Log(ctx, makeEntry(u.ID, "grant_badge")))
	require.NoError(t, r.Log(ctx, makeEntry(u.ID, "freeze_user")))

	list, total, err := r.List(ctx, repo.AuditLogFilter{Action: "grant_badge"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	for _, e := range list {
		assert.Equal(t, "grant_badge", e.Action)
	}
}

func TestAuditRepo_List_DateRange(t *testing.T) {
	r, db := newAuditRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "audit-daterange@test.local")

	require.NoError(t, r.Log(ctx, makeEntry(u.ID, "grant_badge")))

	// 24h назад — должен включать нашу запись.
	from := time.Now().Add(-24 * time.Hour)
	list, total, err := r.List(ctx, repo.AuditLogFilter{FromTime: &from})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)

	// From в будущем — ничего.
	future := time.Now().Add(24 * time.Hour)
	list, total, err = r.List(ctx, repo.AuditLogFilter{FromTime: &future})
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, list)
}

func TestAuditRepo_List_Pagination(t *testing.T) {
	r, db := newAuditRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "audit-page@test.local")

	for range 7 {
		require.NoError(t, r.Log(ctx, makeEntry(u.ID, "grant_badge")))
	}

	// Page 1, size 3.
	list, total, err := r.List(ctx, repo.AuditLogFilter{Page: 1, PageSize: 3})
	require.NoError(t, err)
	assert.Equal(t, int64(7), total)
	assert.Len(t, list, 3)

	// Page 3 — одна оставшаяся запись.
	list, total, err = r.List(ctx, repo.AuditLogFilter{Page: 3, PageSize: 3})
	require.NoError(t, err)
	assert.Equal(t, int64(7), total)
	assert.Len(t, list, 1)
}

// TestAuditLog_UpdateRejected — критический тест append-only semantics.
// Проверяет что после REVOKE UPDATE попытка изменить historical запись
// возвращает permission denied (SQLSTATE 42501). Это compliance gate:
// если тест упадёт, audit log больше не append-only.
func TestAuditLog_UpdateRejected(t *testing.T) {
	r, db := newAuditRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "audit-update@test.local")

	require.NoError(t, r.Log(ctx, makeEntry(u.ID, "grant_badge")))

	// Пока таблица со всеми правами — UPDATE работает.
	// Применяем REVOKE из миграции 000018 прямо здесь.
	applyAppendOnlyGuard(t, db)

	err := db.Exec("UPDATE audit_log SET action = 'hacked' WHERE admin_id = ?", u.ID).Error
	require.Error(t, err, "UPDATE должен быть отклонён после REVOKE")
	assert.True(t,
		strings.Contains(err.Error(), "append-only") ||
			strings.Contains(err.Error(), "insufficient_privilege") ||
			strings.Contains(err.Error(), "SQLSTATE 42501"),
		"ошибка должна указывать на append-only trigger, got: %v", err,
	)
}

// TestAuditLog_DeleteRejected — симметричный тест для DELETE.
func TestAuditLog_DeleteRejected(t *testing.T) {
	r, db := newAuditRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "audit-delete@test.local")

	require.NoError(t, r.Log(ctx, makeEntry(u.ID, "grant_badge")))

	applyAppendOnlyGuard(t, db)

	err := db.Exec("DELETE FROM audit_log WHERE admin_id = ?", u.ID).Error
	require.Error(t, err, "DELETE должен быть отклонён после REVOKE")
	assert.True(t,
		strings.Contains(err.Error(), "append-only") ||
			strings.Contains(err.Error(), "insufficient_privilege") ||
			strings.Contains(err.Error(), "SQLSTATE 42501"),
		"ошибка должна указывать на append-only trigger, got: %v", err,
	)
}

// TestAuditLog_InsertStillWorksAfterRevoke — убедиться что REVOKE не сломал
// insert-ы (нужно именно INSERT + SELECT, UPDATE/DELETE отозваны).
func TestAuditLog_InsertStillWorksAfterRevoke(t *testing.T) {
	r, db := newAuditRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "audit-insertafter@test.local")

	applyAppendOnlyGuard(t, db)

	require.NoError(t, r.Log(ctx, makeEntry(u.ID, "grant_badge")))

	list, total, err := r.List(ctx, repo.AuditLogFilter{AdminID: &u.ID})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}
