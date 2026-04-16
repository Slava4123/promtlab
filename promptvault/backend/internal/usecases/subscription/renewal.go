package subscription

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/payment"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// Retry-политика автопродления. Три попытки списания с интервалом 24ч — даёт юзеру
// время обновить карту или пополнить баланс. После maxRenewalAttempts подписка
// остаётся в past_due до истечения current_period_end, затем expirationLoop переводит
// в expired.
const (
	maxRenewalAttempts = 3
	renewalRetryDelay  = 24 * time.Hour
)

// RenewalNotifier — абстракция для уведомлений юзера о событиях автопродления.
// Email нельзя требовать обязательным (юзер мог зарегистрироваться через OAuth
// без верификации email) — поэтому интерфейс, а реализация проверяет доступность.
type RenewalNotifier interface {
	// NotifyRenewalFailed — отправляется на каждую из N попыток списания.
	// attempt — номер попытки (1..maxRenewalAttempts), graceUntil — опциональный
	// дедлайн grace period после последней неудачи (nil если grace ещё не применён).
	NotifyRenewalFailed(to, planName string, attempt, maxAttempts int, endsAt time.Time, graceUntil *time.Time) error
}

// RenewalLoop пытается продлить подписки за `lookahead` до окончания периода
// и повторяет списание для past_due подписок с rate limiting.
// Логика для каждой подписки: Init+Charge через T-Bank. При успехе webhook продлит
// период через ExtendPeriod. При фейле — RecordRenewalFailure (past_due + attempts++).
type RenewalLoop struct {
	subs      repo.SubscriptionRepository
	plans     repo.PlanRepository
	pays      repo.PaymentRepository
	users     repo.UserRepository
	payment   payment.PaymentProvider
	notifier  RenewalNotifier
	cfg       *config.PaymentConfig
	interval  time.Duration
	lookahead time.Duration
	stopCh    chan struct{}
}

// NewRenewalLoop создаёт цикл автопродления. interval — частота проверки
// (рекомендуется 1 час), lookahead — за сколько до конца периода пытаться
// продлить (рекомендуется 48 часов: даёт буфер на retry).
// notifier может быть nil — тогда уведомления не отправляются (dev-режим).
func NewRenewalLoop(
	subs repo.SubscriptionRepository,
	plans repo.PlanRepository,
	pays repo.PaymentRepository,
	users repo.UserRepository,
	pay payment.PaymentProvider,
	notifier RenewalNotifier,
	cfg *config.PaymentConfig,
	interval, lookahead time.Duration,
) *RenewalLoop {
	return &RenewalLoop{
		subs: subs, plans: plans, pays: pays, users: users,
		payment: pay, notifier: notifier, cfg: cfg,
		interval: interval, lookahead: lookahead,
		stopCh: make(chan struct{}),
	}
}

func (l *RenewalLoop) Start() {
	if l.payment == nil || l.cfg == nil || !l.cfg.Enabled {
		slog.Info("subscription.renewal.disabled", "reason", "payment provider not configured")
		return
	}
	go l.run()
}

func (l *RenewalLoop) Stop() { close(l.stopCh) }

func (l *RenewalLoop) run() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()
	l.tick()
	for {
		select {
		case <-ticker.C:
			l.tick()
		case <-l.stopCh:
			return
		}
	}
}

func (l *RenewalLoop) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	now := time.Now()
	deadline := now.Add(l.lookahead)
	retryAfter := now.Add(-renewalRetryDelay)

	subs, err := l.subs.ListReadyForRenewal(ctx, deadline, retryAfter, maxRenewalAttempts)
	if err != nil {
		slog.Error("subscription.renewal.list_failed", "error", err)
		return
	}

	for _, sub := range subs {
		if sub.CancelAtPeriodEnd {
			// Юзер отменил — не продлеваем.
			continue
		}
		if err := l.renewOne(ctx, &sub); err != nil {
			slog.Error("subscription.renewal.failed",
				"sub_id", sub.ID, "user_id", sub.UserID, "attempts", sub.RenewalAttempts, "error", err)
			continue
		}
	}
}

