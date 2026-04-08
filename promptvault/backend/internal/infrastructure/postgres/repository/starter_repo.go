package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type starterRepo struct {
	db *gorm.DB
}

func NewStarterRepository(db *gorm.DB) *starterRepo {
	return &starterRepo{db: db}
}

// InstallTemplates оборачивает batch insert промптов и UPDATE users в одну
// транзакцию. Любая ошибка → rollback всех изменений. Паттерн идентичен
// team_repo.CreateWithOwner и team_repo.AcceptInvitationTx.
//
// Конкурентная защита: UPDATE использует условие
// `WHERE onboarding_completed_at IS NULL`. Если строка не затронута
// (RowsAffected == 0), значит другая транзакция уже пометила юзера —
// возвращаем repo.ErrConflict, чтобы транзакция откатилась и созданные
// промпты не утекли. Это закрывает TOCTOU между service-уровневой
// проверкой и фактическим UPDATE.
func (r *starterRepo) InstallTemplates(ctx context.Context, userID uint, prompts []*models.Prompt) (time.Time, error) {
	var completedAt time.Time
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		completedAt = time.Now().UTC()
		if len(prompts) > 0 {
			// GORM v2 batch insert: один INSERT с несколькими VALUES.
			if err := tx.Create(&prompts).Error; err != nil {
				return err
			}
		}
		res := tx.Model(&models.User{}).
			Where("id = ? AND onboarding_completed_at IS NULL", userID).
			UpdateColumn("onboarding_completed_at", completedAt)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// Юзер либо уже завершил онбординг (concurrent install),
			// либо был удалён между service-check и tx. В обоих случаях
			// откатываем INSERT promptов.
			return repo.ErrConflict
		}
		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	return completedAt, nil
}
