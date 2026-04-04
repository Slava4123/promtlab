package team

type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateTeamRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

type AddMemberRequest struct {
	Query string `json:"query" validate:"required"`
	Role  string `json:"role" validate:"required,oneof=editor viewer"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=editor viewer"`
}
