package share

import (
	"time"

	shareuc "promptvault/internal/usecases/share"
)

type ShareLinkResponse struct {
	ID           uint       `json:"id"`
	Token        string     `json:"token"`
	URL          string     `json:"url"`
	IsActive     bool       `json:"is_active"`
	ViewCount    int        `json:"view_count"`
	LastViewedAt *time.Time `json:"last_viewed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type PublicPromptResponse struct {
	Title     string              `json:"title"`
	Content   string              `json:"content"`
	Model     string              `json:"model,omitempty"`
	Tags      []PublicTagResponse `json:"tags"`
	Author    AuthorResponse      `json:"author"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	// MN-31: BrandingDTO вместо *models.BrandingInfo. Раньше DTO leak — JSON
	// shape публичного API зависит от GORM-модели; миграция/добавление поля
	// в модель меняет contract без явного версионирования.
	Branding *BrandingDTO `json:"branding,omitempty"`
}

// BrandingDTO — public-shape брендинг информация для /s/:token.
// MN-31: явный transport-DTO, JSON-теги стабильны независимо от models.BrandingInfo.
type BrandingDTO struct {
	LogoSource       string `json:"logo_source,omitempty"`
	EffectiveLogoURL string `json:"effective_logo_url,omitempty"`
	Tagline          string `json:"tagline,omitempty"`
	Website          string `json:"website,omitempty"`
	PrimaryColor     string `json:"primary_color,omitempty"`
}

type PublicTagResponse struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type AuthorResponse struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

func toShareLinkResponse(info *shareuc.ShareLinkInfo) ShareLinkResponse {
	return ShareLinkResponse{
		ID:           info.ID,
		Token:        info.Token,
		URL:          info.URL,
		IsActive:     info.IsActive,
		ViewCount:    info.ViewCount,
		LastViewedAt: info.LastViewedAt,
		CreatedAt:    info.CreatedAt,
	}
}

func toPublicPromptResponse(info *shareuc.PublicPromptInfo) PublicPromptResponse {
	tags := make([]PublicTagResponse, len(info.Tags))
	for i, t := range info.Tags {
		tags[i] = PublicTagResponse{Name: t.Name, Color: t.Color}
	}
	var branding *BrandingDTO
	if info.Branding != nil {
		branding = &BrandingDTO{
			LogoSource:       info.Branding.LogoSource,
			EffectiveLogoURL: info.Branding.EffectiveLogoURL,
			Tagline:          info.Branding.Tagline,
			Website:          info.Branding.Website,
			PrimaryColor:     info.Branding.PrimaryColor,
		}
	}
	return PublicPromptResponse{
		Title:     info.Title,
		Content:   info.Content,
		Model:     info.Model,
		Tags:      tags,
		Author:    AuthorResponse{Name: info.Author.Name, AvatarURL: info.Author.AvatarURL},
		CreatedAt: info.CreatedAt,
		UpdatedAt: info.UpdatedAt,
		Branding:  branding,
	}
}
