// backend/internal/usecases/referral/reward.go
package referral

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/metrics"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// Service — реализация реферальной выдачи награды (M-7, Phase 3 Pricing v3).
//
// Точка входа — GrantReward, вызывается из ReferralRewardLoop через 14 дней
// после первого платежа реферри (eligibility-окно > T-Bank refund window).
//
// Архитектура: один атомарный CAS в users.MarkReferralRewarded защищает от
// двойной выдачи. Если CAS не прошёл (race / dup webhook) — возвращаем
// ErrAlreadyRewarded и НЕ undo'им уже сделанное продление подписки. Это
// безопасно: продление само по себе не вредит юзеру, "переплата" в 30 дней
// исключена тем что MarkReferralRewarded — global gate, а не per-payment.
type Service struct {
	subs    repo.SubscriptionRepository
	users   repo.UserRepository
	pays    repo.PaymentRepository
	pending repo.ReferralRewardRepository
	nowFn   func() time.Time
}

// NewService — конструктор. pending передаётся для совместимости с Phase 3+
// (ReferralRewardLoop будет вызывать pending.ListEligible/Delete), в GrantReward
// сейчас не используется.
func NewService(
	subs repo.SubscriptionRepository,
	users repo.UserRepository,
	pays repo.PaymentRepository,
	pending repo.ReferralRewardRepository,
) *Service {
	return &Service{
		subs:    subs,
		users:   users,
		pays:    pays,
		pending: pending,
		nowFn:   time.Now,
	}
}

// SetNowFn — for tests. Не делать публично в production-коде; здесь — намеренно,
// чтобы integration test'ы и unit'ы могли заморозить время.
func (s *Service) SetNowFn(fn func() time.Time) { s.nowFn = fn }

// GrantReward выдаёт +RewardDays Pro пригласившему. См. spec в плане Task 14.
//
// Препроверки:
//  1. users.GetByID(referrerID) → ErrReferrerMissing если nil.
//  2. referrer.ReferralRewardedAt != nil → ErrAlreadyRewarded (быстрый путь).
//  3. payment.Status != succeeded → ErrPaymentRefunded.
//
// Затем по plan_id референта:
//   - pro / pro_yearly / max / max_yearly: extend current_period_end на 30d.
//   - free и прочие: создать synthetic active Subscription{pro, auto_renew=false}
//     и SetPlan(referrer, "pro") — это "trial" Pro, истечёт через 30d.
//
// Атомарный CAS MarkReferralRewarded в конце — если returns (false, nil),
// другой webhook опередил нас → ErrAlreadyRewarded.
func (s *Service) GrantReward(ctx context.Context, referrerID, refereeID, paymentID uint) error {
	referrer, err := s.users.GetByID(ctx, referrerID)
	if err != nil {
		// repo.ErrNotFound → ErrReferrerMissing; иначе wrap'аем как infra error.
		if isNotFound(err) {
			return ErrReferrerMissing
		}
		return fmt.Errorf("referral: get referrer %d: %w", referrerID, err)
	}
	if referrer == nil {
		return ErrReferrerMissing
	}
	if referrer.ReferralRewardedAt != nil {
		return ErrAlreadyRewarded
	}

	payment, err := s.pays.GetByID(ctx, paymentID)
	if err != nil {
		if isNotFound(err) {
			return fmt.Errorf("referral: payment %d: %w", paymentID, err)
		}
		return fmt.Errorf("referral: get payment %d: %w", paymentID, err)
	}
	if payment == nil {
		return ErrPaymentRefunded
	}
	if payment.Status != models.PaymentSucceeded {
		// Refunded / failed / pending — все не дают права на reward.
		return ErrPaymentRefunded
	}

	now := s.nowFn()
	rewardDuration := time.Duration(RewardDays) * 24 * time.Hour

	isTrial := false
	switch referrer.PlanID {
	case "pro", "pro_yearly", "max", "max_yearly":
		if err := s.extendActiveSubscription(ctx, referrerID, rewardDuration); err != nil {
			return fmt.Errorf("referral: extend subscription for %d: %w", referrerID, err)
		}
	default:
		// Free и любой другой нераспознанный план — выдаём 30-дневный Pro trial.
		isTrial = true
		if err := s.createTrialPro(ctx, referrerID, now, rewardDuration); err != nil {
			return fmt.Errorf("referral: create trial pro for %d: %w", referrerID, err)
		}
	}

	// Атомарный CAS: только после успешного grant'а помечаем юзера как награждённого.
	// Это защищает от двойной выдачи при race с параллельным webhook'ом.
	ok, err := s.users.MarkReferralRewarded(ctx, referrerID)
	if err != nil {
		return fmt.Errorf("referral: mark rewarded for %d: %w", referrerID, err)
	}
	if !ok {
		slog.WarnContext(ctx, "referral.reward.race",
			"referrer_id", referrerID,
			"referee_id", refereeID,
			"payment_id", paymentID,
		)
		return ErrAlreadyRewarded
	}

	metrics.ReferralRewardsGrantedTotal.WithLabelValues(NormalizePlanLabel(referrer.PlanID)).Inc()

	slog.InfoContext(ctx, "referral.reward.granted",
		"referrer_id", referrerID,
		"referee_id", refereeID,
		"payment_id", paymentID,
		"from_plan", referrer.PlanID,
		"reward_days", RewardDays,
		"is_trial", isTrial,
	)

	return nil
}

