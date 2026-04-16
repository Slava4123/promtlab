package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type streakRepo struct {
	db *gorm.DB
}

func NewStreakRepository(db *gorm.DB) repo.StreakRepository {
	return &streakRepo{db: db}
}

func (r *streakRepo) RecordActivity(ctx context.Context, userID uint, today string) error {
	sql := `
		INSERT INTO user_streaks (user_id, current_streak, longest_streak, last_active_date, updated_at)
		VALUES (?, 1, 1, ?, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			current_streak = CASE
				WHEN user_streaks.last_active_date = ? THEN user_streaks.current_streak
				WHEN user_streaks.last_active_date = ?::date - INTERVAL '1 day' THEN user_streaks.current_streak + 1
				ELSE 1
			END,
			longest_streak = GREATEST(
				user_streaks.longest_streak,
				CASE
					WHEN user_streaks.last_active_date = ? THEN user_streaks.current_streak
					WHEN user_streaks.last_active_date = ?::date - INTERVAL '1 day' THEN user_streaks.current_streak + 1
					ELSE 1
				END
			),
			last_active_date = ?,
			updated_at = NOW()
	`
	return r.db.WithContext(ctx).Exec(sql, userID, today, today, today, today, today, today).Error
}

func (r *streakRepo) GetByUserID(ctx context.Context, userID uint) (*models.UserStreak, error) {
	var streak models.UserStreak
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&streak).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repo.ErrNotFound
	}
	return &streak, err
}

func (r *streakRepo) ListAtRisk(ctx context.Context, todayUTC string, minStreak int) ([]models.UserStreak, error) {
	var out []models.UserStreak
	// At-risk: current_streak > minStreak (что-то стоящее сохранять),
	// last_active_date < today (сегодня ещё не заходили),
	// reminder_sent_on IS NULL OR < today (не слали сегодня).
	err := r.db.WithContext(ctx).
		Where("current_streak > ?", minStreak).
		Where("last_active_date < ?::date", todayUTC).
		Where("reminder_sent_on IS NULL OR reminder_sent_on < ?::date", todayUTC).
		Limit(500).
		Find(&out).Error
	return out, err
}

func (r *streakRepo) MarkReminderSent(ctx context.Context, userID uint, todayUTC string) error {
	return r.db.WithContext(ctx).
		Model(&models.UserStreak{}).
		Where("user_id = ?", userID).
		Update("reminder_sent_on", todayUTC).Error
}
