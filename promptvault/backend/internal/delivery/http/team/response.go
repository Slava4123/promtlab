package team

import (
	"time"

	"promptvault/internal/models"
)

type TeamResponse struct {
	ID          uint      `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Role        string    `json:"role"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TeamDetailResponse struct {
	TeamResponse
	Members []MemberResponse `json:"members"`
}

type MemberResponse struct {
	UserID    uint   `json:"user_id"`
	Name      string `json:"name"`
	Username  string `json:"username,omitempty"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Role      string `json:"role"`
}

func NewTeamResponse(t *models.Team, role models.TeamRole, memberCount int) TeamResponse {
	return TeamResponse{
		ID:          t.ID,
		Slug:        t.Slug,
		Name:        t.Name,
		Description: t.Description,
		Role:        string(role),
		MemberCount: memberCount,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func NewTeamDetailResponse(t *models.Team, role models.TeamRole, members []models.TeamMember) TeamDetailResponse {
	memberResponses := make([]MemberResponse, len(members))
	for i, m := range members {
		memberResponses[i] = NewMemberResponse(m)
	}

	return TeamDetailResponse{
		TeamResponse: NewTeamResponse(t, role, len(members)),
		Members:      memberResponses,
	}
}

func NewMemberResponse(m models.TeamMember) MemberResponse {
	return MemberResponse{
		UserID:    m.UserID,
		Name:      m.User.Name,
		Username:  m.User.Username,
		Email:     m.User.Email,
		AvatarURL: m.User.AvatarURL,
		Role:      string(m.Role),
	}
}

type InvitationResponse struct {
	ID        uint      `json:"id"`
	TeamID    uint      `json:"team_id"`
	TeamName  string    `json:"team_name"`
	TeamSlug  string    `json:"team_slug"`
	Role      string    `json:"role"`
	InviterName string  `json:"inviter_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func NewInvitationResponse(inv models.TeamInvitation) InvitationResponse {
	return InvitationResponse{
		ID:          inv.ID,
		TeamID:      inv.TeamID,
		TeamName:    inv.Team.Name,
		TeamSlug:    inv.Team.Slug,
		Role:        string(inv.Role),
		InviterName: inv.Inviter.Name,
		Status:      string(inv.Status),
		CreatedAt:   inv.CreatedAt,
	}
}

type PendingInvitationResponse struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username,omitempty"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

func NewPendingInvitationResponse(inv models.TeamInvitation) PendingInvitationResponse {
	return PendingInvitationResponse{
		ID:       inv.ID,
		Email:    inv.User.Email,
		Name:     inv.User.Name,
		Username: inv.User.Username,
		Role:     string(inv.Role),
		Status:   string(inv.Status),
	}
}
