package repository

import (
	"context"

	"promptvault/internal/models"
)

// TeamLogoRepository — bytea-хранилище загруженных логотипов команд.
// Один файл на команду; Upsert — INSERT ... ON CONFLICT (team_id) DO UPDATE.
type TeamLogoRepository interface {
	Get(ctx context.Context, teamID uint) (*models.TeamLogoFile, error)
	Upsert(ctx context.Context, file *models.TeamLogoFile) error
	Delete(ctx context.Context, teamID uint) error
}
