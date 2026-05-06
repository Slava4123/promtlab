package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type teamLogoRepo struct {
	db *gorm.DB
}

func NewTeamLogoRepository(db *gorm.DB) *teamLogoRepo {
	return &teamLogoRepo{db: db}
}

func (r *teamLogoRepo) Get(ctx context.Context, teamID uint) (*models.TeamLogoFile, error) {
	var f models.TeamLogoFile
	if err := r.db.WithContext(ctx).Where("team_id = ?", teamID).First(&f).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}

// Upsert — INSERT ... ON CONFLICT (team_id) DO UPDATE.
// Один логотип на команду; новая загрузка атомарно перезатирает старую.
func (r *teamLogoRepo) Upsert(ctx context.Context, file *models.TeamLogoFile) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"content_type", "size_bytes", "sha256", "bytes", "uploaded_at"}),
	}).Create(file).Error
}

// Delete — idempotent: отсутствие записи не считается ошибкой.
func (r *teamLogoRepo) Delete(ctx context.Context, teamID uint) error {
	return r.db.WithContext(ctx).Where("team_id = ?", teamID).Delete(&models.TeamLogoFile{}).Error
}
