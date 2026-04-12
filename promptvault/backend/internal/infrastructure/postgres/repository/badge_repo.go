package repository

import (
	"context"
	"strings"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type badgeRepo struct {
	db *gorm.DB
}

// NewBadgeRepository возвращает GORM-реализацию BadgeRepository.
func NewBadgeRepository(db *gorm.DB) repo.BadgeRepository {
	return &badgeRepo{db: db}
}

func (r *badgeRepo) Unlock(ctx context.Context, userID uint, badgeID string) error {
	badge := &models.UserBadge{
		UserID:  userID,
		BadgeID: badgeID,
	}
	err := r.db.WithContext(ctx).Create(badge).Error
	if err == nil {
		return nil
	}
	if isUniqueViolation(err) {
		return repo.ErrBadgeAlreadyUnlocked
	}
	return err
}

func (r *badgeRepo) UnlockedIDs(ctx context.Context, userID uint) (map[string]struct{}, error) {
	var ids []string
	err := r.db.WithContext(ctx).
		Model(&models.UserBadge{}).
		Where("user_id = ?", userID).
		Pluck("badge_id", &ids).Error
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set, nil
}

func (r *badgeRepo) ListByUser(ctx context.Context, userID uint) ([]models.UserBadge, error) {
	var badges []models.UserBadge
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("unlocked_at DESC").
		Find(&badges).Error
	return badges, err
}

func (r *badgeRepo) DeleteByUserAndBadge(ctx context.Context, userID uint, badgeID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND badge_id = ?", userID, badgeID).
		Delete(&models.UserBadge{}).Error
}

// --- aggregation methods ---
// Все Count/Sum работают через raw Table(...) чтобы явно контролировать
// фильтр deleted_at IS NULL (при Table GORM не применяет soft-delete scope).

func (r *badgeRepo) CountSoloPrompts(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Table("prompts").
		Where("user_id = ? AND team_id IS NULL AND deleted_at IS NULL", userID).
		Count(&n).Error
	return n, err
}

func (r *badgeRepo) CountTeamPrompts(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Table("prompts").
		Where("user_id = ? AND team_id IS NOT NULL AND deleted_at IS NULL", userID).
		Count(&n).Error
	return n, err
}

func (r *badgeRepo) CountAllPrompts(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Table("prompts").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&n).Error
	return n, err
}

func (r *badgeRepo) CountSoloCollections(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Table("collections").
		Where("user_id = ? AND team_id IS NULL AND deleted_at IS NULL", userID).
		Count(&n).Error
	return n, err
}

func (r *badgeRepo) CountTeamCollections(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Table("collections").
		Where("user_id = ? AND team_id IS NOT NULL AND deleted_at IS NULL", userID).
		Count(&n).Error
	return n, err
}

func (r *badgeRepo) SumUsage(ctx context.Context, userID uint) (int64, error) {
	var sum int64
	err := r.db.WithContext(ctx).
		Table("prompts").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Select("COALESCE(SUM(usage_count), 0)").
		Scan(&sum).Error
	return sum, err
}

func (r *badgeRepo) CountVersionedPrompts(ctx context.Context, userID uint, minVersions int) (int64, error) {
	// Промпты юзера, у которых >= minVersions записей в prompt_versions.
	// Используется только при событии prompt_updated (badge.Service) со short-circuit
	// когда у текущего промпта < minVersions версий — expensive subquery изолирован здесь.
	const sql = `
		SELECT COUNT(*)
		FROM prompts p
		WHERE p.user_id = ?
		  AND p.deleted_at IS NULL
		  AND (
		    SELECT COUNT(*) FROM prompt_versions pv WHERE pv.prompt_id = p.id
		  ) >= ?
	`
	var n int64
	err := r.db.WithContext(ctx).Raw(sql, userID, minVersions).Scan(&n).Error
	return n, err
}

// isUniqueViolation проверяет, является ли ошибка PostgreSQL unique constraint
// violation (SQLSTATE 23505). Используется в Unlock для race-safe detection
// дубликатов. Строковая проверка выбрана, чтобы не тащить direct import
// pgx/v5/pgconn (сейчас indirect в go.mod).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23505") ||
		strings.Contains(msg, "duplicate key value violates unique constraint")
}
