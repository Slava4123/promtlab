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

// Service — бизнес-логика подписок: оформление, отмена, обработка webhook.
type Service struct {
	subs    repo.SubscriptionRepository
	plans   repo.PlanRepository
	pays    repo.PaymentRepository
	users   repo.UserRepository
	payment payment.PaymentProvider // может быть nil, если оплата не настроена
	cfg     *config.PaymentConfig
}

// NewService создаёт сервис подписок. payment может быть nil — в этом случае
// Checkout вернёт ErrPaymentNotConfigured.
func NewService(
	subs repo.SubscriptionRepository,
	plans repo.PlanRepository,
	pays repo.PaymentRepository,
	users repo.UserRepository,
	payment payment.PaymentProvider,
	cfg *config.PaymentConfig,
) *Service {
	return &Service{
		subs:    subs,
		plans:   plans,
		pays:    pays,
		users:   users,
		payment: payment,
		cfg:     cfg,
	}
}

// GetPlans возвращает список активных тарифных планов.
func (s *Service) GetPlans(ctx context.Context) ([]models.SubscriptionPlan, error) {
	return s.plans.GetActive(ctx)
}

// GetSubscription возвращает активную подписку пользователя.
// Если подписки нет — возвращает (nil, nil).
func (s *Service) GetSubscription(ctx context.Context, userID uint) (*models.Subscription, error) {
	sub, err := s.subs.GetActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return sub, nil
}

// Checkout инициализирует платёж для оформления подписки.
//
// Flow двухфазного сохранения:
//  1. Создаём Payment в БД со статусом pending и placeholder external_id
//     (до вызова T-Bank Init). Это защищает от orphan-платежей — если БД упадёт
//     после Init, у T-Bank будет платёж, а у нас нет, и webhook бы не нашёл запись.
//  2. Вызываем T-Bank Init — получаем реальный PaymentID.
//  3. UpdateExternalID перезаписывает placeholder. При фейле Init помечаем failed.
func (s *Service) Checkout(ctx context.Context, in CheckoutInput) (*CheckoutResult, error) {
	// Проверяем существование плана
	plan, err := s.plans.GetByID(ctx, in.PlanID)
	if err != nil {
		slog.Warn("subscription.checkout.plan_not_found", "plan_id", in.PlanID)
		return nil, ErrPlanNotFound
	}

	// Если уже на том же плане — отклоняем. Смена плана разрешена:
	// старая подписка отменяется ТОЛЬКО после успешной оплаты (в HandleWebhook).
	existing, err := s.subs.GetActiveByUserID(ctx, in.UserID)
	if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return nil, fmt.Errorf("subscription.checkout: проверка подписки: %w", err)
	}
	if existing != nil && existing.PlanID == in.PlanID {
		return nil, ErrAlreadySubscribed
	}

	// Проверяем что платёжная система настроена
	if s.payment == nil || s.cfg == nil || !s.cfg.Enabled {
		return nil, ErrPaymentNotConfigured
	}

	// Генерируем idempotency key
	idemKey, err := generateIdempotencyKey()
	if err != nil {
		return nil, fmt.Errorf("subscription.checkout: не удалось сгенерировать ключ: %w", err)
	}

	orderID := fmt.Sprintf("sub_%d_%s", in.UserID, idemKey[:12])

	// Phase 1: сохраняем Payment ДО Init — plan_id в ProviderData гарантирует
	// активацию правильного плана, даже если у двух планов совпадают цены.
	providerData, err := json.Marshal(PaymentProviderData{PlanID: plan.ID})
	if err != nil {
		return nil, fmt.Errorf("subscription.checkout: marshal provider_data: %w", err)
	}
	pay := &models.Payment{
		UserID:         in.UserID,
		ExternalID:     "pending_" + idemKey, // placeholder — обновится после Init
		IdempotencyKey: idemKey,
		AmountKop:      plan.PriceKop,
		Currency:       "RUB",
		Status:         models.PaymentPending,
		Provider:       "tbank",
		ProviderData:   providerData,
	}
	if err := s.pays.Create(ctx, pay); err != nil {
		slog.Error("subscription.checkout.save_payment_failed", "user_id", in.UserID, "error", err)
		return nil, fmt.Errorf("subscription.checkout: не удалось сохранить платёж: %w", err)
	}

	// Получаем email юзера для фискального чека (54-ФЗ требует Email/Phone).
	user, userErr := s.users.GetByID(ctx, in.UserID)
	if userErr != nil {
		return nil, fmt.Errorf("subscription.checkout: get user: %w", userErr)
	}

	// Phase 2: Init у T-Bank.
	// Recurrent=true + CustomerKey запускают рекуррент: после первого успешного
	// платежа T-Bank вернёт RebillId в webhook (поле "RebillId"), который
	// мы сохраняем в Subscription для последующих автопродлений через /Charge.
	// RecurrentEnabled=false отключает рекуррент (только для теста 1 T-Bank).
	useRecurrent := s.cfg.RecurrentEnabled
	customerKey := ""
	if useRecurrent {
		customerKey = fmt.Sprintf("%d", in.UserID)
	}
	initResult, err := s.payment.Init(ctx, payment.InitRequest{
		OrderID:     orderID,
		Amount:      plan.PriceKop,
		Description: fmt.Sprintf("Подписка %s", plan.Name),
		SuccessURL:  s.cfg.SuccessURL,
		FailURL:     s.cfg.FailURL,
		WebhookURL:  s.cfg.WebhookBaseURL + "/api/webhooks/tbank",
		Recurrent:   useRecurrent,
		CustomerKey: customerKey,
		Receipt:     buildReceipt(s.cfg, user.Email, plan),
	})
	if err != nil {
		// Помечаем как failed чтобы не висел в pending.
		if upErr := s.pays.UpdateStatus(ctx, pay.ID, models.PaymentFailed); upErr != nil {
			slog.Error("subscription.checkout.mark_failed", "payment_id", pay.ID, "error", upErr)
		}
		slog.Error("subscription.checkout.init_failed", "user_id", in.UserID, "error", err)
		return nil, ErrPaymentFailed
	}

	// Phase 3: записываем реальный PaymentID T-Bank — webhook'и будут искать по нему.
	if err := s.pays.UpdateExternalID(ctx, pay.ID, initResult.ExternalID); err != nil {
		slog.Error("subscription.checkout.update_external_id", "payment_id", pay.ID, "error", err)
		return nil, fmt.Errorf("subscription.checkout: не удалось обновить external_id: %w", err)
	}

	slog.Info("subscription.checkout.created",
		"user_id", in.UserID,
		"plan_id", in.PlanID,
		"order_id", orderID,
		"external_id", initResult.ExternalID,
	)

	return &CheckoutResult{PaymentURL: initResult.PaymentURL}, nil
}