// renewOne инициирует рекуррентное списание для одной подписки.
// При успехе T-Bank ответит CONFIRMED через webhook — ExtendPeriod продлит период
// и сбросит attempts. При ошибке Init/Charge — RecordRenewalFailure (past_due, attempts++)
// и email юзеру (только на первой неудаче, чтобы не спамить при retry).
func (l *RenewalLoop) renewOne(ctx context.Context, sub *models.Subscription) error {
	plan, err := l.plans.GetByID(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("get plan: %w", err)
	}

	idemKey, err := generateRenewalKey()
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	orderID := fmt.Sprintf("renew_%d_%s", sub.UserID, idemKey[:12])

	providerData, _ := json.Marshal(PaymentProviderData{PlanID: plan.ID, Renewal: "true"})
	pay := &models.Payment{
		UserID:         sub.UserID,
		SubscriptionID: &sub.ID,
		ExternalID:     "pending_" + idemKey,
		IdempotencyKey: idemKey,
		AmountKop:      plan.PriceKop,
		Currency:       "RUB",
		Status:         models.PaymentPending,
		Provider:       "tbank",
		ProviderData:   providerData,
	}
	if err := l.pays.Create(ctx, pay); err != nil {
		return fmt.Errorf("create payment: %w", err)
	}

	// Init без Recurrent (рекуррент уже подключен) — нужен PaymentId для Charge.
	initResult, err := l.payment.Init(ctx, payment.InitRequest{
		OrderID:     orderID,
		Amount:      plan.PriceKop,
		Description: fmt.Sprintf("Автопродление подписки %s", plan.Name),
		WebhookURL:  l.cfg.WebhookBaseURL + "/api/webhooks/tbank",
		CustomerKey: fmt.Sprintf("%d", sub.UserID),
	})
	if err != nil {
		if statusErr := l.pays.UpdateStatus(ctx, pay.ID, models.PaymentFailed); statusErr != nil {
			// Zombie-платёж: status остался pending, но Init фэйлился — reconcile потребуется вручную.
			slog.Error("subscription.renewal.update_status_failed_after_init",
				"payment_id", pay.ID, "user_id", sub.UserID, "error", statusErr)
		}
		l.handleFailure(ctx, sub, plan, "init failed")
		return fmt.Errorf("init: %w", err)
	}
	if err := l.pays.UpdateExternalID(ctx, pay.ID, initResult.ExternalID); err != nil {
		return fmt.Errorf("update external id: %w", err)
	}

	// Charge — собственно безакцептное списание.
	chargeResult, err := l.payment.Charge(ctx, payment.ChargeRequest{
		PaymentID: initResult.ExternalID,
		RebillID:  sub.RebillId,
	})
	if err != nil {
		if statusErr := l.pays.UpdateStatus(ctx, pay.ID, models.PaymentFailed); statusErr != nil {
			// Важно для audit — ExternalID уже ушёл в T-Bank, повторный Charge на него невозможен.
			slog.Error("subscription.renewal.update_status_failed_after_charge",
				"payment_id", pay.ID, "external_id", initResult.ExternalID,
				"user_id", sub.UserID, "error", statusErr)
		}
		l.handleFailure(ctx, sub, plan, "charge failed")
		return fmt.Errorf("charge: %w", err)
	}

	slog.Info("subscription.renewal.charge_sent",
		"sub_id", sub.ID, "user_id", sub.UserID, "plan_id", plan.ID,
		"external_id", chargeResult.ExternalID, "status", chargeResult.Status,
		"attempt", sub.RenewalAttempts+1)

	// Финальная активация и продление — в HandleWebhook (там же extractPlanID
	// прочитает plan_id из ProviderData и продлит подписку, сбросив attempts).
	return nil
}

// GracePeriod — насколько дольше current_period_end продлеваем доступ,
// если все retries провалились. Даёт юзеру шанс обновить карту/вернуть средства
// и не прерывает его workflow внезапно (M-9).
const GracePeriod = 7 * 24 * time.Hour

// handleFailure — общая обработка неудачи Init/Charge: фиксирует попытку в БД
// и шлёт email на каждую из maxRenewalAttempts попыток (разный текст по attempt),
// чтобы юзер знал о прогрессе retry, а не узнавал о проблеме только после downgrade.
func (l *RenewalLoop) handleFailure(ctx context.Context, sub *models.Subscription, plan *models.SubscriptionPlan, reason string) {
	if err := l.subs.RecordRenewalFailure(ctx, sub.ID); err != nil {
		slog.Error("subscription.renewal.record_failure_failed",
			"sub_id", sub.ID, "user_id", sub.UserID, "error", err)
		return
	}

	attempt := sub.RenewalAttempts + 1 // после RecordRenewalFailure это текущая попытка
	slog.Warn("subscription.renewal.failure_recorded",
		"sub_id", sub.ID, "user_id", sub.UserID, "reason", reason,
		"attempt", attempt, "max_attempts", maxRenewalAttempts,
		"period_end", sub.CurrentPeriodEnd)

	if l.notifier == nil {
		return
	}
	user, err := l.users.GetByID(ctx, sub.UserID)
	if err != nil || user == nil || user.Email == "" {
		// Email недоступен (OAuth без email, или user удалён) — не фатально.
		if err != nil && !errors.Is(err, repo.ErrNotFound) {
			slog.Warn("subscription.renewal.notify_user_fetch_failed",
				"sub_id", sub.ID, "user_id", sub.UserID, "error", err)
		}
		return
	}
	// graceUntil: nil для первых N-1 попыток. На последней — показываем юзеру
	// конкретный дедлайн обновить карту, до которого доступ ещё сохранён.
	var graceUntil *time.Time
	if attempt >= maxRenewalAttempts {
		t := sub.CurrentPeriodEnd.Add(GracePeriod)
		graceUntil = &t
	}
	if err := l.notifier.NotifyRenewalFailed(user.Email, plan.Name, attempt, maxRenewalAttempts, sub.CurrentPeriodEnd, graceUntil); err != nil {
		slog.Warn("subscription.renewal.notify_email_failed",
			"sub_id", sub.ID, "user_id", sub.UserID, "error", err)
	}
}

func generateRenewalKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
