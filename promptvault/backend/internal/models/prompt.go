package models

import (
	"time"

	"gorm.io/gorm"
)

type Prompt struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	TeamID      *uint          `gorm:"index" json:"team_id,omitempty"`
	Title       string         `gorm:"size:300;not null" json:"title"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Model       string         `gorm:"size:100" json:"model,omitempty"`
	Favorite    bool           `gorm:"default:false" json:"favorite"`
	UsageCount  int            `gorm:"default:0" json:"usage_count"`
	// IsPublic / Slug — публичный SEO-URL /p/:slug. Отличается от share-link:
	// публичный индексируется, share-link — приватный по токену.
	IsPublic    bool           `gorm:"column:is_public;not null;default:false" json:"is_public"`
	Slug        string         `gorm:"column:slug;size:120" json:"slug,omitempty"`
	// SlugAliases — JSONB-массив строк со старыми slug'ами этого промпта.
	// Phase 15: backward-compat для resluggified cyrillic-titles. Lookup
	// в share-handler делает `slug = ? OR slug_aliases @> ?`.
	SlugAliases []string       `gorm:"column:slug_aliases;serializer:json;type:jsonb;default:'[]'" json:"-"`
	LastUsedAt  *time.Time     `gorm:"" json:"last_used_at,omitempty"`
	User        User           `gorm:"foreignKey:UserID" json:"-"`
	Tags        []Tag            `gorm:"many2many:prompt_tags" json:"tags,omitempty"`
	Collections []Collection     `gorm:"many2many:prompt_collections" json:"collections,omitempty"`
	Versions    []PromptVersion  `gorm:"foreignKey:PromptID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// NewPrompt — конструктор Prompt с обязательными полями (UserID, Title, Content)
// и явными дефолтами. MN-32: до этого callers использовали struct literal —
// легко забыть init поля, легко неправильно поставить Favorite/UsageCount.
//
// teamID — указатель: nil для personal, non-nil для team-цепочки.
func NewPrompt(userID uint, title, content string, teamID *uint) *Prompt {
	return &Prompt{
		UserID:      userID,
		TeamID:      teamID,
		Title:       title,
		Content:     content,
		Favorite:    false,
		UsageCount:  0,
		IsPublic:    false,
		SlugAliases: []string{},
	}
}

// BeforeSave — для пустого Slug омитим колонку из INSERT/UPDATE.
//
// Партиальный unique-index `idx_prompts_slug ON prompts (slug) WHERE slug IS NOT NULL`
// (миграция 000034) предполагает NULL для непубличных промптов. Но Go-string
// шлёт `''` вместо NULL, и индекс считает пустую строку валидным значением —
// при втором непубличном промпте получаем 23505 duplicate key.
// Omit убирает колонку из statement — БД использует дефолт (NULL).
func (p *Prompt) BeforeSave(tx *gorm.DB) error {
	if p.Slug == "" {
		tx.Statement.Omit("Slug")
	}
	return nil
}