// Downgrade немедленно переводит юзера на Free план, отменяя активную подписку.
func (s *Service) Downgrade(ctx context.Context, userID uint) error {
	sub, err := s.subs.GetActiveByUserID(ctx, userID)
	if err == nil && sub != nil {
		if expErr := s.subs.ExpireAndDowngrade(ctx, sub.ID, userID); expErr != nil {
			slog.Error("subscription.downgrade.failed", "user_id", userID, "error", expErr)
			return fmt.Errorf("subscription.downgrade: %w", expErr)
		}
		slog.Info("subscription.downgrade.completed", "user_id", userID, "old_plan", sub.PlanID)
	} else {
		// Нет подписки, просто обновляем plan_id
		if updErr := s.users.Update(ctx, &models.User{ID: userID, PlanID: "free"}); updErr != nil {
			return fmt.Errorf("subscription.downgrade: %w", updErr)
		}
	}
	return nil
}

// Cancel помечает активную подписку для отмены в конце текущего периода.
// Если передан Reason — пишет его в subscription_cancellations (M-6b exit survey).
// Ошибка записи Reason не блокирует отмену — логируется и swallow'ится.
func (s *Service) Cancel(ctx context.Context, in CancelInput) error {
	if !IsValidCancelReason(in.Reason) {
		return ErrInvalidCancelReason
	}
	sub, err := s.subs.GetActiveByUserID(ctx, in.UserID)
	if err != nil {
		slog.Warn("subscription.cancel.no_subscription", "user_id", in.UserID)
		return ErrNoActiveSubscription
	}

	if err := s.subs.CancelAtPeriodEnd(ctx, sub.ID); err != nil {
		slog.Error("subscription.cancel.failed", "user_id", in.UserID, "sub_id", sub.ID, "error", err)
		return fmt.Errorf("subscription.cancel: %w", err)
	}

	if in.Reason != "" {
		rec := &models.SubscriptionCancellation{
			UserID:         in.UserID,
			SubscriptionID: sub.ID,
			PlanID:         sub.PlanID,
			Reason:         in.Reason,
			OtherText:      in.Other,
			CancelledAt:    time.Now(),
		}
		if err := s.subs.RecordCancellation(ctx, rec); err != nil {
			// Не фейлим flow — отмена уже зафиксирована, потеря reason для аналитики допустима.
			slog.Error("subscription.cancel.record_reason_failed", "user_id", in.UserID, "error", err)
		}
	}

	slog.Info("subscription.cancel.scheduled",
		"user_id", in.UserID, "sub_id", sub.ID, "period_end", sub.CurrentPeriodEnd, "reason", in.Reason)
	return nil
}

