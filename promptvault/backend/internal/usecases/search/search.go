package search

import (
	"context"
	"strings"

	repo "promptvault/internal/interface/repository"
)

const (
	maxPrompts     = 5
	maxCollections = 3
	maxTags        = 3
	maxDescription = 120
)

type Service struct {
	prompts     repo.PromptRepository
	collections repo.CollectionRepository
	tags        repo.TagRepository
}

func NewService(
	prompts repo.PromptRepository,
	collections repo.CollectionRepository,
	tags repo.TagRepository,
) *Service {
	return &Service{
		prompts:     prompts,
		collections: collections,
		tags:        tags,
	}
}

func (s *Service) Search(ctx context.Context, userID uint, teamID *uint, query string) (*SearchOutput, error) {
	query = strings.TrimSpace(query)

	out := &SearchOutput{
		Prompts:     []SearchResult{},
		Collections: []SearchResult{},
		Tags:        []SearchResult{},
	}

	if query == "" {
		return out, nil
	}

	// Промпты
	prompts, err := s.prompts.SearchByQuery(ctx, userID, teamID, query, maxPrompts)
	if err != nil {
		return nil, err
	}
	for _, p := range prompts {
		desc := p.Content
		if len([]rune(desc)) > maxDescription {
			desc = string([]rune(desc)[:maxDescription]) + "..."
		}
		out.Prompts = append(out.Prompts, SearchResult{
			ID:          p.ID,
			Type:        "prompt",
			Title:       p.Title,
			Description: desc,
		})
	}

	// Коллекции
	collections, err := s.collections.SearchByQuery(ctx, userID, teamID, query, maxCollections)
	if err != nil {
		return nil, err
	}
	for _, c := range collections {
		out.Collections = append(out.Collections, SearchResult{
			ID:    c.ID,
			Type:  "collection",
			Title: c.Name,
			Color: c.Color,
			Icon:  c.Icon,
		})
	}

	// Теги
	tags, err := s.tags.SearchByQuery(ctx, userID, teamID, query, maxTags)
	if err != nil {
		return nil, err
	}
	for _, t := range tags {
		out.Tags = append(out.Tags, SearchResult{
			ID:    t.ID,
			Type:  "tag",
			Title: t.Name,
			Color: t.Color,
		})
	}

	return out, nil
}
