package feedback

// SubmitInput — входные данные для отправки обратной связи.
type SubmitInput struct {
	UserID  uint
	Type    string
	Message string
	PageURL string
}

// SubmitResult — результат успешной отправки.
type SubmitResult struct {
	ID uint
}
