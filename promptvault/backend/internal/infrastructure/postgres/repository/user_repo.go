package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// ListPaidUsers — ID активных Pro/Max-юзеров (Pricing Iteration v3 Task 7).
// Заменил предыдущий ListMaxUsers (plan_id LIKE 'max%'): теперь explicit IN
// со всеми paid plan_id (pro/pro_yearly/max/max_yearly). Per-plan dispatch
// (Pro → 2 teaser типа, Max → все 7) делается на loop-уровне через
// users.GetByID(uid) + insightsForPlan(planID).
func (r *userRepo) ListPaidUsers(ctx context.Context) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("plan_id IN ? AND status = ?",
			[]string{"pro", "pro_yearly", "max", "max_yearly"}, "active").
		Pluck("id", &ids).Error
	return ids, err
}

func (r *userRepo) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("LOWER(username) = LOWER(?)", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) SearchUsers(ctx context.Context, query string, limit int) ([]models.User, error) {
	escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(query)
	search := "%" + escaped + "%"
	var users []models.User
	err := r.db.WithContext(ctx).
		Where("username != '' AND (username ILIKE ? OR name ILIKE ? OR email ILIKE ?)", search, search, search).
		Limit(limit).Find(&users).Error
	return users, err
}

func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// SetPlan — partial UPDATE users.plan_id, безопасный для webhook'а.
// gorm.Save(&User{ID, PlanID}) обнулял бы все остальные поля (email,
// password_hash, role) — см. CR-2 в REVIEW_2026-05-07.md.
func (r *userRepo) SetPlan(ctx context.Context, userID uint, planID string) error {
	res := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"plan_id":    planID,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *userRepo) SetQuotaWarningSentOn(ctx context.Context, userID uint, date time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"quota_warning_sent_on": date,
			"updated_at":            time.Now(),
		}).Error
}

func (r *userRepo) TouchLastLogin(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"last_login_at": now,
			"updated_at":    now,
		}).Error
}

func (r *userRepo) ListInactiveForReengagement(ctx context.Context, inactiveBefore, sentBefore time.Time, limit int) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).
		Where("email_verified = TRUE").
		Where("status = ?", models.StatusActive).
		Where("email <> ''").
		Where("last_login_at IS NOT NULL AND last_login_at < ?", inactiveBefore).
		Where("reengagement_sent_at IS NULL OR reengagement_sent_at < ?", sentBefore).
		Order("last_login_at ASC").
		Limit(limit).
		Find(&users).Error
	return users, err
}

// SetInsightEmailsEnabled атомарно тоглит флаг. Phase 14 M-10 (opt-in ФЗ-152).
func (r *userRepo) SetInsightEmailsEnabled(ctx context.Context, userID uint, enabled bool) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"insight_emails_enabled": enabled,
			"updated_at":             time.Now(),
		}).Error
}

func (r *userRepo) MarkReengagementSent(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"reengagement_sent_at": now,
			"updated_at":           now,
		}).Error
}

func (r *userRepo) CountReferredBy(ctx context.Context, code string) (int64, error) {
	if code == "" {
		return 0, nil
	}
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("referred_by = ?", code).
		Count(&count).Error
	return count, err
}

func (r *userRepo) GetByReferralCode(ctx context.Context, code string) (*models.User, error) {
	if code == "" {
		return nil, repo.ErrNotFound
	}
	var user models.User
	err := r.db.WithContext(ctx).
		Where("referral_code = ?", code).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// MarkReferralRewarded — атомарный "compare-and-set": UPDATE с WHERE
// referral_rewarded_at IS NULL. Возвращает true если RowsAffected=1.
// Это единственный безопасный путь — несколько webhook'ов могут прийти
// параллельно при дублировании уведомлений от T-Bank.
func (r *userRepo) MarkReferralRewarded(ctx context.Context, userID uint) (bool, error) {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND referral_rewarded_at IS NULL", userID).
		Updates(map[string]any{
			"referral_rewarded_at": now,
			"updated_at":           now,
		})
	return res.RowsAffected == 1, res.Error
}
