package analytics

import "context"

// InsightsNotifier — интерфейс пост-compute-hook для внешних каналов
// (email, push, webhook). ComputeInsights вызывает OnInsightsChanged
// после успешного upsert каждого типа инсайта.
//
// Реализации:
//   - NoopNotifier — default, ничего не делает. Тесты и офлайн прогон.
//   - EmailInsightsNotifier — (infrastructure/email) SMTP-рассылка
//     с rate-limit 1 письмо/неделю на пару (user, insightType), ФЗ-152
//     compliant opt-in через users.insight_emails_enabled.
type InsightsNotifier interface {
	OnInsightsChanged(ctx context.Context, userID uint, teamID *uint, insightType string, payload any)
}

// NoopNotifier — zero-overhead noop для prod с выключенным feature-flag.
type NoopNotifier struct{}

func (NoopNotifier) OnInsightsChanged(context.Context, uint, *uint, string, any) {}
