package ai

type RewriteStyle string

const (
	StyleFormal     RewriteStyle = "formal"
	StyleConcise    RewriteStyle = "concise"
	StyleCreative   RewriteStyle = "creative"
	StyleDetailed   RewriteStyle = "detailed"
	StyleTechnical  RewriteStyle = "technical"
)

type EnhanceInput struct {
	UserID  uint
	Content string
	Model   string
}

type RewriteInput struct {
	UserID  uint
	Content string
	Model   string
	Style   RewriteStyle
}

type AnalyzeInput struct {
	UserID  uint
	Content string
	Model   string
}

type VariationsInput struct {
	UserID  uint
	Content string
	Model   string
	Count   int // default 3
}
