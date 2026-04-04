package ai

type EnhanceRequest struct {
	Content string `json:"content" validate:"required,max=10000"`
	Model   string `json:"model"   validate:"required"`
}

type RewriteRequest struct {
	Content string `json:"content" validate:"required,max=10000"`
	Model   string `json:"model"   validate:"required"`
	Style   string `json:"style"   validate:"required,oneof=formal concise creative detailed technical"`
}

type AnalyzeRequest struct {
	Content string `json:"content" validate:"required,max=10000"`
	Model   string `json:"model"   validate:"required"`
}

type VariationsRequest struct {
	Content string `json:"content" validate:"required,max=10000"`
	Model   string `json:"model"   validate:"required"`
	Count   int    `json:"count"   validate:"omitempty,min=1,max=5"`
}
