package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type tagRepo struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *tagRepo {
	return &tagRepo{db: db}
}

func (r *tagRepo) GetOrCreate(ctx context.Context, name, color string, userID uint, teamID *uint) (*models.Tag, error) {
	// Atomic upsert: INSERT ... ON CONFLICT DO NOTHING, then SELECT.
	// The unique index idx_tags_name_user_team (from migration 000002)
	// guarantees no duplicates even under concurrent inserts.
	coalescedTeamID := uint(0)
	if teamID != nil {
		coalescedTeamID = *teamID
	}

	// Attempt insert; silently ignored if the unique constraint fires.
	r.db.WithContext(ctx).Exec(
		`INSERT INTO tags (name, color, user_id, team_id, created_at, updated_at)
		 VALUES (?, ?, ?, NULLIF(?, 0), NOW(), NOW())
		 ON CONFLICT (name, user_id, COALESCE(team_id, 0)) DO NOTHING`,
		name, color, userID, coalescedTeamID,
	)

	// Always SELECT to return the existing (or just-inserted) row.
	var tag models.Tag
	q := r.db.WithContext(ctx).Where("name = ?", name)
	if teamID != nil {
		q = q.Where("team_id = ?", *teamID)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL", userID)
	}
	if err := q.First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepo) List(ctx context.Context, userID uint, teamID *uint) ([]models.Tag, error) {
	var tags []models.Tag
	q := r.db.WithContext(ctx)
	if teamID != nil {
		q = q.Where("team_id = ?", *teamID)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL", userID)
	}
	err := q.Order("name").Find(&tags).Error
	return tags, err
}

func (r *tagRepo) GetByID(ctx context.Context, id uint) (*models.Tag, error) {
	var tag models.Tag
	if err := r.db.WithContext(ctx).First(&tag, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepo) GetByIDs(ctx context.Context, ids []uint) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&tags).Error
	return tags, err
}

func (r *tagRepo) SearchByQuery(ctx context.Context, userID uint, teamID *uint, query string, limit int) ([]models.Tag, error) {
	search := "%" + query + "%"
	var tags []models.Tag
	q := r.db.WithContext(ctx)
	if teamID != nil {
		q = q.Where("team_id = ? AND name ILIKE ?", *teamID, search)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL AND name ILIKE ?", userID, search)
	}
	err := q.Order("name").
		Limit(limit).
		Find(&tags).Error
	return tags, err
}

func (r *tagRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM prompt_tags WHERE tag_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Tag{}, id).Error
	})
}
