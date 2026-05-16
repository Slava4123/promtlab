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

// UpdateRequest — partial update. Все поля nullable: nil = «не трогать», "" = «очистить».
// До этой правки `Description string` интерпретировался как `""` при отсутствии
// поля → silent description loss при PUT с только {name}. Теперь usecase
// обновляет только non-nil поля.
type UpdateRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	Color       *string `json:"color" validate:"omitempty,hexcolor"`
	Icon        *string `json:"icon" validate:"omitempty,max=10"`
}
