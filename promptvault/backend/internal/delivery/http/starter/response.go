package starter

import (
	"time"

	prompthttp "promptvault/internal/delivery/http/prompt"
	starteruc "promptvault/internal/usecases/starter"
)

// CatalogResponse — то же самое что starter.Catalog. Возвращается как-есть,
// никаких дополнительных полей не добавляется.
type CatalogResponse struct {
	Version    int                `json:"version"`
	Lang       string             `json:"lang"`
	Categories []categoryResponse `json:"categories"`
	Templates  []templateResponse `json:"templates"`
}

type categoryResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	UseCases    []string `json:"use_cases"`
}

type templateResponse struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Model    string `json:"model"`
}

func NewCatalogResponse(c *starteruc.Catalog) CatalogResponse {
	cats := make([]categoryResponse, 0, len(c.Categories))
	for _, cat := range c.Categories {
		cats = append(cats, categoryResponse{
			ID:          cat.ID,
			Name:        cat.Name,
			Description: cat.Description,
			Icon:        cat.Icon,
			UseCases:    cat.UseCases,
		})
	}
	tpls := make([]templateResponse, 0, len(c.Templates))
	for _, t := range c.Templates {
		tpls = append(tpls, templateResponse{
			ID:       t.ID,
			Category: t.Category,
			Title:    t.Title,
			Content:  t.Content,
			Model:    t.Model,
		})
	}
	return CatalogResponse{
		Version:    c.Version,
		Lang:       c.Lang,
		Categories: cats,
		Templates:  tpls,
	}
}

// CompleteResponse — body 200 OK после POST /api/starter/complete.
type CompleteResponse struct {
	Installed             []prompthttp.PromptResponse `json:"installed"`
	OnboardingCompletedAt time.Time                   `json:"onboarding_completed_at"`
}

func NewCompleteResponse(result *starteruc.InstallResult) CompleteResponse {
	installed := make([]prompthttp.PromptResponse, 0, len(result.Prompts))
	for _, p := range result.Prompts {
		installed = append(installed, prompthttp.NewPromptResponse(*p))
	}
	return CompleteResponse{
		Installed:             installed,
		OnboardingCompletedAt: result.CompletedAt,
	}
}
