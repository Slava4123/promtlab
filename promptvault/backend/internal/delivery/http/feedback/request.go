package feedback

// SubmitRequest — тело POST /api/feedback.
type SubmitRequest struct {
	Type    string `json:"type" validate:"required,oneof=bug feature other"`
	Message string `json:"message" validate:"required,max=2000"`
	PageURL string `json:"page_url" validate:"max=2000"`
}
