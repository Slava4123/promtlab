package apikey

type CreateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}
