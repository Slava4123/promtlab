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
