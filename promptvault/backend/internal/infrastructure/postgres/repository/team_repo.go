package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type teamRepo struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *teamRepo {
	return &teamRepo{db: db}
}

func (r *teamRepo) CreateWithOwner(ctx context.Context, team *models.Team, ownerUserID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		member := &models.TeamMember{
			TeamID: team.ID,
			UserID: ownerUserID,
			Role:   models.RoleOwner,
		}
		return tx.Create(member).Error
	})
}

func (r *teamRepo) GetBySlug(ctx context.Context, slug string) (*models.Team, error) {
	var team models.Team
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&team).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &team, nil
}

// GetByID — Phase 14. PK lookup. Нужен для share.GetPublicPrompt
// (prompts.team_id есть, slug нет).
func (r *teamRepo) GetByID(ctx context.Context, id uint) (*models.Team, error) {
	var team models.Team
	if err := r.db.WithContext(ctx).First(&team, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &team, nil
}

func (r *teamRepo) ListByUserID(ctx context.Context, userID uint) ([]models.Team, error) {
	var teams []models.Team
	err := r.db.WithContext(ctx).
		Joins("JOIN team_members ON team_members.team_id = teams.id").
		Where("team_members.user_id = ?", userID).
		Order("teams.name").
		Find(&teams).Error
	return teams, err
}

func (r *teamRepo) ListByUserIDWithRolesAndCounts(ctx context.Context, userID uint) ([]models.TeamWithRoleAndCount, error) {
	var results []models.TeamWithRoleAndCount
	err := r.db.WithContext(ctx).
		Table("teams").
		Select("teams.*, tm.role, COALESCE(mc.cnt, 0) AS member_count").
		Joins("JOIN team_members tm ON tm.team_id = teams.id AND tm.user_id = ?", userID).
		Joins("LEFT JOIN (SELECT team_id, COUNT(*) AS cnt FROM team_members GROUP BY team_id) mc ON mc.team_id = teams.id").
		Order("teams.name").
		Scan(&results).Error
	return results, err
}

func (r *teamRepo) Update(ctx context.Context, team *models.Team) error {
	return r.db.WithContext(ctx).Save(team).Error
}

// UpdateBranding — Phase 14. Точечный UPDATE brand_* полей без затрагивания
// name/description/created_by. Пустая строка сохраняется как "", что эквивалентно
// очистке (не пишем NULL — поле VARCHAR без NOT NULL).
func (r *teamRepo) UpdateBranding(ctx context.Context, teamID uint, logoURL, tagline, website, primaryColor string) error {
	return r.db.WithContext(ctx).Model(&models.Team{}).
		Where("id = ?", teamID).
		Updates(map[string]any{
			"brand_logo_url":      logoURL,
			"brand_tagline":       tagline,
			"brand_website":       website,
			"brand_primary_color": primaryColor,
		}).Error
}

func (r *teamRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Удалить приглашения
		if err := tx.Where("team_id = ?", id).Delete(&models.TeamInvitation{}).Error; err != nil {
			return err
		}
		// Удалить всех участников
		if err := tx.Where("team_id = ?", id).Delete(&models.TeamMember{}).Error; err != nil {
			return err
		}
		// Обнулить team_id в коллекциях, промптах, тегах
		if err := tx.Model(&models.Collection{}).Where("team_id = ?", id).Update("team_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Prompt{}).Where("team_id = ?", id).Update("team_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Tag{}).Where("team_id = ?", id).Update("team_id", nil).Error; err != nil {
			return err
		}
		// Удалить команду
		return tx.Delete(&models.Team{}, id).Error
	})
}

func (r *teamRepo) GetMember(ctx context.Context, teamID, userID uint) (*models.TeamMember, error) {
	var member models.TeamMember
	if err := r.db.WithContext(ctx).Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *teamRepo) UpdateMemberRole(ctx context.Context, teamID, userID uint, role models.TeamRole) error {
	return r.db.WithContext(ctx).
		Model(&models.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Update("role", role).Error
}

func (r *teamRepo) CountMembers(ctx context.Context, teamID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.TeamMember{}).Where("team_id = ?", teamID).Count(&count).Error
	return int(count), err
}

func (r *teamRepo) RemoveMember(ctx context.Context, teamID, userID uint) error {
	return r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&models.TeamMember{}).Error
}

func (r *teamRepo) ListMembers(ctx context.Context, teamID uint) ([]models.TeamMember, error) {
	var members []models.TeamMember
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Preload("User").
		Find(&members).Error
	return members, err
}

// Invitations

func (r *teamRepo) CreateInvitation(ctx context.Context, inv *models.TeamInvitation) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *teamRepo) GetInvitationByID(ctx context.Context, id uint) (*models.TeamInvitation, error) {
	var inv models.TeamInvitation
	if err := r.db.WithContext(ctx).Preload("Team").Preload("Inviter").First(&inv, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *teamRepo) GetPendingInvitation(ctx context.Context, teamID, userID uint) (*models.TeamInvitation, error) {
	var inv models.TeamInvitation
	if err := r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ? AND status = ?", teamID, userID, models.InvitationPending).
		First(&inv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *teamRepo) ListPendingByUserID(ctx context.Context, userID uint) ([]models.TeamInvitation, error) {
	var invitations []models.TeamInvitation
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, models.InvitationPending).
		Preload("Team").
		Preload("Inviter").
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

func (r *teamRepo) ListPendingByTeamID(ctx context.Context, teamID uint) ([]models.TeamInvitation, error) {
	var invitations []models.TeamInvitation
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND status = ?", teamID, models.InvitationPending).
		Preload("User").
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

func (r *teamRepo) UpdateInvitationStatus(ctx context.Context, id uint, status models.InvitationStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.TeamInvitation{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *teamRepo) DeleteInvitation(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.TeamInvitation{}, id).Error
}

func (r *teamRepo) AcceptInvitationTx(ctx context.Context, invID uint, member *models.TeamMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		return tx.Model(&models.TeamInvitation{}).Where("id = ?", invID).Update("status", models.InvitationAccepted).Error
	})
}
