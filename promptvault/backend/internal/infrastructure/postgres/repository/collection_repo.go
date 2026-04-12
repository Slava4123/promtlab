package repository

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type collectionRepo struct {
	db *gorm.DB
}

func NewCollectionRepository(db *gorm.DB) *collectionRepo {
	return &collectionRepo{db: db}
}

func (r *collectionRepo) Create(ctx context.Context, c *models.Collection) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *collectionRepo) GetByID(ctx context.Context, id uint) (*models.Collection, error) {
	var c models.Collection
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *collectionRepo) Update(ctx context.Context, c *models.Collection) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *collectionRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Collection{}).Error
}

func (r *collectionRepo) CountPrompts(ctx context.Context, collectionID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw("SELECT count(*) FROM prompt_collections WHERE collection_id = ?", collectionID).Scan(&count).Error
	return count, err
}

func (r *collectionRepo) ListWithCounts(ctx context.Context, userID uint, teamIDs []uint) ([]models.CollectionWithCount, error) {
	var rows []models.CollectionWithCount
	q := r.db.WithContext(ctx).
		Table("collections").
		Select("collections.*, COALESCE(pc.cnt, 0) as prompt_count").
		Joins("LEFT JOIN (SELECT pc.collection_id, COUNT(*) as cnt FROM prompt_collections pc JOIN prompts p ON p.id = pc.prompt_id AND p.deleted_at IS NULL GROUP BY pc.collection_id) pc ON pc.collection_id = collections.id").
		Where("collections.deleted_at IS NULL")

	if len(teamIDs) > 0 {
		q = q.Where("collections.team_id IN ?", teamIDs)
	} else {
		q = q.Where("collections.user_id = ? AND collections.team_id IS NULL", userID)
	}

	err := q.Order("collections.name").Find(&rows).Error
	return rows, err
}

func (r *collectionRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Collection, error) {
	search := "%" + query + "%"
	var collections []models.Collection
	q := r.db.WithContext(ctx)
	if teamID != nil {
		q = q.Where("team_id = ? AND (name ILIKE ? OR description ILIKE ?)", *teamID, search, search)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL AND (name ILIKE ? OR description ILIKE ?)", userID, search, search)
	}
	err := q.Order("name").
		Limit(limit).
		Find(&collections).Error
	return collections, err
}

func (r *collectionRepo) SuggestByPrefix(ctx context.Context, userID uint, teamID *uint, prefix string, limit int) ([]string, error) {
	pattern := strings.ToLower(prefix) + "%"
	var names []string
	q := r.db.WithContext(ctx).Model(&models.Collection{}).Select("DISTINCT name")
	if teamID != nil {
		q = q.Where("team_id = ? AND lower(name) LIKE ?", *teamID, pattern)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL AND lower(name) LIKE ?", userID, pattern)
	}
	err := q.Order("name").Limit(limit).Pluck("name", &names).Error
	return names, err
}

func (r *collectionRepo) GetByIDs(ctx context.Context, ids []uint) ([]models.Collection, error) {
	var collections []models.Collection
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&collections).Error
	return collections, err
}
