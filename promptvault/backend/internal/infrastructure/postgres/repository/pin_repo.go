package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type pinRepo struct {
	db *gorm.DB
}

func NewPinRepository(db *gorm.DB) *pinRepo {
	return &pinRepo{db: db}
}

func (r *pinRepo) Create(ctx context.Context, pin *models.PromptPin) error {
	return r.db.WithContext(ctx).Create(pin).Error
}

func (r *pinRepo) Delete(ctx context.Context, promptID, userID uint, teamWide bool) error {
	q := r.db.WithContext(ctx).Where("prompt_id = ? AND is_team_wide = ?", promptID, teamWide)
	if !teamWide {
		q = q.Where("user_id = ?", userID)
	}
	return q.Delete(&models.PromptPin{}).Error
}

func (r *pinRepo) Get(ctx context.Context, promptID, userID uint, teamWide bool) (*models.PromptPin, error) {
	var pin models.PromptPin
	q := r.db.WithContext(ctx).Where("prompt_id = ? AND is_team_wide = ?", promptID, teamWide)
	if !teamWide {
		q = q.Where("user_id = ?", userID)
	}
	if err := q.First(&pin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &pin, nil
}

func (r *pinRepo) GetStatuses(ctx context.Context, promptIDs []uint, userID uint) (map[uint]repo.PinStatus, error) {
	if len(promptIDs) == 0 {
		return make(map[uint]repo.PinStatus), nil
	}

	var pins []models.PromptPin
	err := r.db.WithContext(ctx).
		Where("prompt_id IN ? AND (user_id = ? OR is_team_wide = TRUE)", promptIDs, userID).
		Find(&pins).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]repo.PinStatus, len(promptIDs))
	for _, pin := range pins {
		status := result[pin.PromptID]
		if pin.IsTeamWide {
			status.PinnedTeam = true
		} else if pin.UserID == userID {
			status.PinnedPersonal = true
		}
		if status.PinnedAt == nil || pin.PinnedAt.After(*status.PinnedAt) {
			t := pin.PinnedAt
			status.PinnedAt = &t
		}
		result[pin.PromptID] = status
	}
	return result, nil
}

func (r *pinRepo) ListPinned(ctx context.Context, userID uint, teamID *uint, limit int) ([]models.Prompt, error) {
	q := r.db.WithContext(ctx).
		Model(&models.Prompt{}).
		Joins("JOIN prompt_pins pp ON pp.prompt_id = prompts.id").
		Preload("Tags").Preload("Collections")

	if teamID != nil {
		q = q.Where("prompts.team_id = ? AND (pp.user_id = ? OR pp.is_team_wide = TRUE)", *teamID, userID)
	} else {
		q = q.Where("prompts.user_id = ? AND prompts.team_id IS NULL AND pp.user_id = ?", userID, userID)
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	var prompts []models.Prompt
	err := q.Order("pp.pinned_at DESC").Limit(limit).Find(&prompts).Error
	return prompts, err
}
