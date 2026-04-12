package trash

import (
	"math"
	"time"

	"promptvault/internal/models"
)

type TrashPromptResponse struct {
	ID         uint               `json:"id"`
	Title      string             `json:"title"`
	Content    string             `json:"content"`
	Model      string             `json:"model,omitempty"`
	Favorite   bool               `json:"favorite"`
	Tags       []TagBrief         `json:"tags"`
	Collections []CollectionBrief `json:"collections"`
	DeletedAt  time.Time          `json:"deleted_at"`
	CreatedAt  time.Time          `json:"created_at"`
	DaysLeft   int                `json:"days_left"`
}

type TrashCollectionResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon,omitempty"`
	DeletedAt   time.Time `json:"deleted_at"`
	DaysLeft    int       `json:"days_left"`
}

type TagBrief struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type CollectionBrief struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Color string `json:"color"`
}

const retentionDays = 30

func daysLeft(deletedAt time.Time) int {
	expires := deletedAt.AddDate(0, 0, retentionDays)
	left := int(math.Ceil(time.Until(expires).Hours() / 24))
	if left < 0 {
		return 0
	}
	return left
}

func NewTrashPromptResponse(p models.Prompt) TrashPromptResponse {
	tags := make([]TagBrief, 0, len(p.Tags))
	for _, t := range p.Tags {
		tags = append(tags, TagBrief{ID: t.ID, Name: t.Name, Color: t.Color})
	}
	cols := make([]CollectionBrief, 0, len(p.Collections))
	for _, c := range p.Collections {
		cols = append(cols, CollectionBrief{ID: c.ID, Name: c.Name, Color: c.Color})
	}
	return TrashPromptResponse{
		ID:          p.ID,
		Title:       p.Title,
		Content:     p.Content,
		Model:       p.Model,
		Favorite:    p.Favorite,
		Tags:        tags,
		Collections: cols,
		DeletedAt:   p.DeletedAt.Time,
		CreatedAt:   p.CreatedAt,
		DaysLeft:    daysLeft(p.DeletedAt.Time),
	}
}

func NewTrashPromptListResponse(prompts []models.Prompt) []TrashPromptResponse {
	res := make([]TrashPromptResponse, 0, len(prompts))
	for _, p := range prompts {
		res = append(res, NewTrashPromptResponse(p))
	}
	return res
}

func NewTrashCollectionResponse(c models.Collection) TrashCollectionResponse {
	return TrashCollectionResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Color:       c.Color,
		Icon:        c.Icon,
		DeletedAt:   c.DeletedAt.Time,
		DaysLeft:    daysLeft(c.DeletedAt.Time),
	}
}

func NewTrashCollectionListResponse(cols []models.Collection) []TrashCollectionResponse {
	res := make([]TrashCollectionResponse, 0, len(cols))
	for _, c := range cols {
		res = append(res, NewTrashCollectionResponse(c))
	}
	return res
}

