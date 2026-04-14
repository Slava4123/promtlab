package subscription

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/payment"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// RenewalLoop пытается продлить подписки за `lookahead` дней до окончания периода.
// Логика: для каждой active подписки с auto_renew=true и непустым rebill_id —
// сделать Init+Charge через T-Bank. При успехе webhook сам продлит период.
//
// Здесь только инициирование Charge; финальная активация — в HandleWebhook.
type RenewalLoop struct {
	subs      repo.SubscriptionRepository
	plans     repo.PlanRepository
	pays      repo.PaymentRepository
	payment   payment.PaymentProvider
	cfg       *config.PaymentConfig
	interval  time.Duration
	lookahead time.Duration
	stopCh    chan struct{}
}

// NewRenewalLoop создаёт цикл автопродления. interval — частота проверки
// (рекомендуется 1 час), lookahead — за сколько до конца периода пытаться
// продлить (рекомендуется 24-72 часа: даёт буфер на retry если карта отклонена).
func NewRenewalLoop(
	subs repo.SubscriptionRepository,
	plans repo.PlanRepository,
	pays repo.PaymentRepository,
	pay payment.PaymentProvider,
	cfg *config.PaymentConfig,
	interval, lookahead time.Duration,
) *RenewalLoop {
	return &RenewalLoop{
		subs: subs, plans: plans, pays: pays,
		payment: pay, cfg: cfg,
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

	deadline := time.Now().Add(l.lookahead)
	subs, err := l.subs.ListReadyForRenewal(ctx, deadline)
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
				"sub_id", sub.ID, "user_id", sub.UserID, "error", err)
			continue
		}
	}
}

// renewOne инициирует рекуррентное списание для одной подписки.
// При успехе T-Bank ответит CONFIRMED синхронно или асинхронно через webhook —
// в любом случае webhook обработает активацию следующего периода.
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

	providerData, _ := json.Marshal(map[string]string{"plan_id": plan.ID, "renewal": "true"})
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
		_ = l.pays.UpdateStatus(ctx, pay.ID, models.PaymentFailed)
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
		_ = l.pays.UpdateStatus(ctx, pay.ID, models.PaymentFailed)
		return fmt.Errorf("charge: %w", err)
	}

	slog.Info("subscription.renewal.charge_sent",
		"sub_id", sub.ID, "user_id", sub.UserID, "plan_id", plan.ID,
		"external_id", chargeResult.ExternalID, "status", chargeResult.Status)

	// Финальная активация и продление — в HandleWebhook (там же extractPlanID
	// прочитает plan_id из ProviderData и продлит подписку).
	return nil
}

func generateRenewalKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