// Pause ставит подписку на паузу на N месяцев (M-6). Работает только для платных
// активных подписок. В период паузы user.PlanID='free' (квоты падают на Free-лимиты),
// при Resume — восстанавливается.
//
// В течение паузы current_period_end НЕ меняется — мы сохраняем момент входа
// (paused_at), чтобы при Resume вычислить remaining и сдвинуть period_end вперёд.
func (s *Service) Pause(ctx context.Context, in PauseInput) error {
	if in.Months < 1 || in.Months > 3 {
		return ErrInvalidPauseMonths
	}
	sub, err := s.subs.GetActiveByUserID(ctx, in.UserID)
	if err != nil {
		return ErrNoActiveSubscription
	}
	if sub.Status == models.SubStatusPaused {
		return ErrSubscriptionPaused
	}
	if sub.Status != models.SubStatusActive {
		return ErrSubscriptionNotPausable
	}
	// Паузить имеет смысл только платный план; для free пауза бессмысленна,
	// а для past_due блокируем до разрешения биллингового состояния.
	plan, err := s.plans.GetByID(ctx, sub.PlanID)
	if err != nil || plan == nil || plan.PriceKop <= 0 {
		return ErrSubscriptionNotPausable
	}

	now := time.Now()
	pausedUntil := now.AddDate(0, in.Months, 0)

	if err := s.subs.Pause(ctx, sub.ID, in.UserID, now, pausedUntil); err != nil {
		slog.Error("subscription.pause.failed", "user_id", in.UserID, "sub_id", sub.ID, "error", err)
		return fmt.Errorf("subscription.pause: %w", err)
	}
	slog.Info("subscription.pause.scheduled",
		"user_id", in.UserID, "sub_id", sub.ID, "months", in.Months, "paused_until", pausedUntil)
	return nil
}

// Resume досрочно возобновляет приостановленную подписку (M-6).
// Новый period_end = now + (old_period_end - paused_at) — юзер не теряет
// оставшиеся дни оплаченного периода.
func (s *Service) Resume(ctx context.Context, userID uint) error {
	sub, err := s.subs.GetActiveByUserID(ctx, userID)
	if err != nil {
		return ErrNoActiveSubscription
	}
	if sub.Status != models.SubStatusPaused {
		return ErrSubscriptionNotPaused
	}
	if sub.PausedAt == nil {
		// Защитная ветка: status=paused без paused_at — corrupted state.
		// Логируем и используем now (лучше чем панить юзера).
		slog.Error("subscription.resume.missing_paused_at", "user_id", userID, "sub_id", sub.ID)
		pausedAt := sub.CurrentPeriodEnd // fallback: 0 remaining
		sub.PausedAt = &pausedAt
	}

	now := time.Now()
	remaining := sub.CurrentPeriodEnd.Sub(*sub.PausedAt)
	if remaining < 0 {
		remaining = 0 // на всякий случай — не продлеваем в прошлое
	}
	newEnd := now.Add(remaining)

	if err := s.subs.Resume(ctx, sub.ID, userID, now, newEnd); err != nil {
		slog.Error("subscription.resume.failed", "user_id", userID, "sub_id", sub.ID, "error", err)
		return fmt.Errorf("subscription.resume: %w", err)
	}
	slog.Info("subscription.resume.completed",
		"user_id", userID, "sub_id", sub.ID, "new_period_end", newEnd)
	return nil
}

