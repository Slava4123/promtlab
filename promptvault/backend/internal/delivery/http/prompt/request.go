package prompt

type CreatePromptRequest struct {
	Title         string `json:"title" validate:"required,min=1,max=300"`
	Content       string `json:"content" validate:"required,max=15000"`
	Model         string `json:"model" validate:"max=100"`
	TeamID        *uint  `json:"team_id"`
	CollectionIDs []uint `json:"collection_ids"`
	TagIDs        []uint `json:"tag_ids"`
}

type PinRequest struct {
	TeamWide bool `json:"team_wide"`
}

type UpdatePromptRequest struct {
	Title         *string `json:"title" validate:"omitempty,min=1,max=300"`
	Content       *string `json:"content" validate:"omitempty,max=15000"`
	Model         *string `json:"model" validate:"omitempty,max=100"`
	ChangeNote    string  `json:"change_note" validate:"max=300"`
	CollectionIDs []uint  `json:"collection_ids"`
	TagIDs        []uint  `json:"tag_ids"`
	IsPublic      *bool   `json:"is_public"`
}
