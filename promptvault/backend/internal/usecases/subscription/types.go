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
// Reason опционален — для exit-survey (M-6): too_expensive/not_using/missing_feature/found_alternative/other.
type CancelInput struct {
	UserID uint
	Reason string
	Other  string
}

// PaymentProviderData — унифицированная схема для Payment.ProviderData JSONB-колонки.
// Пишется в renewal.go (и checkout'е) и читается в extractPlanID / isRenewalPayment.
// Renewal использует строковый "true"/"false" для совместимости с ранее записанными
// платежами (renewal.go:141 писал map[string]string).
type PaymentProviderData struct {
	PlanID  string `json:"plan_id"`
	Renewal string `json:"renewal,omitempty"`
}