// HandleWebhook обрабатывает webhook-уведомление от платёжного провайдера.
// Возвращает ErrInvalidWebhookSignature как sentinel (без обёртки) — handler
// отличает её через errors.Is и отвечает 400, чтобы T-Bank не ретраил.
func (s *Service) HandleWebhook(ctx context.Context, provider string, params map[string]string) error {
	token := params["Token"]

	// Проверяем подпись
	if s.payment == nil {
		return ErrPaymentNotConfigured
	}
	if !s.payment.VerifyWebhookSignature(params, token) {
		slog.Warn("subscription.webhook.invalid_signature", "provider", provider)
		return ErrInvalidWebhookSignature
	}

	externalID := params["PaymentId"]
	status := params["Status"]
	rebillID := params["RebillId"] // непустой только в webhook первого Recurrent=Y платежа

	slog.Info("subscription.webhook.received", "provider", provider, "external_id", externalID, "status", status, "has_rebill_id", rebillID != "")

	// Находим платёж по external_id
	pay, err := s.pays.GetByExternalID(ctx, provider, externalID)
	if err != nil {
		slog.Error("subscription.webhook.payment_not_found", "provider", provider, "external_id", externalID, "error", err)
		return fmt.Errorf("subscription.webhook: платёж не найден: %w", err)
	}

	// Idempotency: refunded и failed — терминальные состояния, повторная обработка
	// ничего не меняет. succeeded — НЕ терминальное: из него возможен переход
	// в refunded (юзер сделал возврат после успешной оплаты).
	if pay.Status == models.PaymentRefunded || pay.Status == models.PaymentFailed {
		slog.Info("subscription.webhook.terminal_state", "payment_id", pay.ID, "status", pay.Status)
		return nil
	}

	// Маппим статус T-Bank → наш
	var newStatus models.PaymentStatus
	switch status {
	case "CONFIRMED":
		newStatus = models.PaymentSucceeded
	case "REJECTED", "REVERSED", "DEADLINE_EXPIRED":
		newStatus = models.PaymentFailed
	case "REFUNDED", "PARTIAL_REFUNDED":
		newStatus = models.PaymentRefunded
	default:
		slog.Info("subscription.webhook.ignored_status", "status", status, "payment_id", pay.ID)
		return nil
	}

	// Не даунгрейдим статус: succeeded → succeeded игнорируем.
	if newStatus == pay.Status {
		slog.Info("subscription.webhook.same_status", "payment_id", pay.ID, "status", pay.Status)
		return nil
	}

	// Conditional UPDATE предотвращает race при параллельных webhook'ах:
	// только один запрос перейдёт к активации подписки. Если два webhook'а
	// пришли одновременно, второй получит transitioned=false и вернётся.
	transitioned, err := s.pays.TransitionStatus(ctx, pay.ID, pay.Status, newStatus)
	if err != nil {
		slog.Error("subscription.webhook.transition_failed", "payment_id", pay.ID, "error", err)
		return fmt.Errorf("subscription.webhook: не удалось обновить статус: %w", err)
	}
	if !transitioned {
		slog.Info("subscription.webhook.concurrent_skip", "payment_id", pay.ID, "expected", pay.Status, "next", newStatus)
		return nil
	}

	// При успешной оплате — активируем подписку. Только один webhook дойдёт сюда
	// благодаря conditional UPDATE выше.
	if newStatus == models.PaymentSucceeded {
		if err := s.activateSubscription(ctx, pay, rebillID); err != nil {
			return err
		}
	}

	// При возврате — экспайрим текущую подписку юзера и даунгрейдим на free.
	// Юзер вернул деньги, значит не должен оставаться на платном тарифе.
	if newStatus == models.PaymentRefunded {
		if err := s.handleRefund(ctx, pay); err != nil {
			return err
		}
	}

	return nil
}

