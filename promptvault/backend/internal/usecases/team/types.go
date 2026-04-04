package team

import "promptvault/internal/models"

type CreateInput struct {
	Name        string
	Description string
}

type UpdateInput struct {
	Name        *string
	Description *string
}

type AddMemberInput struct {
	Query string
	Role  models.TeamRole
}

type TeamListItem struct {
	Team        models.Team
	Role        models.TeamRole
	MemberCount int
}
