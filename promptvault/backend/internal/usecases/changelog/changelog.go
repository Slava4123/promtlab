package changelog

import (
	"context"
	"errors"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
)

// Service — бизнес-логика changelog'а.
//
// Записи загружаются из embedded JSON один раз при старте и хранятся в памяти.
// Непрочитанность определяется сравнением user.LastChangelogSeenAt с датой
// последней записи.
type Service struct {
	changelog *Changelog
	users     repo.UserRepository
}

func NewService(users repo.UserRepository) (*Service, error) {
	c, err := loadEmbeddedChangelog()
	if err != nil {
		return nil, err
	}
	return &Service{
		changelog: c,
		users:     users,
	}, nil
}

// List возвращает все записи changelog'а и флаг непрочитанных для юзера.
func (s *Service) List(ctx context.Context, userID uint) (*ChangelogOutput, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	hasUnread := s.computeHasUnread(user.LastChangelogSeenAt)

	return &ChangelogOutput{
		Entries:   s.changelog.Entries,
		HasUnread: hasUnread,
	}, nil
}

// MarkSeen обновляет user.LastChangelogSeenAt на текущее время.
func (s *Service) MarkSeen(ctx context.Context, userID uint) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	user.LastChangelogSeenAt = &now
	return s.users.Update(ctx, user)
}

// HasUnread возвращает true, если у юзера есть непрочитанные записи.
// Используется в /auth/me для badge-индикатора.
func (s *Service) HasUnread(ctx context.Context, userID uint) (bool, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return s.computeHasUnread(user.LastChangelogSeenAt), nil
}

// latestDate возвращает дату самой свежей записи changelog'а.
func (s *Service) latestDate() string {
	if len(s.changelog.Entries) == 0 {
		return ""
	}
	return s.changelog.Entries[0].Date
}

// computeHasUnread проверяет, есть ли записи новее LastChangelogSeenAt.
func (s *Service) computeHasUnread(lastSeen *time.Time) bool {
	latest := s.latestDate()
	if latest == "" {
		return false
	}
	if lastSeen == nil {
		return true
	}
	latestTime, err := time.Parse("2006-01-02", latest)
	if err != nil {
		slog.Error("changelog.date_parse_failed", "date", latest, "error", err)
		return false
	}
	return latestTime.After(*lastSeen)
}
