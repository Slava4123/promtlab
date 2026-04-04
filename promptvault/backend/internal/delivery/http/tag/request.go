package tag

type CreateRequest struct {
	Name   string `json:"name" validate:"required,min=1,max=50"`
	Color  string `json:"color" validate:"max=7"`
	TeamID *uint  `json:"team_id"`
}
