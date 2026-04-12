package share

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	"promptvault/internal/usecases/teamcheck"
)

const (
	tokenPrefix     = "ps_"
	tokenRandBytes  = 16 // 128 bits of entropy
	viewCountTimeout = 5 * time.Second
)

type Service struct {
	shares      repo.ShareLinkRepository
	prompts     repo.PromptRepository
	teams       repo.TeamRepository
	frontendURL string
}

func NewService(
	shares repo.ShareLinkRepository,
	prompts repo.PromptRepository,
	teams repo.TeamRepository,
	frontendURL string,
) *Service {
	return &Service{
		shares:      shares,
		prompts:     prompts,
		teams:       teams,
		frontendURL: frontendURL,
	}
}

// CreateOrGet creates a new share link or returns the existing active one (idempotent).
func (s *Service) CreateOrGet(ctx context.Context, promptID, userID uint) (*ShareLinkInfo, bool, error) {
	prompt, err := s.prompts.GetByID(ctx, promptID)
	if err != nil {
		return nil, false, s.mapPromptErr(err)
	}

	if err := s.requireOwnerOrEditor(ctx, prompt, userID); err != nil {
		return nil, false, err
	}

	// Return existing active link if present.
	existing, err := s.shares.GetActiveByPromptID(ctx, promptID)
	if err == nil {
		return s.toInfo(existing), false, nil
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return nil, false, err
	}

	token, err := generateToken()
	if err != nil {
		return nil, false, err
	}

	link := &models.ShareLink{
		PromptID: promptID,
		UserID:   userID,
		Token:    token,
		IsActive: true,
	}
	if err := s.shares.Create(ctx, link); err != nil {
		// Race condition: another request created a link between our check and insert.
		// The partial unique index rejects the duplicate — retry the lookup.
		if existing, retryErr := s.shares.GetActiveByPromptID(ctx, promptID); retryErr == nil {
			return s.toInfo(existing), false, nil
		}
		return nil, false, err
	}

	return s.toInfo(link), true, nil
}

// GetByPromptID returns the active share link for a prompt (management UI).
func (s *Service) GetByPromptID(ctx context.Context, promptID, userID uint) (*ShareLinkInfo, error) {
	prompt, err := s.prompts.GetByID(ctx, promptID)
	if err != nil {
		return nil, s.mapPromptErr(err)
	}

	if err := s.requireOwnerOrMember(ctx, prompt, userID); err != nil {
		return nil, err
	}

	link, err := s.shares.GetActiveByPromptID(ctx, promptID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return s.toInfo(link), nil
}

// Deactivate disables the active share link for a prompt.
func (s *Service) Deactivate(ctx context.Context, promptID, userID uint) error {
	prompt, err := s.prompts.GetByID(ctx, promptID)
	if err != nil {
		return s.mapPromptErr(err)
	}

	if err := s.requireOwnerOrEditor(ctx, prompt, userID); err != nil {
		return err
	}

	if err := s.shares.Deactivate(ctx, promptID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// GetPublicPrompt returns a sanitized prompt for public viewing (no auth required).
func (s *Service) GetPublicPrompt(ctx context.Context, token string) (*PublicPromptInfo, error) {
	link, err := s.shares.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Soft-deleted prompt: GORM preload returns zero-value struct.
	if link.Prompt.ID == 0 {
		return nil, ErrNotFound
	}

	// Async view count increment (best-effort, same pattern as apikey.UpdateLastUsed).
	go func(id uint) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("share.view_count.panic", "error", r)
			}
		}()
		bgCtx, cancel := context.WithTimeout(context.Background(), viewCountTimeout)
		defer cancel()
		if err := s.shares.IncrementViewCount(bgCtx, id); err != nil {
			slog.Error("share.view_count.failed", "id", id, "error", err)
		}
	}(link.ID)

	p := &link.Prompt
	tags := make([]PublicTag, len(p.Tags))
	for i, t := range p.Tags {
		tags[i] = PublicTag{Name: t.Name, Color: t.Color}
	}

	return &PublicPromptInfo{
		Title:   p.Title,
		Content: p.Content,
		Model:   p.Model,
		Tags:    tags,
		Author: PublicAuthor{
			Name:      p.User.Name,
			AvatarURL: p.User.AvatarURL,
		},
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}, nil
}

// --- helpers ---

func (s *Service) requireOwnerOrEditor(ctx context.Context, prompt *models.Prompt, userID uint) error {
	if prompt.TeamID == nil {
		if prompt.UserID != userID {
			return ErrForbidden
		}
		return nil
	}
	if err := teamcheck.RequireEditor(ctx, s.teams, prompt.TeamID, userID); err != nil {
		return s.mapTeamErr(err)
	}
	return nil
}

func (s *Service) requireOwnerOrMember(ctx context.Context, prompt *models.Prompt, userID uint) error {
	if prompt.TeamID == nil {
		if prompt.UserID != userID {
			return ErrForbidden
		}
		return nil
	}
	_, err := s.teams.GetMember(ctx, *prompt.TeamID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrForbidden
		}
		return err
	}
	return nil
}

func (s *Service) mapPromptErr(err error) error {
	if errors.Is(err, repo.ErrNotFound) {
		return ErrPromptNotFound
	}
	return err
}

func (s *Service) mapTeamErr(err error) error {
	if errors.Is(err, teamcheck.ErrForbidden) {
		return ErrForbidden
	}
	if errors.Is(err, teamcheck.ErrViewerReadOnly) {
		return ErrViewerReadOnly
	}
	return err
}

func (s *Service) toInfo(link *models.ShareLink) *ShareLinkInfo {
	return &ShareLinkInfo{
		ID:           link.ID,
		Token:        link.Token,
		URL:          s.frontendURL + "/s/" + link.Token,
		IsActive:     link.IsActive,
		ViewCount:    link.ViewCount,
		LastViewedAt: link.LastViewedAt,
		CreatedAt:    link.CreatedAt,
	}
}

func generateToken() (string, error) {
	b := make([]byte, tokenRandBytes)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return tokenPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}