// NormalizePlanLabel маппит plan_id (pro_yearly/max_yearly) на metric label (pro/max).
// Free и неизвестные значения → "free". Используется в Prometheus label'ах для
// `referral_rewards_granted_total{referrer_plan}` и `analytics_insights_gated_total{plan}`,
// чтобы yearly/monthly не давали отдельные series (cardinality contained).
//
// Экспортирован, чтобы analytics package мог переиспользовать без дублирования.
func NormalizePlanLabel(planID string) string {
	switch planID {
	case "pro", "pro_yearly":
		return "pro"
	case "max", "max_yearly":
		return "max"
	default:
		return "free"
	}
}

// extendActiveSubscription продлевает текущую active/past_due подписку на duration.
// Использует UpdatePeriodEnd (а не ExtendPeriod), потому что:
//   - не сбрасываем renewal_attempts/pre_expire_stage (reward — это бонус, не renewal);
//   - не переводим past_due → active (это решение биллинга после успешного списания);
//   - не меняем current_period_start.
func (s *Service) extendActiveSubscription(ctx context.Context, userID uint, duration time.Duration) error {
	sub, err := s.subs.GetActiveByUserID(ctx, userID)
	if err != nil {
		if isNotFound(err) {
			return fmt.Errorf("no active subscription for user %d (cannot extend)", userID)
		}
		return fmt.Errorf("get active subscription for user %d: %w", userID, err)
	}
	if sub == nil {
		return fmt.Errorf("no active subscription for user %d (cannot extend)", userID)
	}
	return s.subs.UpdatePeriodEnd(ctx, sub.ID, sub.CurrentPeriodEnd.Add(duration))
}

// createTrialPro создаёт synthetic Pro-подписку для free-юзера и переводит
// users.plan_id в "pro". AutoRenew=false / RebillId="" — это trial, по
// истечении 30 дней expirationLoop downgrade'нёт юзера обратно в free
// (стандартный механизм SubStatusExpired → SetPlan free).
//
// Намеренно НЕ используем models.NewSubscription — он ставит AutoRenew=true,
// для trial это неверно.
func (s *Service) createTrialPro(ctx context.Context, userID uint, now time.Time, duration time.Duration) error {
	sub := &models.Subscription{
		UserID:             userID,
		PlanID:             "pro",
		Status:             models.SubStatusActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.Add(duration),
		RebillId:           "",
		AutoRenew:          false,
	}
	if err := s.subs.Create(ctx, sub); err != nil {
		return fmt.Errorf("create trial subscription: %w", err)
	}
	if err := s.users.SetPlan(ctx, userID, "pro"); err != nil {
		return fmt.Errorf("set plan pro: %w", err)
	}
	return nil
}

// isNotFound — typed check для repo.ErrNotFound. Локальная утилита,
// не экспортируем — использовать errors.Is напрямую где это удобнее.
func isNotFound(err error) bool {
	return errors.Is(err, repo.ErrNotFound)
}
