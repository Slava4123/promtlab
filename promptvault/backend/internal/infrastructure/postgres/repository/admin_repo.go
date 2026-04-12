package repository

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type adminRepo struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) repo.AdminRepository {
	return &adminRepo{db: db}
}

func (r *adminRepo) ListUsers(ctx context.Context, filter repo.UserListFilter) ([]repo.UserSummary, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.User{})

	if filter.Query != "" {
		escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(filter.Query)
		like := "%" + escaped + "%"
		q = q.Where("email ILIKE ? OR username ILIKE ? OR name ILIKE ?", like, like, like)
	}
	if filter.Role != "" {
		q = q.Where("role = ?", filter.Role)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Сортировка: whitelist колонок, чтобы не пропустить SQL-инъекцию через SortBy.
	sortCol := "created_at"
	if filter.SortBy == "email" {
		sortCol = "email"
	}
	order := sortCol + " ASC"
	if filter.SortDesc || filter.SortBy == "" {
		// По умолчанию created_at DESC — свежие сверху.
		order = sortCol + " DESC"
	}

	page := max(filter.Page, 1)
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var users []models.User
	if err := q.Order(order).Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	summaries := make([]repo.UserSummary, 0, len(users))
	for _, u := range users {
		summaries = append(summaries, repo.UserSummary{
			ID:            u.ID,
			Email:         u.Email,
			Name:          u.Name,
			Username:      u.Username,
			Role:          string(u.Role),
			Status:        string(u.Status),
			EmailVerified: u.EmailVerified,
			CreatedAt:     u.CreatedAt,
		})
	}
	return summaries, total, nil
}

func (r *adminRepo) GetUserDetail(ctx context.Context, userID uint) (*repo.UserDetail, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Preload("LinkedAccounts").
		First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}

	detail := &repo.UserDetail{
		User: &user,
	}

	// Aggregations: отдельные COUNT queries — дёшево и читаемо.
	// Если станет узким местом (вряд ли для admin-панели), объединить в
	// один CTE-запрос с UNION'ом.
	if err := r.db.WithContext(ctx).
		Model(&models.Prompt{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&detail.PromptCount).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Collection{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&detail.CollectionCount).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&models.UserBadge{}).
		Where("user_id = ?", userID).
		Count(&detail.BadgeCount).Error; err != nil {
		return nil, err
	}

	var totalUsage *int64
	if err := r.db.WithContext(ctx).
		Model(&models.Prompt{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Select("COALESCE(SUM(usage_count), 0)").
		Scan(&totalUsage).Error; err != nil {
		return nil, err
	}
	if totalUsage != nil {
		detail.TotalUsage = *totalUsage
	}

	// LinkedProviders — извлекаем из preload'енного slice для удобства frontend.
	providers := make([]string, 0, len(user.LinkedAccounts))
	for _, la := range user.LinkedAccounts {
		providers = append(providers, la.Provider)
	}
	detail.LinkedProviders = providers

	// UnlockedBadgeIDs — список badge_id'шек для admin UI (отличать unlocked vs locked).
	// Всегда non-nil slice чтобы json marshal выдавал [] а не null.
	unlockedIDs := make([]string, 0)
	if err := r.db.WithContext(ctx).
		Model(&models.UserBadge{}).
		Where("user_id = ?", userID).
		Pluck("badge_id", &unlockedIDs).Error; err != nil {
		return nil, err
	}
	detail.UnlockedBadgeIDs = unlockedIDs

	return detail, nil
}

func (r *adminRepo) UpdateStatus(ctx context.Context, userID uint, status models.UserStatus) error {
	res := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *adminRepo) CountUsers(ctx context.Context) (total, admins, active, frozen int64, err error) {
	// Одним запросом через FILTER — дешевле, чем 4 отдельных COUNT.
	// COALESCE не нужен: COUNT возвращает 0 для пустого набора.
	const sql = `
		SELECT
			COUNT(*)                                    AS total,
			COUNT(*) FILTER (WHERE role = 'admin')     AS admins,
			COUNT(*) FILTER (WHERE status = 'active')  AS active,
			COUNT(*) FILTER (WHERE status = 'frozen')  AS frozen
		FROM users
	`
	row := r.db.WithContext(ctx).Raw(sql).Row()
	if err = row.Scan(&total, &admins, &active, &frozen); err != nil {
		return 0, 0, 0, 0, err
	}
	return total, admins, active, frozen, nil
}
