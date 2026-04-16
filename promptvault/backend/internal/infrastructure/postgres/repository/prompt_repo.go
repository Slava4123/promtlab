package repository

import (
	"context"
	"errors"
	"strings"

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

	// Коллекция (many-to-many). P-2: INNER JOIN вместо IN (SELECT ...) —
	// дубликатов нет благодаря unique constraint (prompt_id, collection_id),
	// поэтому DISTINCT не нужен. JOIN даёт планировщику больше свободы
	// (например, использовать индекс на collection_id напрямую).
	if filter.CollectionID != nil {
		q = q.Joins(
			"INNER JOIN prompt_collections pc ON pc.prompt_id = prompts.id AND pc.collection_id = ?",
			*filter.CollectionID,
		)
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

	// Теги. P-2: EXISTS — семантически правильный semi-join без дубликатов
	// (если у промпта несколько тегов из фильтра, INNER JOIN дал бы N копий
	// и потребовал бы DISTINCT, что медленнее). PostgreSQL реализует EXISTS
	// через semi-join node, потенциально используя индекс на prompt_tags.prompt_id.
	if len(filter.TagIDs) > 0 {
		q = q.Where(
			"EXISTS (SELECT 1 FROM prompt_tags WHERE prompt_id = prompts.id AND tag_id IN ?)",
			filter.TagIDs,
		)
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

func (r *promptRepo) SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error) {
	pattern := strings.ToLower(prefix) + "%"
	var titles []string
	q := r.db.WithContext(ctx).Model(&models.Prompt{}).Select("DISTINCT title")
	if teamID != nil {
		q = q.Where("team_id = ? AND lower(title) LIKE ?", *teamID, pattern)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL AND lower(title) LIKE ?", userID, pattern)
	}
	err := q.Order("title").Limit(limit).Pluck("title", &titles).Error
	return titles, err
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
		Updates(map[string]interface{}{
			"usage_count": gorm.Expr("usage_count + 1"),
			"last_used_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *promptRepo) UpdateLastUsed(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Prompt{}).
		Where("id = ?", id).
		Update("last_used_at", gorm.Expr("NOW()")).Error
}

func (r *promptRepo) ListRecent(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error) {
	q := r.db.WithContext(ctx).
		Where("last_used_at IS NOT NULL").
		Preload("Tags").Preload("Collections")

	if teamID != nil {
		q = q.Where("team_id = ?", *teamID)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL", userID)
	}

	if limit < 1 || limit > 100 {
		limit = 10
	}

	var prompts []models.Prompt
	err := q.Order("last_used_at DESC").Limit(limit).Find(&prompts).Error
	return prompts, err
}

func (r *promptRepo) LogUsage(ctx context.Context, userID, promptID uint) error {
	log := &models.PromptUsageLog{UserID: userID, PromptID: promptID}
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *promptRepo) ListUsageHistory(ctx context.Context, userID uint, teamID *uint, page, pageSize int) ([]models.PromptUsageLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.PromptUsageLog{}).
		Joins("JOIN prompts ON prompts.id = prompt_usage_log.prompt_id AND prompts.deleted_at IS NULL").
		Where("prompt_usage_log.user_id = ?", userID)

	if teamID != nil {
		q = q.Where("prompts.team_id = ?", *teamID)
	} else {
		q = q.Where("prompts.team_id IS NULL")
	}

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
	offset := (page - 1) * pageSize

	var logs []models.PromptUsageLog
	err := q.Preload("Prompt").Preload("Prompt.Tags").Preload("Prompt.Collections").
		Order("prompt_usage_log.used_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&logs).Error

	return logs, total, err
}

func (r *promptRepo) GetPublicBySlug(ctx context.Context, slug string) (*models.Prompt, error) {
	if slug == "" {
		return nil, repo.ErrNotFound
	}
	var p models.Prompt
	err := r.db.WithContext(ctx).
		Preload("Tags").Preload("Collections").
		Where("slug = ? AND is_public = TRUE", slug).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repo.ErrNotFound
	}
	return &p, err
}

func (r *promptRepo) ListPublic(ctx context.Context, limit int) ([]models.Prompt, error) {
	if limit <= 0 || limit > 10000 {
		limit = 1000
	}
	var out []models.Prompt
	err := r.db.WithContext(ctx).
		Select("id, slug, title, updated_at").
		Where("is_public = TRUE").
		Order("updated_at DESC").
		Limit(limit).
		Find(&out).Error
	return out, err
}
