package team

type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateTeamRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// AddMemberRequest принимает email/username через поле `query`.
// `Email` оставлен для backward-compat со старыми клиентами расширения,
// которые слали `{email, role}` (исправлено в lib/api.ts:579, но
// inflight extension'ы могут продолжать слать email-поле). Handler берёт
// сначала `Query`, при пустом — `Email`. Один из них должен быть непустым.
type AddMemberRequest struct {
	Query string `json:"query"`
	Email string `json:"email"`
	Role  string `json:"role" validate:"required,oneof=editor viewer"`
}

// LookupValue — единый getter для имени/email независимо от поля.
func (r AddMemberRequest) LookupValue() string {
	if r.Query != "" {
		return r.Query
	}
	return r.Email
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=editor viewer"`
}
