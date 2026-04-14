package admin

import (
	"cmp"
	"time"

	repo "promptvault/internal/interface/repository"
	badgeuc "promptvault/internal/usecases/badge"
)

// UserSummaryResponse — элемент списка /admin/users.
// ParseFields остаются на стороне frontend — тут просто mapping.
type UserSummaryResponse struct {
	ID            uint      `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Username      string    `json:"username,omitempty"`
	Role          string    `json:"role"`
	Status        string    `json:"status"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewUserSummaryResponse(u repo.UserSummary) UserSummaryResponse {
	return UserSummaryResponse{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		Username:      u.Username,
		Role:          u.Role,
		Status:        u.Status,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
	}
}

// UserDetailResponse — детальная страница /admin/users/{id}.
type UserDetailResponse struct {
	ID               uint      `json:"id"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	Username         string    `json:"username,omitempty"`
	AvatarURL        string    `json:"avatar_url,omitempty"`
	Role             string    `json:"role"`
	Status           string    `json:"status"`
	EmailVerified    bool      `json:"email_verified"`
	DefaultModel     string    `json:"default_model"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PromptCount      int64     `json:"prompt_count"`
	CollectionCount  int64     `json:"collection_count"`
	BadgeCount       int64     `json:"badge_count"`
	TotalUsage       int64     `json:"total_usage"`
	LinkedProviders  []string  `json:"linked_providers"`
	UnlockedBadgeIDs []string  `json:"unlocked_badge_ids"` // для admin UI — отличать unlocked vs locked
	// Tier — stub. Всегда "free" пока subscription system не появится.
	Tier string `json:"tier"`
}

func NewUserDetailResponse(d *repo.UserDetail) UserDetailResponse {
	u := d.User
	providers := d.LinkedProviders
	if providers == nil {
		providers = []string{}
	}
	unlockedIDs := d.UnlockedBadgeIDs
	if unlockedIDs == nil {
		unlockedIDs = []string{}
	}
	return UserDetailResponse{
		ID:               u.ID,
		Email:            u.Email,
		Name:             u.Name,
		Username:         u.Username,
		AvatarURL:        u.AvatarURL,
		Role:             string(u.Role),
		Status:           string(u.Status),
		EmailVerified:    u.EmailVerified,
		DefaultModel:     u.DefaultModel,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
		PromptCount:      d.PromptCount,
		CollectionCount:  d.CollectionCount,
		BadgeCount:       d.BadgeCount,
		TotalUsage:       d.TotalUsage,
		LinkedProviders:  providers,
		UnlockedBadgeIDs: unlockedIDs,
		Tier:             cmp.Or(d.User.PlanID, "free"),
	}
}

// ActionResponse — подтверждение успешного destructive action.
// `ok: true` + action name для UX («Пользователь заморожен»).
type ActionResponse struct {
	OK     bool   `json:"ok"`
	Action string `json:"action"`
}

// GrantBadgeResponse — подтверждение grant с деталями бейджа.
type GrantBadgeResponse struct {
	OK    bool              `json:"ok"`
	Badge GrantedBadgeBrief `json:"badge"`
}

type GrantedBadgeBrief struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

func NewGrantBadgeResponse(b *badgeuc.Badge) GrantBadgeResponse {
	return GrantBadgeResponse{
		OK: true,
		Badge: GrantedBadgeBrief{
			ID:    b.ID,
			Title: b.Title,
			Icon:  b.Icon,
		},
	}
}

// AuditEntryResponse — элемент GET /api/admin/audit.
type AuditEntryResponse struct {
	ID          uint            `json:"id"`
	AdminID     uint            `json:"admin_id"`
	Action      string          `json:"action"`
	TargetType  string          `json:"target_type"`
	TargetID    *uint           `json:"target_id,omitempty"`
	BeforeState any             `json:"before_state,omitempty"`
	AfterState  any             `json:"after_state,omitempty"`
	IP          string          `json:"ip"`
	UserAgent   string          `json:"user_agent,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// HealthResponse — GET /api/admin/health. Простой health-check, пока без
// глубоких метрик — будет расширяться по мере необходимости.
type HealthResponse struct {
	Status      string    `json:"status"`
	Time        time.Time `json:"time"`
	TotalUsers  int64     `json:"total_users"`
	AdminUsers  int64     `json:"admin_users"`
	ActiveUsers int64     `json:"active_users"`
	FrozenUsers int64     `json:"frozen_users"`
}
