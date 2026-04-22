package share

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"strings"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	activityuc "promptvault/internal/usecases/activity"
	quotauc "promptvault/internal/usecases/quota"
	"promptvault/internal/usecases/subscription"
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
	quotas      *quotauc.Service
	// activity — опциональный team activity feed (Phase 14).
	activity *activityuc.Service
	// viewLogger — опциональный write-path в share_view_log. Пишет только для
	// Pro+ owner'ов (план читается из уже preload'ленного link.Prompt.User).
	// Phase 14, B.2.
	viewLogger repo.AnalyticsRepository
	// brandingProvider — опциональный lookup для branded share pages (Phase 14 D).
	// Возвращает BrandingInfo только если владелец team на Max. nil в остальных случаях.
	brandingProvider BrandingProvider
}

// BrandingProvider — интерфейс подгрузки branding по team_id. Избегает прямой
// зависимости share.Service от team.Service (DIP) и mock-friendly.
type BrandingProvider interface {
	GetBrandingForShare(ctx context.Context, teamID uint) (*models.BrandingInfo, error)
}

// ViewMeta — дополнительный контекст HTTP-запроса для логирования просмотров.
// Передаётся из handler'а (referer/user-agent недоступны в usecase-слое напрямую).
type ViewMeta struct {
	Referer   string
	UserAgent string
}

func NewService(
	shares repo.ShareLinkRepository,
	prompts repo.PromptRepository,
	teams repo.TeamRepository,
	frontendURL string,
	quotas *quotauc.Service,
) *Service {
	return &Service{
		shares:      shares,
		prompts:     prompts,
		teams:       teams,
		frontendURL: frontendURL,
		quotas:      quotas,
	}
}

// SetActivity подключает team_activity_log хуки (Phase 14).
func (s *Service) SetActivity(activity *activityuc.Service) {
	s.activity = activity
}

// SetViewLogger подключает write-path в share_view_log (Phase 14, B.2).
// Nil-safe: если не подключён — LogShareView в GetPublicPrompt no-op'ит.
// План владельца читается из уже preload'ленного link.Prompt.User.PlanID —
// отдельный UserRepository здесь не нужен (H7/M9).
func (s *Service) SetViewLogger(analytics repo.AnalyticsRepository) {
	s.viewLogger = analytics
}

// SetBrandingLookup подключает branded share pages (Phase 14 D).
// Nil-safe: если не подключён — Branding в PublicPromptInfo = nil.
func (s *Service) SetBrandingLookup(p BrandingProvider) {
	s.brandingProvider = p
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

	// Проверка квот (только при создании новой).
	// Phase 14: daily лимит создаваемых ссылок (fixed window UTC-полночь).
	// Unchanged: total active cap (опциональный soft cap, -1 = unlimited).
	if s.quotas != nil {
		if err := s.quotas.CheckDailyShareCreation(ctx, userID); err != nil {
			return nil, false, err
		}
		if err := s.quotas.CheckShareLinkQuota(ctx, userID); err != nil {
			return nil, false, err
		}
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

	// Increment daily counter AFTER successful INSERT. Best-effort: если инкремент
	// fail'ит, это не откатывает создание (юзер получил ссылку, но счётчик
	// недосчитал). Error-level — revenue-leak сигнал для SRE/метрик (M2).
	if s.quotas != nil {
		if err := s.quotas.IncrementShareCreation(ctx, userID); err != nil {
			slog.ErrorContext(ctx, "share.quota.increment_failed", "user_id", userID, "error", err)
		}
	}

	// Activity feed hook (Phase 14) — только для team-промптов.
	if prompt.TeamID != nil {
		s.activity.LogSafe(ctx, activityuc.Event{
			TeamID:      *prompt.TeamID,
			ActorID:     userID,
			EventType:   models.ActivityShareCreated,
			TargetType:  models.TargetShare,
			TargetID:    &link.ID,
			TargetLabel: prompt.Title,
			Metadata:    map[string]any{"prompt_id": prompt.ID, "token": link.Token},
		})
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

	// Activity feed hook.
	if prompt.TeamID != nil {
		s.activity.LogSafe(ctx, activityuc.Event{
			TeamID:      *prompt.TeamID,
			ActorID:     userID,
			EventType:   models.ActivityShareRevoked,
			TargetType:  models.TargetShare,
			TargetLabel: prompt.Title,
			Metadata:    map[string]any{"prompt_id": prompt.ID},
		})
	}
	return nil
}

// GetPublicPrompt returns a sanitized prompt for public viewing (no auth required).
// Phase 14 (B.2): если viewLogger подключён и owner на Pro+, async пишет запись
// в share_view_log через ViewMeta с referer/user-agent из HTTP-запроса.
func (s *Service) GetPublicPrompt(ctx context.Context, token string, meta ViewMeta) (*PublicPromptInfo, error) {
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

	// Async timeline log в share_view_log (Phase 14, B.2). Только для Pro+ owner'ов.
	// План владельца читается из уже preload'ленного link.Prompt.User.PlanID —
	// избегаем лишнего users.GetByID на hot-path /s/:token (M9).
	if s.viewLogger != nil {
		go s.logShareView(link.ID, link.Prompt.User.PlanID, meta)
	}

	p := &link.Prompt
	tags := make([]PublicTag, len(p.Tags))
	for i, t := range p.Tags {
		tags[i] = PublicTag{Name: t.Name, Color: t.Color}
	}

	info := &PublicPromptInfo{
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
	}

	// Phase 14 D: branded share pages — если промпт в команде и provider подключён.
	if p.TeamID != nil && s.brandingProvider != nil {
		if branding, berr := s.brandingProvider.GetBrandingForShare(ctx, *p.TeamID); berr == nil && branding != nil && !branding.IsEmpty() {
			info.Branding = branding
		}
	}
	return info, nil
}

// logShareView — goroutine, пишет в share_view_log если владелец на Pro+.
// Best-effort: любая ошибка → slog.Warn, не блокирует основную операцию.
// План владельца приходит уже resolved из preload'ленного User (M9) —
// это избегает лишний users.GetByID на hot-path /s/:token.
func (s *Service) logShareView(shareLinkID uint, ownerPlanID string, meta ViewMeta) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("share.view_log.panic", "error", r)
		}
	}()
	// Free не пишет в timeline — фича Pro+.
	if !subscription.IsPaid(ownerPlanID) {
		return
	}
	bgCtx, cancel := context.WithTimeout(context.Background(), viewCountTimeout)
	defer cancel()

	view := &models.ShareView{
		ShareLinkID:     shareLinkID,
		Referer:         truncateString(meta.Referer, 500),
		UserAgentFamily: uaFamily(meta.UserAgent),
	}
	if err := s.viewLogger.LogShareView(bgCtx, view); err != nil {
		slog.Warn("share.view_log.insert_failed", "share_link_id", shareLinkID, "error", err)
	}
}

// truncateString — безопасно обрезает до max байт (без выхода за границы рун).
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Обрезаем по байтам — для DB VARCHAR достаточно.
	return s[:maxLen]
}

// uaFamily — возвращает короткий идентификатор браузера (Chrome/Safari/Firefox/Edge/Other).
// Простая эвристика без зависимости от user-agent-parser — для dashboard-метрик достаточно.
func uaFamily(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "edg/"):
		return "Edge"
	case strings.Contains(ua, "chrome/"):
		return "Chrome"
	case strings.Contains(ua, "firefox/"):
		return "Firefox"
	case strings.Contains(ua, "safari/"):
		return "Safari"
	case ua == "":
		return ""
	default:
		return "Other"
	}
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