// handleRefund обрабатывает возврат платежа: экспайрит активную подписку
// юзера и переводит его на free. Если подписки нет (уже expired) — просто
// обновляет users.plan_id.
func (s *Service) handleRefund(ctx context.Context, pay *models.Payment) error {
	sub, err := s.subs.GetActiveByUserID(ctx, pay.UserID)
	if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return fmt.Errorf("subscription.refund.get_active: %w", err)
	}
	if sub != nil {
		if expErr := s.subs.ExpireAndDowngrade(ctx, sub.ID, pay.UserID); expErr != nil {
			slog.Error("subscription.refund.expire_failed", "user_id", pay.UserID, "error", expErr)
			return fmt.Errorf("subscription.refund.expire: %w", expErr)
		}
	} else {
		// Нет активной — просто убеждаемся что users.plan_id=free.
		if updErr := s.users.Update(ctx, &models.User{ID: pay.UserID, PlanID: "free"}); updErr != nil {
			slog.Error("subscription.refund.update_plan_failed", "user_id", pay.UserID, "error", updErr)
		}
	}
	slog.Info("subscription.refund.processed", "user_id", pay.UserID, "payment_id", pay.ID)
	return nil
}

// activateSubscription создаёт/обновляет подписку после успешной оплаты.
// PlanID читается из Payment.ProviderData — это защищает от неправильной активации,
// если у двух планов совпадают цены (напр., промо-акция Max за цену Pro).
// rebillID — выдаётся T-Bank в webhook первого Recurrent=Y платежа; сохраняется
// в Subscription и используется renewLoop для безакцептных списаний.
func (s *Service) activateSubscription(ctx context.Context, pay *models.Payment, rebillID string) error {
	planID, err := extractPlanID(pay)
	if err != nil {
		slog.Error("subscription.activate.plan_id_missing", "payment_id", pay.ID, "error", err)
		return fmt.Errorf("subscription.activate: %w", err)
	}

	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		slog.Error("subscription.activate.plan_not_found", "plan_id", planID, "error", err)
		return fmt.Errorf("subscription.activate: план %q не найден: %w", planID, err)
	}

	// Renewal-флаг ProviderData отличает автопродление от первой оплаты:
	// renewal=true → продлеваем существующую подписку (ExtendPeriod), а не
	// создаём новую (иначе нарушим partial unique index "одна active на юзера").
	isRenewal := isRenewalPayment(pay)

	// Expire старую подписку (если есть) — при upgrade Pro→Max.
	// Не делаем при renewal — там подписка та же, период просто продлевается.
	existing, existErr := s.subs.GetActiveByUserID(ctx, pay.UserID)
	if existErr != nil && !errors.Is(existErr, repo.ErrNotFound) {
		slog.Error("subscription.activate.get_existing_failed", "user_id", pay.UserID, "error", existErr)
	}
	if isRenewal && existing != nil {
		newEnd := existing.CurrentPeriodEnd.AddDate(0, 0, plan.PeriodDays)
		if err := s.subs.ExtendPeriod(ctx, existing.ID, newEnd); err != nil {
			slog.Error("subscription.renewal.extend_failed", "sub_id", existing.ID, "error", err)
			return fmt.Errorf("subscription.renewal.extend: %w", err)
		}
		slog.Info("subscription.renewal.extended",
			"sub_id", existing.ID, "user_id", pay.UserID, "plan_id", plan.ID, "new_end", newEnd)
		return nil
	}
	if existing != nil {
		if expErr := s.subs.ExpireAndDowngrade(ctx, existing.ID, pay.UserID); expErr != nil {
			slog.Error("subscription.activate.expire_old_failed", "user_id", pay.UserID, "old_sub_id", existing.ID, "error", expErr)
		}
	}

	now := time.Now()
	sub := &models.Subscription{
		UserID:             pay.UserID,
		PlanID:             plan.ID,
		Status:             models.SubStatusActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 0, plan.PeriodDays),
		RebillId:           rebillID,
		AutoRenew:          true,
	}

	if err := s.subs.ActivateWithPlanUpdate(ctx, sub, pay.UserID, plan.ID); err != nil {
		slog.Error("subscription.activate.failed", "user_id", pay.UserID, "plan_id", plan.ID, "error", err)
		return fmt.Errorf("subscription.activate: %w", err)
	}

	// Связываем платёж с созданной подпиской. Non-critical: если UPDATE упадёт,
	// подписка уже активна и юзер не должен страдать — логируем Warn для Sentry.
	// Без связки handleRefund может экспайрить неправильную подписку при upgrade.
	if linkErr := s.pays.LinkSubscription(ctx, pay.ID, sub.ID); linkErr != nil {
		slog.Warn("subscription.activate.link_payment_failed",
			"payment_id", pay.ID, "sub_id", sub.ID, "user_id", pay.UserID, "error", linkErr)
	}

	// Warn если первичная активация с auto_renew=true, но T-Bank не выдал RebillId.
	// В prod это не должно случаться (Recurrent=Y всегда возвращает RebillId в CONFIRMED
	// webhook). Ловим через Sentry чтобы узнать раньше, чем юзер попытается автопродлиться.
	if rebillID == "" && sub.AutoRenew && s.cfg != nil && s.cfg.RecurrentEnabled {
		slog.Warn("subscription.activate.missing_rebill_id",
			"user_id", pay.UserID, "sub_id", sub.ID, "payment_id", pay.ID,
			"hint", "T-Bank не вернул RebillId — автопродление не сработает")
	}

	slog.Info("subscription.activate.success",
		"user_id", pay.UserID,
		"plan_id", plan.ID,
		"period_end", sub.CurrentPeriodEnd,
		"has_rebill_id", rebillID != "",
	)
	return nil
}

