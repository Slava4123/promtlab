package prompt

import (
	"time"

	"promptvault/internal/models"
)

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

type TagResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type CollectionResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon,omitempty"`
}

func NewPromptResponse(p models.Prompt) PromptResponse {
	tags := make([]TagResponse, 0, len(p.Tags))
	for _, t := range p.Tags {
		tags = append(tags, TagResponse{ID: t.ID, Name: t.Name, Color: t.Color})
	}

	cols := make([]CollectionResponse, 0, len(p.Collections))
	for _, c := range p.Collections {
		cols = append(cols, CollectionResponse{ID: c.ID, Name: c.Name, Color: c.Color, Icon: c.Icon})
	}

	return PromptResponse{
		ID:          p.ID,
		Title:       p.Title,
		Content:     p.Content,
		Model:       p.Model,
		Favorite:    p.Favorite,
		UsageCount:  p.UsageCount,
		Tags:        tags,
		Collections: cols,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func NewPromptListResponse(prompts []models.Prompt) []PromptResponse {
	res := make([]PromptResponse, 0, len(prompts))
	for _, p := range prompts {
		res = append(res, NewPromptResponse(p))
	}
	return res
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

func NewVersionResponse(v models.PromptVersion) VersionResponse {
	return VersionResponse{
		ID:            v.ID,
		VersionNumber: v.VersionNumber,
		Title:         v.Title,
		Content:       v.Content,
		Model:         v.Model,
		ChangeNote:    v.ChangeNote,
		CreatedAt:     v.CreatedAt,
	}
}

func NewVersionListResponse(versions []models.PromptVersion) []VersionResponse {
	res := make([]VersionResponse, 0, len(versions))
	for _, v := range versions {
		res = append(res, NewVersionResponse(v))
	}
	return res
}
