package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"promptvault/internal/models"
)

type quotaRepo struct {
	db *gorm.DB
}

func NewQuotaRepository(db *gorm.DB) *quotaRepo {
	return &quotaRepo{db: db}
}

// CountPersonalPrompts — соло-библиотека юзера (team_id IS NULL).
// Командные промпты учитываются отдельно через CountTeamPrompts.
func (r *quotaRepo) CountPersonalPrompts(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Prompt{}).
		Where("user_id = ? AND team_id IS NULL AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// CountTeamPrompts — pool команды. Считаются ВСЕ промпты команды независимо
// от того, кто из участников их создал. Используется в CheckTeamPromptQuota
// против плана owner'а команды.
func (r *quotaRepo) CountTeamPrompts(ctx context.Context, teamID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Prompt{}).
		Where("team_id = ? AND deleted_at IS NULL", teamID).
		Count(&count).Error
	return count, err
}

// CountPersonalCollections — личные коллекции юзера (team_id IS NULL).
func (r *quotaRepo) CountPersonalCollections(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Collection{}).
		Where("user_id = ? AND team_id IS NULL AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// CountTeamCollections — pool коллекций команды.
func (r *quotaRepo) CountTeamCollections(ctx context.Context, teamID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Collection{}).
		Where("team_id = ? AND deleted_at IS NULL", teamID).
		Count(&count).Error
	return count, err
}

func (r *quotaRepo) CountTeamsOwned(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Team{}).
		Where("created_by = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *quotaRepo) CountActiveShareLinks(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ShareLink{}).
		Where("user_id = ? AND is_active = true", userID).
		Count(&count).Error
	return count, err
}

func (r *quotaRepo) CountTeamMembers(ctx context.Context, teamID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.TeamMember{}).
		Where("team_id = ?", teamID).
		Count(&count).Error
	return int(count), err
}

func (r *quotaRepo) GetDailyUsage(ctx context.Context, userID uint, date time.Time, featureType string) (int, error) {
	var usage models.DailyFeatureUsage
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND usage_date = ? AND feature_type = ?", userID, date.Format("2006-01-02"), featureType).
		First(&usage).Error
	if err != nil {
		// MN-25: errors.Is вместо bare == — на случай wrapping через
		// fmt.Errorf("scope: %w", err) в будущих рефакторах.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return usage.Count, nil
}

func (r *quotaRepo) GetTotalUsage(ctx context.Context, userID uint, featureType string) (int, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.DailyFeatureUsage{}).
		Where("user_id = ? AND feature_type = ?", userID, featureType).
		Select("COALESCE(SUM(count), 0)").
		Scan(&total).Error
	return int(total), err
}

func (r *quotaRepo) IncrementDailyUsage(ctx context.Context, userID uint, date time.Time, featureType string) error {
	return r.db.WithContext(ctx).Exec(
		`INSERT INTO daily_feature_usage (user_id, usage_date, feature_type, count)
		 VALUES (?, ?, ?, 1)
		 ON CONFLICT (user_id, usage_date, feature_type)
		 DO UPDATE SET count = daily_feature_usage.count + 1`,
		userID, date.Format("2006-01-02"), featureType,
	).Error
}

// DeleteOldDailyUsage удаляет строки daily_feature_usage старше olderThanDays.
// Read-path использует только сегодняшний день — всё остальное мёртвый balast.
// Без cleanup таблица растёт ~N юзеров × M фич × дни ≈ десятки миллионов строк
// в год на средней базе.
//
// SQL: используем make_interval(days => ?) вместо string-concat (... || ' days')
// — pgx не умеет encode int в text при конкатенации, валится с
// "cannot find encode plan" (OID 25). make_interval принимает integer напрямую.
func (r *quotaRepo) DeleteOldDailyUsage(ctx context.Context, olderThanDays int) (int64, error) {
	res := r.db.WithContext(ctx).Exec(
		"DELETE FROM daily_feature_usage WHERE usage_date < CURRENT_DATE - make_interval(days => ?)",
		olderThanDays,
	)
	return res.RowsAffected, res.Error
}

// CountPersonalChains — личные цепочки юзера (team_id IS NULL).
func (r *quotaRepo) CountPersonalChains(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.PromptChain{}).
		Where("user_id = ? AND team_id IS NULL AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// CountTeamChains — pool цепочек команды.
func (r *quotaRepo) CountTeamChains(ctx context.Context, teamID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.PromptChain{}).
		Where("team_id = ? AND deleted_at IS NULL", teamID).
		Count(&count).Error
	return count, err
}

func (r *quotaRepo) CountStepsByChain(ctx context.Context, chainID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.PromptChainStep{}).
		Where("chain_id = ?", chainID).
		Count(&count).Error
	return count, err
}
