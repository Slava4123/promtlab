package subscription

import (
	"encoding/json"
	"time"

	"promptvault/internal/models"
	quotauc "promptvault/internal/usecases/quota"
)

// PlanResponse — DTO тарифного плана для API.
type PlanResponse struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	PriceKop           int             `json:"price_kop"`
	PeriodDays         int             `json:"period_days"`
	MaxPrompts         int             `json:"max_prompts"`
	MaxCollections     int             `json:"max_collections"`
	MaxAIRequestsDaily int             `json:"max_ai_requests_daily"`
	AIRequestsIsTotal  bool            `json:"ai_requests_is_total"`
	MaxTeams           int             `json:"max_teams"`
	MaxTeamMembers     int             `json:"max_team_members"`
	MaxShareLinks      int             `json:"max_share_links"`
	MaxExtUsesDaily    int             `json:"max_ext_uses_daily"`
	MaxMCPUsesDaily    int             `json:"max_mcp_uses_daily"`
	Features           json.RawMessage `json:"features"`
	SortOrder          int             `json:"sort_order"`
}

// NewPlanResponse конвертирует модель плана в DTO.
func NewPlanResponse(p models.SubscriptionPlan) PlanResponse {
	return PlanResponse{
		ID:                 p.ID,
		Name:               p.Name,
		PriceKop:           p.PriceKop,
		PeriodDays:         p.PeriodDays,
		MaxPrompts:         p.MaxPrompts,
		MaxCollections:     p.MaxCollections,
		MaxAIRequestsDaily: p.MaxAIRequestsDaily,
		AIRequestsIsTotal:  p.AIRequestsIsTotal,
		MaxTeams:           p.MaxTeams,
		MaxTeamMembers:     p.MaxTeamMembers,
		MaxShareLinks:      p.MaxShareLinks,
		MaxExtUsesDaily:    p.MaxExtUsesDaily,
		MaxMCPUsesDaily:    p.MaxMCPUsesDaily,
		Features:           p.Features,
		SortOrder:          p.SortOrder,
	}
}

// NewPlansResponse конвертирует список планов в DTO.
func NewPlansResponse(plans []models.SubscriptionPlan) []PlanResponse {
	out := make([]PlanResponse, 0, len(plans))
	for _, p := range plans {
		out = append(out, NewPlanResponse(p))
	}
	return out
}

// SubscriptionResponse — DTO подписки для API.
type SubscriptionResponse struct {
	ID                 uint          `json:"id"`
	PlanID             string        `json:"plan_id"`
	Status             string        `json:"status"`
	CurrentPeriodStart time.Time     `json:"current_period_start"`
	CurrentPeriodEnd   time.Time     `json:"current_period_end"`
	CancelAtPeriodEnd  bool          `json:"cancel_at_period_end"`
	CancelledAt        *time.Time    `json:"cancelled_at,omitempty"`
	AutoRenew          bool          `json:"auto_renew"`
	PausedAt           *time.Time    `json:"paused_at,omitempty"`
	PausedUntil        *time.Time    `json:"paused_until,omitempty"`
	Plan               *PlanResponse `json:"plan,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
}

// NewSubscriptionResponse конвертирует модель подписки в DTO.
func NewSubscriptionResponse(s *models.Subscription) *SubscriptionResponse {
	if s == nil {
		return nil
	}
	resp := &SubscriptionResponse{
		ID:                 s.ID,
		PlanID:             s.PlanID,
		Status:             string(s.Status),
		CurrentPeriodStart: s.CurrentPeriodStart,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		CancelledAt:        s.CancelledAt,
		AutoRenew:          s.AutoRenew,
		PausedAt:           s.PausedAt,
		PausedUntil:        s.PausedUntil,
		CreatedAt:          s.CreatedAt,
	}
	if s.Plan.ID != "" {
		pr := NewPlanResponse(s.Plan)
		resp.Plan = &pr
	}
	return resp
}

// CheckoutResponse — ответ на POST /api/subscription/checkout.
type CheckoutResponse struct {
	PaymentURL string `json:"payment_url"`
}

// UsageResponse — DTO сводки использования для API.
type UsageResponse struct {
	PlanID       string             `json:"plan_id"`
	Prompts      quotauc.QuotaInfo  `json:"prompts"`
	Collections  quotauc.QuotaInfo  `json:"collections"`
	AIRequests   quotauc.QuotaInfo  `json:"ai_requests"`
	Teams        quotauc.QuotaInfo  `json:"teams"`
	ShareLinks   quotauc.QuotaInfo  `json:"share_links"`
	ExtUsesToday quotauc.QuotaInfo  `json:"ext_uses_today"`
	MCPUsesToday quotauc.QuotaInfo  `json:"mcp_uses_today"`
}

// NewUsageResponse конвертирует UsageSummary в DTO.
func NewUsageResponse(s *quotauc.UsageSummary) UsageResponse {
	return UsageResponse{
		PlanID:       s.PlanID,
		Prompts:      s.Prompts,
		Collections:  s.Collections,
		AIRequests:   s.AIRequests,
		Teams:        s.Teams,
		ShareLinks:   s.ShareLinks,
		ExtUsesToday: s.ExtUsesToday,
		MCPUsesToday: s.MCPUsesToday,
	}
}
