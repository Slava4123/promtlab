package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type trashRepo struct {
	db *gorm.DB
}

func NewTrashRepository(db *gorm.DB) *trashRepo {
	return &trashRepo{db: db}
}

// ---------- List ----------

func (r *trashRepo) ListDeletedPrompts(ctx context.Context, userID uint, teamIDs []uint, page, pageSize int) ([]models.Prompt, int64, error) {
	q := r.db.WithContext(ctx).Unscoped().Model(&models.Prompt{}).Where("deleted_at IS NOT NULL")

	if len(teamIDs) > 0 {
		q = q.Where("team_id IN ?", teamIDs)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL", userID)
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

	var prompts []models.Prompt
	err := q.Preload("Tags").Preload("Collections").
		Order("deleted_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&prompts).Error
	return prompts, total, err
}

func (r *trashRepo) ListDeletedCollections(ctx context.Context, userID uint, teamIDs []uint) ([]models.Collection, error) {
	q := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL")

	if len(teamIDs) > 0 {
		q = q.Where("team_id IN ?", teamIDs)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL", userID)
	}

	var collections []models.Collection
	err := q.Order("deleted_at DESC").Find(&collections).Error
	return collections, err
}

// ---------- Count ----------

func (r *trashRepo) CountDeleted(ctx context.Context, userID uint, teamIDs []uint) (repo.TrashCounts, error) {
	var counts repo.TrashCounts

	countQuery := func(model any) *gorm.DB {
		q := r.db.WithContext(ctx).Unscoped().Model(model).Where("deleted_at IS NOT NULL")
		if len(teamIDs) > 0 {
			q = q.Where("team_id IN ?", teamIDs)
		} else {
			q = q.Where("user_id = ? AND team_id IS NULL", userID)
		}
		return q
	}

	if err := countQuery(&models.Prompt{}).Count(&counts.Prompts).Error; err != nil {
		return counts, err
	}
	if err := countQuery(&models.Collection{}).Count(&counts.Collections).Error; err != nil {
		return counts, err
	}

	return counts, nil
}

// ---------- Get deleted by ID ----------

func (r *trashRepo) GetDeletedPrompt(ctx context.Context, id uint) (*models.Prompt, error) {
	var p models.Prompt
	if err := r.db.WithContext(ctx).Unscoped().
		Preload("Tags").Preload("Collections").
		Where("id = ? AND deleted_at IS NOT NULL", id).
		First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *trashRepo) GetDeletedCollection(ctx context.Context, id uint) (*models.Collection, error) {
	var c models.Collection
	if err := r.db.WithContext(ctx).Unscoped().
		Where("id = ? AND deleted_at IS NOT NULL", id).
		First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

// ---------- Restore ----------

func (r *trashRepo) RestorePrompt(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Unscoped().
		Model(&models.Prompt{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *trashRepo) RestoreCollection(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Unscoped().
		Model(&models.Collection{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

// ---------- Hard delete ----------

func (r *trashRepo) HardDeletePrompt(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Unscoped().
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Delete(&models.Prompt{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *trashRepo) HardDeleteCollection(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Unscoped().
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Delete(&models.Collection{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

// ---------- Bulk operations ----------

// promptsInChainsSubquery — Phase 16. Исключение для prompts, используемых в
// активных цепочках. FK prompt_chain_steps.prompt_id без CASCADE дал бы 23503
// при purge → вместо этого silent-skip. При CHAINS_ENABLED=false таблица
// prompt_chain_steps пуста → подзапрос no-op.
const promptsInChainsSubquery = `id NOT IN (SELECT DISTINCT prompt_id FROM prompt_chain_steps)`

func (r *trashRepo) PurgeExpired(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	var total int64

	// Prompts: skip те, что используются в цепочках (FK guard).
	res := r.db.WithContext(ctx).Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).
		Where(promptsInChainsSubquery).
		Delete(&models.Prompt{})
	if res.Error != nil {
		return total, res.Error
	}
	total += res.RowsAffected

	// Collections — chain не зависит от collection, обычный delete.
	res = r.db.WithContext(ctx).Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).
		Delete(&models.Collection{})
	if res.Error != nil {
		return total, res.Error
	}
	total += res.RowsAffected

	return total, nil
}

func (r *trashRepo) EmptyTrash(ctx context.Context, userID uint, teamIDs []uint) (int64, error) {
	var total int64

	// Prompts с защитой от FK 23503.
	q := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL")
	if len(teamIDs) > 0 {
		q = q.Where("team_id IN ?", teamIDs)
	} else {
		q = q.Where("user_id = ? AND team_id IS NULL", userID)
	}
	res := q.Where(promptsInChainsSubquery).Delete(&models.Prompt{})
	if res.Error != nil {
		return total, res.Error
	}
	total += res.RowsAffected

	q2 := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL")
	if len(teamIDs) > 0 {
		q2 = q2.Where("team_id IN ?", teamIDs)
	} else {
		q2 = q2.Where("user_id = ? AND team_id IS NULL", userID)
	}
	res = q2.Delete(&models.Collection{})
	if res.Error != nil {
		return total, res.Error
	}
	total += res.RowsAffected

	return total, nil
}
