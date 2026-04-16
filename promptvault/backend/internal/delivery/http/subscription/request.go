package subscription

// CheckoutRequest — body POST /api/subscription/checkout.
type CheckoutRequest struct {
	PlanID string `json:"plan_id" validate:"required"`
}

// AutoRenewRequest — body POST /api/subscription/auto-renew.
type AutoRenewRequest struct {
	AutoRenew *bool `json:"auto_renew" validate:"required"`
}

// CancelRequest — body POST /api/subscription/cancel. Для M-6b exit-survey.
// reason и other_text опциональны; если reason задан — пишем в subscription_cancellations.
// reason должен быть из {too_expensive,not_using,missing_feature,found_alternative,other}.
type CancelRequest struct {
	Reason    string `json:"reason,omitempty" validate:"omitempty,oneof=too_expensive not_using missing_feature found_alternative other"`
	OtherText string `json:"other_text,omitempty" validate:"omitempty,max=500"`
}

// PauseRequest — body POST /api/subscription/pause. M-6.
type PauseRequest struct {
	Months int `json:"months" validate:"required,min=1,max=3"`
}
