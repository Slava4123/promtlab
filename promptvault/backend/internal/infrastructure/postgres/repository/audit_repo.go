package repository

import (
	"context"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type auditRepo struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) repo.AuditRepository {
	return &auditRepo{db: db}
}

func (r *auditRepo) Log(ctx context.Context, entry *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *auditRepo) List(ctx context.Context, filter repo.AuditLogFilter) ([]models.AuditLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.AuditLog{})

	if filter.AdminID != nil {
		q = q.Where("admin_id = ?", *filter.AdminID)
	}
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.TargetType != "" {
		q = q.Where("target_type = ?", filter.TargetType)
	}
	if filter.TargetID != nil {
		q = q.Where("target_id = ?", *filter.TargetID)
	}
	if filter.FromTime != nil {
		q = q.Where("created_at >= ?", *filter.FromTime)
	}
	if filter.ToTime != nil {
		q = q.Where("created_at <= ?", *filter.ToTime)
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

	var entries []models.AuditLog
	err := q.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&entries).Error
	return entries, total, err
}
