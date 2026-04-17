package prompt

type CreateInput struct {
	UserID        uint
	TeamID        *uint
	Title         string
	Content       string
	Model         string
	CollectionIDs []uint
	TagIDs        []uint
	// IsPublic — пометить промпт публичным сразу при создании. Slug генерируется
	// после INSERT (нужен ID), потом второй UPDATE проставляет slug.
	IsPublic bool
}

type UpdateInput struct {
	Title         *string
	Content       *string
	Model         *string
	ChangeNote    string
	CollectionIDs []uint
	TagIDs        []uint
	// IsPublic — если не nil, переключает публичность промпта. При true
	// автоматически генерируется slug (если пуст). При false slug сохраняется
	// (чтобы URL не пропадал при повторной публикации).
	IsPublic *bool
}

type PinInput struct {
	PromptID uint
	UserID   uint
	TeamWide bool
}

type PinResult struct {
	Pinned   bool `json:"pinned"`
	TeamWide bool `json:"team_wide"`
}
