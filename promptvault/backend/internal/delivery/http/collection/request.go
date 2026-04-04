package collection

type CreateRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color" validate:"max=20"`
	Icon        string `json:"icon" validate:"max=10"`
	TeamID      *uint  `json:"team_id"`
}

type UpdateRequest struct {
	Name        string `json:"name" validate:"max=200"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color" validate:"max=20"`
	Icon        string `json:"icon" validate:"max=10"`
}
