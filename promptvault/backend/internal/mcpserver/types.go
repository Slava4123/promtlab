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
	ID            uint      `json:"id"`
	VersionNumber uint      `json:"version_number"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Model         string    `json:"model,omitempty"`
	ChangeNote    string    `json:"change_note,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
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
