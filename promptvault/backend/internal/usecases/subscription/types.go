package subscription

// CheckoutInput — входные данные для создания платежа.
type CheckoutInput struct {
	UserID uint
	PlanID string
}

// CheckoutResult — результат создания платежа.
type CheckoutResult struct {
	PaymentURL string
}

// CancelInput — входные данные для отмены подписки.
type CancelInput struct {
	UserID uint
}
