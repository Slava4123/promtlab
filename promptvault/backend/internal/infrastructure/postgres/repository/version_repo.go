package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type versionRepo struct {
	db *gorm.DB
}

func NewVersionRepository(db *gorm.DB) *versionRepo {
	return &versionRepo{db: db}
}

// CreateWithNextVersion атомарно вычисляет MAX(version_number)+1 и создаёт версию в одной транзакции.
func (r *versionRepo) CreateWithNextVersion(ctx context.Context, v *models.PromptVersion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Блокируем строки промпта для предотвращения race condition
		var locked []models.PromptVersion
		if err := tx.Raw("SELECT id FROM prompt_versions WHERE prompt_id = ? FOR UPDATE", v.PromptID).Scan(&locked).Error; err != nil {
			return err
		}

		var next uint
		if err := tx.Raw(
			"SELECT COALESCE(MAX(version_number), 0) + 1 FROM prompt_versions WHERE prompt_id = ?",
			v.PromptID,
		).Scan(&next).Error; err != nil {
			return err
		}
		v.VersionNumber = next
		return tx.Create(v).Error
	})
}

func (r *versionRepo) ListByPromptID(ctx context.Context, promptID uint, page, pageSize int) ([]models.PromptVersion, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.PromptVersion{}).Where("prompt_id = ?", promptID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var versions []models.PromptVersion
	err := q.Order("version_number DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&versions).Error
	return versions, total, err
}

// GetByIDForPrompt возвращает версию только если она принадлежит указанному промпту.
func (r *versionRepo) GetByIDForPrompt(ctx context.Context, versionID, promptID uint) (*models.PromptVersion, error) {
	var v models.PromptVersion
	if err := r.db.WithContext(ctx).Where("id = ? AND prompt_id = ?", versionID, promptID).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}
