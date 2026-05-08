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

type feedbackRepo struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) repo.FeedbackRepository {
	return &feedbackRepo{db: db}
}

func (r *feedbackRepo) Create(ctx context.Context, feedback *models.Feedback) error {
	return r.db.WithContext(ctx).Create(feedback).Error
}

// listRow — сырая строка из SELECT с join'ом, маппится в FeedbackListItem.
// Дублирует структуру намеренно: time.Time из БД → RFC3339-string в DTO,
// чтобы JSON был однозначен (без зависимости от сериализатора).
type listRow struct {
	ID        uint                  `gorm:"column:id"`
	UserID    uint                  `gorm:"column:user_id"`
	UserEmail string                `gorm:"column:user_email"`
	UserName  string                `gorm:"column:user_name"`
	Type      models.FeedbackType   `gorm:"column:type"`
	Status    models.FeedbackStatus `gorm:"column:status"`
	Message   string                `gorm:"column:message"`
	PageURL   string                `gorm:"column:page_url"`
	CreatedAt time.Time             `gorm:"column:created_at"`
}

func (r listRow) toItem() repo.FeedbackListItem {
	return repo.FeedbackListItem{
		ID:        r.ID,
		UserID:    r.UserID,
		UserEmail: r.UserEmail,
		UserName:  r.UserName,
		Type:      r.Type,
		Status:    r.Status,
		Message:   r.Message,
		PageURL:   r.PageURL,
		CreatedAt: r.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (r *feedbackRepo) List(ctx context.Context, filter repo.FeedbackListFilter) ([]repo.FeedbackListItem, int64, error) {
	q := r.db.WithContext(ctx).
		Table("feedbacks AS f").
		Joins("LEFT JOIN users AS u ON u.id = f.user_id")

	if filter.Type != "" {
		q = q.Where("f.type = ?", filter.Type)
	}
	if filter.Status != "" {
		q = q.Where("f.status = ?", filter.Status)
	}
	if filter.Query != "" {
		// Escape ILIKE wildcards в user-input, чтобы поиск был literal.
		escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(filter.Query)
		like := "%" + escaped + "%"
		q = q.Where("f.message ILIKE ? OR u.email ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := max(filter.Page, 1)
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var rows []listRow
	err := q.Select(`
		f.id          AS id,
		f.user_id     AS user_id,
		u.email       AS user_email,
		u.name        AS user_name,
		f.type        AS type,
		f.status      AS status,
		f.message     AS message,
		f.page_url    AS page_url,
		f.created_at  AS created_at
	`).Order("f.created_at DESC").Limit(pageSize).Offset(offset).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]repo.FeedbackListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toItem())
	}
	return items, total, nil
}

func (r *feedbackRepo) GetByID(ctx context.Context, id uint) (*repo.FeedbackDetail, error) {
	var row listRow
	err := r.db.WithContext(ctx).
		Table("feedbacks AS f").
		Joins("LEFT JOIN users AS u ON u.id = f.user_id").
		Select(`
			f.id          AS id,
			f.user_id     AS user_id,
			u.email       AS user_email,
			u.name        AS user_name,
			f.type        AS type,
			f.status      AS status,
			f.message     AS message,
			f.page_url    AS page_url,
			f.created_at  AS created_at
		`).
		Where("f.id = ?", id).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	item := row.toItem()
	return &item, nil
}

func (r *feedbackRepo) UpdateStatus(ctx context.Context, id uint, status models.FeedbackStatus) error {
	res := r.db.WithContext(ctx).
		Model(&models.Feedback{}).
		Where("id = ?", id).
		Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *feedbackRepo) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&models.Feedback{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}
