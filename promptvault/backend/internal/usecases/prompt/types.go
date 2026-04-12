package prompt

type CreateInput struct {
	UserID        uint
	TeamID        *uint
	Title         string
	Content       string
	Model         string
	CollectionIDs []uint
	TagIDs        []uint
}

type UpdateInput struct {
	Title         *string
	Content       *string
	Model         *string
	ChangeNote    string
	CollectionIDs []uint
	TagIDs        []uint
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
