package mcpserver

import "time"

// MCP response DTOs — не экспортируют user_id, team_id, deleted_at

type PromptResponse struct {
	ID          uint                 `json:"id"`
	Title       string               `json:"title"`
	Content     string               `json:"content"`
	Model       string               `json:"model,omitempty"`
	Favorite    bool                 `json:"favorite"`
	UsageCount  int                  `json:"usage_count"`
	Tags        []TagResponse        `json:"tags"`
	Collections []CollectionResponse `json:"collections"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type CollectionResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color"`
	Icon        string `json:"icon,omitempty"`
}

type CollectionWithCountResponse struct {
	CollectionResponse
	PromptCount int64 `json:"prompt_count"`
}

type TagResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type VersionResponse struct {
	ID             uint      `json:"id"`
	VersionNumber  uint      `json:"version_number"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Model          string    `json:"model,omitempty"`
	ChangeNote     string    `json:"change_note,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	// Phase 14: автор версии (nil для старых записей без changed_by).
	ChangedByID    *uint     `json:"changed_by_id,omitempty"`
	ChangedByEmail string    `json:"changed_by_email,omitempty"`
	ChangedByName  string    `json:"changed_by_name,omitempty"`
}

// ActivityItemResponse — запись team_activity_log для feed/history (Phase 14).
type ActivityItemResponse struct {
	ID          uint           `json:"id"`
	TeamID      uint           `json:"team_id"`
	ActorID     *uint          `json:"actor_id,omitempty"`
	ActorEmail  string         `json:"actor_email"`
	ActorName   string         `json:"actor_name,omitempty"`
	EventType   string         `json:"event_type"`
	TargetType  string         `json:"target_type"`
	TargetID    *uint          `json:"target_id,omitempty"`
	TargetLabel string         `json:"target_label,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// AnalyticsSummaryResponse — компактная сводка для MCP (Phase 14).
// Дашборд-endpoints отдают полные данные, MCP — top-N + агрегаты для AI-ассистента.
type AnalyticsSummaryResponse struct {
	Range        string               `json:"range"`
	TotalUses    int64                `json:"total_uses"`
	TotalViews   int64                `json:"total_views"`
	TopPrompts   []PromptUsageSummary `json:"top_prompts,omitempty"`
	Contributors []ContributorSummary `json:"contributors,omitempty"` // team scope only
}

type PromptUsageSummary struct {
	PromptID uint   `json:"prompt_id"`
	Title    string `json:"title"`
	Uses     int64  `json:"uses"`
}

type ContributorSummary struct {
	UserID         uint   `json:"user_id"`
	Name           string `json:"name"`
	PromptsCreated int64  `json:"prompts_created"`
	PromptsEdited  int64  `json:"prompts_edited"`
	Uses           int64  `json:"uses"`
}

type PinResultResponse struct {
	Pinned   bool `json:"pinned"`
	TeamWide bool `json:"team_wide"`
}

type ShareLinkResponse struct {
	ID           uint       `json:"id"`
	Token        string     `json:"token"`
	URL          string     `json:"url"`
	IsActive     bool       `json:"is_active"`
	ViewCount    int        `json:"view_count"`
	LastViewedAt *time.Time `json:"last_viewed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type TeamResponse struct {
	ID          uint      `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Role        string    `json:"role"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserResponse struct {
	ID            uint   `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Username      string `json:"username,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	PlanID        string `json:"plan_id"`
	EmailVerified bool   `json:"email_verified"`
}
