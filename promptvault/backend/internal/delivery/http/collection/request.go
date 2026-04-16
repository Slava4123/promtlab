package collection

// Color — hex-цвет (#RGB/#RGBA/#RRGGBB/#RRGGBBAA). Строгая валидация защищает
// от CSS-injection во frontend inline-style: `tag.color + "15"` → `backgroundColor`.
type CreateRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color" validate:"omitempty,hexcolor"`
	Icon        string `json:"icon" validate:"max=10"`
	TeamID      *uint  `json:"team_id"`
}

type UpdateRequest struct {
	Name        string `json:"name" validate:"max=200"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color" validate:"omitempty,hexcolor"`
	Icon        string `json:"icon" validate:"max=10"`
}
