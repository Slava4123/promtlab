package email

import (
	"context"
	"log/slog"
	"time"

	repo "promptvault/internal/interface/repository"
)

// EmailInsightsNotifier — SMTP-реализация analytics.InsightsNotifier.
// Шлёт краткий digest владельцу insight, если:
//  1. users.insight_emails_enabled = true (opt-in ФЗ-152);
//  2. за последние 7 дней такого типа инсайта юзеру ещё не слали.
//
// Вызывается из analytics.Service.upsertSafe после успешного UpsertInsight.
// Fail-silent: ошибки не возвращаются (нет точки возврата из OnInsightsChanged),
// только логируются.
type EmailInsightsNotifier struct {
	email        *Service
	users        repo.UserRepository
	notifs       repo.InsightNotificationRepository
	frontendURL  string
	rateLimit    time.Duration // по умолчанию 7d
}

// Rate-limit один раз в неделю на пару (user, insight_type).
// Меняется только через NewEmailInsightsNotifier (для тестов).
const defaultInsightsRateLimit = 7 * 24 * time.Hour

// NewEmailInsightsNotifier создаёт реализацию. frontendURL — для CTA-ссылки
// в письме на страницу /analytics/insights.
func NewEmailInsightsNotifier(email *Service, users repo.UserRepository, notifs repo.InsightNotificationRepository, frontendURL string) *EmailInsightsNotifier {
	return &EmailInsightsNotifier{
		email:       email,
		users:       users,
		notifs:      notifs,
		frontendURL: frontendURL,
		rateLimit:   defaultInsightsRateLimit,
	}
}

// OnInsightsChanged — hook, который analytics.Service вызывает после upsert.
// Все проверки и side-effects внутри: smtp-конфиг, opt-in, rate-limit.
func (n *EmailInsightsNotifier) OnInsightsChanged(ctx context.Context, userID uint, teamID *uint, insightType string, _ any) {
	if n.email == nil || !n.email.Configured() {
		return
	}
	// Team-scope инсайты не шлём на email — юзер сам зайдёт в team analytics.
	if teamID != nil {
		return
	}

	recent, err := n.notifs.RecentlySent(ctx, userID, insightType, n.rateLimit)
	if err != nil {
		slog.WarnContext(ctx, "insights_notifier.recently_sent_failed", "err", err, "user_id", userID, "type", insightType)
		return
	}
	if recent {
		return
	}

	user, err := n.users.GetByID(ctx, userID)
	if err != nil {
		slog.WarnContext(ctx, "insights_notifier.user_lookup_failed", "err", err, "user_id", userID)
		return
	}
	if !user.InsightEmailsEnabled || !user.EmailVerified || user.Email == "" {
		return
	}

	if err := n.email.SendInsightsDigest(user.Email, user.Name, insightType, n.frontendURL); err != nil {
		slog.WarnContext(ctx, "insights_notifier.send_failed", "err", err, "user_id", userID, "type", insightType)
		return
	}
	if err := n.notifs.Record(ctx, userID, insightType); err != nil {
		slog.WarnContext(ctx, "insights_notifier.record_failed", "err", err, "user_id", userID, "type", insightType)
	}
}
