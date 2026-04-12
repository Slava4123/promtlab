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

	suggestPrompts     = 4
	suggestCollections = 2
	suggestTags        = 1
	maxSuggestions     = 7
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

func (s *Service) Suggest(ctx context.Context, userID uint, teamID *uint, prefix string) (*SuggestOutput, error) {
	prefix = strings.TrimSpace(prefix)

	out := &SuggestOutput{Suggestions: []Suggestion{}}
	if prefix == "" {
		return out, nil
	}

	promptTitles, err := s.prompts.SuggestByPrefix(ctx, userID, teamID, prefix, suggestPrompts)
	if err != nil {
		return nil, err
	}

	collNames, err := s.collections.SuggestByPrefix(ctx, userID, teamID, prefix, suggestCollections)
	if err != nil {
		return nil, err
	}

	tagNames, err := s.tags.SuggestByPrefix(ctx, userID, teamID, prefix, suggestTags)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	add := func(text, typ string) {
		key := strings.ToLower(text)
		if _, dup := seen[key]; dup {
			return
		}
		seen[key] = struct{}{}
		out.Suggestions = append(out.Suggestions, Suggestion{Text: text, Type: typ})
	}

	for _, t := range promptTitles {
		add(t, "prompt")
	}
	for _, n := range collNames {
		add(n, "collection")
	}
	for _, n := range tagNames {
		add(n, "tag")
	}

	if len(out.Suggestions) > maxSuggestions {
		out.Suggestions = out.Suggestions[:maxSuggestions]
	}

	return out, nil
}
