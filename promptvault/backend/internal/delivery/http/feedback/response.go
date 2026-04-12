package feedback

// SubmitResponse — ответ на успешную отправку обратной связи.
type SubmitResponse struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}
