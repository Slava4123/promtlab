package user

import (
	"context"
	"strings"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

type Service struct {
	users repo.UserRepository
}

func NewService(users repo.UserRepository) *Service {
	return &Service{users: users}
}

type SearchResult struct {
	ID        uint
	Name      string
	Username  string
	AvatarURL string
	Email     string
}

func (s *Service) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if len(query) < 2 {
		return nil, nil
	}
	users, err := s.users.SearchUsers(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	results := make([]SearchResult, 0, len(users))
	for _, u := range users {
		results = append(results, SearchResult{
			ID:        u.ID,
			Name:      u.Name,
			Username:  u.Username,
			AvatarURL: u.AvatarURL,
			Email:     maskEmail(u.Email),
		})
	}
	return results, nil
}

// GetByID возвращает пользователя по ID. Используется, например, MCP whoami.
func (s *Service) GetByID(ctx context.Context, id uint) (*models.User, error) {
	return s.users.GetByID(ctx, id)
}

func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "***"
	}
	return string(parts[0][0]) + "***@" + parts[1]
}