// SetAutoRenew управляет автопродлением подписки текущего пользователя.
// Если active подписки нет — возвращает ErrNoActiveSubscription.
func (s *Service) SetAutoRenew(ctx context.Context, userID uint, autoRenew bool) error {
	sub, err := s.subs.GetActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrNoActiveSubscription
		}
		return err
	}
	if err := s.subs.SetAutoRenew(ctx, sub.ID, autoRenew); err != nil {
		return fmt.Errorf("subscription.set_auto_renew: %w", err)
	}
	slog.Info("subscription.auto_renew.changed", "user_id", userID, "sub_id", sub.ID, "auto_renew", autoRenew)
	return nil
}

// extractPlanID достаёт plan_id из Payment.ProviderData. Возвращает ошибку,
// если поле пустое, невалидный JSON или не содержит plan_id.
func extractPlanID(pay *models.Payment) (string, error) {
	if len(pay.ProviderData) == 0 {
		return "", fmt.Errorf("provider_data пуст для payment %d", pay.ID)
	}
	var data PaymentProviderData
	if err := json.Unmarshal(pay.ProviderData, &data); err != nil {
		return "", fmt.Errorf("unmarshal provider_data: %w", err)
	}
	if data.PlanID == "" {
		return "", fmt.Errorf("plan_id отсутствует в provider_data для payment %d", pay.ID)
	}
	return data.PlanID, nil
}

// isRenewalPayment проверяет флаг "renewal" в ProviderData. true — платёж создан
// renewLoop (не первый checkout юзера), значит надо ExtendPeriod, а не Activate.
// Ошибку unmarshal логируем, чтобы не маскировать битый JSONB (Q-4).
func isRenewalPayment(pay *models.Payment) bool {
	if len(pay.ProviderData) == 0 {
		return false
	}
	var data PaymentProviderData
	if err := json.Unmarshal(pay.ProviderData, &data); err != nil {
		slog.Error("subscription.is_renewal.unmarshal_failed",
			"payment_id", pay.ID, "error", err)
		return false
	}
	return data.Renewal == "true"
}

// generateIdempotencyKey генерирует криптографически случайный ключ (32 hex символа).
func generateIdempotencyKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// buildReceipt формирует фискальный чек 54-ФЗ для подписки. Возвращает nil
// если ReceiptEnabled=false (онлайн-касса не подключена) — тогда T-Bank
// не будет формировать чек, а в кабинете не появится фискальный документ.
//
// Подписка квалифицируется как услуга (PaymentObject=service) с полной оплатой
// (PaymentMethod=full_payment). НДС=none для УСН.
func buildReceipt(cfg *config.PaymentConfig, email string, plan *models.SubscriptionPlan) *payment.Receipt {
	if cfg == nil || !cfg.ReceiptEnabled {
		return nil
	}
	return &payment.Receipt{
		Email:    email,
		Taxation: cfg.Taxation,
		Items: []payment.ReceiptItem{
			{
				Name:          fmt.Sprintf("Подписка ПромтЛаб %s на 1 месяц", plan.Name),
				PriceKop:      plan.PriceKop,
				Quantity:      1,
				AmountKop:     plan.PriceKop,
				Tax:           "none",
				PaymentMethod: "full_payment",
				PaymentObject: "service",
			},
		},
	}
}
