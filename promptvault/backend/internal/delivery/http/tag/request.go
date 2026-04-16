package tag

// Color — hex-цвет в формате #RGB/#RGBA/#RRGGBB/#RRGGBBAA (validator hexcolor tag).
// Строгая валидация защищает от CSS-injection во frontend inline-style.
type CreateRequest struct {
	Name   string `json:"name" validate:"required,min=1,max=50"`
	Color  string `json:"color" validate:"omitempty,hexcolor"`
	TeamID *uint  `json:"team_id"`
}
