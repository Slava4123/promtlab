package share

import (
	"time"

	"promptvault/internal/models"
)

type ShareLinkInfo struct {
	ID           uint       `json:"id"`
	Token        string     `json:"token"`
	URL          string     `json:"url"`
	IsActive     bool       `json:"is_active"`
	ViewCount    int        `json:"view_count"`
	LastViewedAt *time.Time `json:"last_viewed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type PublicPromptInfo struct {
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	Model     string       `json:"model,omitempty"`
	Tags      []PublicTag  `json:"tags"`
	Author    PublicAuthor `json:"author"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	// Phase 14 Branded share pages (Max-only). nil для Free/Pro и не-team промптов.
	Branding *models.BrandingInfo `json:"branding,omitempty"`
}

type PublicTag struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type PublicAuthor struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}
