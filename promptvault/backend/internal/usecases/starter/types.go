package starter

import (
	"time"

	"promptvault/internal/models"
)

// Catalog — десериализованное содержимое catalog.json. Загружается один раз
// при старте сервиса (см. embed.go) и хранится в памяти.
type Catalog struct {
	Version    int        `json:"version"`
	Lang       string     `json:"lang"`
	Categories []Category `json:"categories"`
	Templates  []Template `json:"templates"`
}

// Category — секция wizard'а: визуальное разделение для юзера. id используется
// для фильтрации templates на фронте.
type Category struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	UseCases    []string `json:"use_cases"`
}

// Template — единица starter каталога. При install создаётся как обычный
// models.Prompt в личном workspace юзера (TeamID = nil).
type Template struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Model    string `json:"model"`
}

// InstallResult — что Service.Install возвращает HTTP-слою.
type InstallResult struct {
	Prompts     []*models.Prompt
	CompletedAt time.Time
}
