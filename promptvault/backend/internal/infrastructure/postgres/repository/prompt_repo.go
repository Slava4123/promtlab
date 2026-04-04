package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type promptRepo struct {
	db *gorm.DB
}

func NewPromptRepository(db *gorm.DB) *promptRepo {
	return &promptRepo{db: db}
}

func (r *promptRepo) Create(ctx context.Context, prompt *models.Prompt) error {
	return r.db.WithContext(ctx).Create(prompt).Error
}

func (r *promptRepo) GetByID(ctx context.Context, id uint) (*models.Prompt, error) {
	var prompt models.Prompt
	if err := r.db.WithContext(ctx).Preload("Tags").Preload("Collections").First(&prompt, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &prompt, nil
}

func (r *promptRepo) Update(ctx context.Context, prompt *models.Prompt) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(prompt).Association("Tags").Replace(prompt.Tags); err != nil {
			return err
		}
		if err := tx.Model(prompt).Association("Collections").Replace(prompt.Collections); err != nil {
			return err
		}
		return tx.Save(prompt).Error
	})
}

func (r *promptRepo) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Prompt{}).Error
}

func (r *promptRepo) List(ctx context.Context, filter repo.PromptListFilter) ([]models.Prompt, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Prompt{})

	// Контекст: команда или личные
	if len(filter.TeamIDs) > 0 {
		q = q.Where("team_id IN ?", filter.TeamIDs)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL", filter.UserID)
	}

	// Коллекция (many-to-many)
	if filter.CollectionID != nil {
		q = q.Where("id IN (SELECT prompt_id FROM prompt_collections WHERE collection_id = ?)", *filter.CollectionID)
	}

	// Избранное
	if filter.FavoriteOnly {
		q = q.Where("favorite = ?", true)
	}

	// Поиск
	if filter.Query != "" {
		search := "%" + filter.Query + "%"
		q = q.Where("title ILIKE ? OR content ILIKE ?", search, search)
	}

	// Теги
	if len(filter.TagIDs) > 0 {
		q = q.Where("id IN (SELECT prompt_id FROM prompt_tags WHERE tag_id IN ?)", filter.TagIDs)
	}

	// Считаем общее количество
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Пагинация
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var prompts []models.Prompt
	err := q.Preload("Tags").Preload("Collections").
		Order("updated_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&prompts).Error

	return prompts, total, err
}

func (r *promptRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Prompt, error) {
	search := "%" + query + "%"
	var prompts []models.Prompt
	q := r.db.WithContext(ctx)
	if teamID != nil {
		q = q.Where("team_id = ? AND (title ILIKE ? OR content ILIKE ?)", *teamID, search, search)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL AND (title ILIKE ? OR content ILIKE ?)", userID, search, search)
	}
	err := q.Preload("Tags").Preload("Collections").
		Order("updated_at DESC").
		Limit(limit).
		Find(&prompts).Error
	return prompts, err
}

func (r *promptRepo) SetFavorite(ctx context.Context, id uint, favorite bool) error {
	return r.db.WithContext(ctx).
		Model(&models.Prompt{}).
		Where("id = ?", id).
		Update("favorite", favorite).Error
}

func (r *promptRepo) IncrementUsage(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Prompt{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}
