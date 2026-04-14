package subscription

// CheckoutRequest — body POST /api/subscription/checkout.
type CheckoutRequest struct {
	PlanID string `json:"plan_id" validate:"required"`
}

// AutoRenewRequest — body POST /api/subscription/auto-renew.
type AutoRenewRequest struct {
	AutoRenew *bool `json:"auto_renew" validate:"required"`
}
