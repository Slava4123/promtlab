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
// Reason опционален — для exit-survey (M-6b).
type CancelInput struct {
	UserID uint
	Reason string
	Other  string
}

// PauseInput — входные данные для паузы подписки (M-6). Months 1..3.
type PauseInput struct {
	UserID uint
	Months int
}

// CancelReason — допустимые причины отмены для exit-survey (M-6b).
// Валидируются на Go-стороне, в БД хранится как varchar(30).
const (
	CancelReasonTooExpensive     = "too_expensive"
	CancelReasonNotUsing         = "not_using"
	CancelReasonMissingFeature   = "missing_feature"
	CancelReasonFoundAlternative = "found_alternative"
	CancelReasonOther            = "other"
)

// IsValidCancelReason — true если reason пустой (опционально) или из списка выше.
func IsValidCancelReason(reason string) bool {
	switch reason {
	case "", CancelReasonTooExpensive, CancelReasonNotUsing, CancelReasonMissingFeature,
		CancelReasonFoundAlternative, CancelReasonOther:
		return true
	}
	return false
}

// PaymentProviderData — унифицированная схема для Payment.ProviderData JSONB-колонки.
// Пишется в renewal.go (и checkout'е) и читается в extractPlanID / isRenewalPayment.
// Renewal использует строковый "true"/"false" для совместимости с ранее записанными
// платежами (renewal.go:141 писал map[string]string).
type PaymentProviderData struct {
	PlanID  string `json:"plan_id"`
	Renewal string `json:"renewal,omitempty"`
}
