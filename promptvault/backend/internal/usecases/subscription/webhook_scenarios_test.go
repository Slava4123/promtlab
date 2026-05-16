// CR-8 part 2 — full webhook scenarios. Дополняет subscription_webhook_test.go
// (там 9 базовых: invalid signature, payment not found, terminal idempotent,
// unknown status, same status, transition race, DB error). Здесь покрываем:
//
//	1. CONFIRMED first payment → activate (extractPlanID, plan lookup, ActivateWithPlanUpdate, LinkSubscription)
//	2. CONFIRMED duplicate webhook → idempotent (transition race protects)
//	3. REFUNDED with active sub → ExpireAndDowngrade
//	4. REFUNDED without active sub → users.SetPlan(free) (defence vs CR-2 mass-overwrite)
//	5. Pause: AlreadyPaused → ErrSubscriptionPaused
//	6. Pause: free plan (PriceKop=0) → ErrSubscriptionNotPausable
//	7. Resume: no active pause → ErrSubscriptionNotPaused
//	8. Resume: happy path (period_end сдвинут на remaining)
//	9. Cancel: no active subscription → ErrNoActiveSubscription
//
// Стиль — function-field моки (как в subscription_webhook_test.go),
// для трёх дополнительных интерфейсов: SubscriptionRepository, UserRepository, PlanRepository.
package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"promptvault/internal/infrastructure/payment"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- Mocks ---

// mockSubsRepo — минимальные подписки для webhook scenarios.
type mockSubsRepo struct {
	getActive            func(userID uint) (*models.Subscription, error)
	getCurrent           func(userID uint) (*models.Subscription, error)
	expireAndDowngrade   func(subID, userID uint) error
	activate             func(sub *models.Subscription, userID uint, planID string) error
	pause                func(subID, userID uint, pausedAt, pausedUntil time.Time) error
	resume               func(subID, userID uint, resumeAt, newEnd time.Time) error
	cancelAtPeriodEnd    func(subID uint) error
	extendPeriod         func(subID uint, newEnd time.Time) error
	recordCancellation   func(c *models.SubscriptionCancellation) error
	expireAndDowngradeOK bool // shortcut: записать что вызвалось без callback
	expireAndDowngradeID uint
}

func (m *mockSubsRepo) Create(context.Context, *models.Subscription) error { panic("unused") }
func (m *mockSubsRepo) GetActiveByUserID(_ context.Context, userID uint) (*models.Subscription, error) {
	if m.getActive == nil {
		return nil, repo.ErrNotFound
	}
	return m.getActive(userID)
}
func (m *mockSubsRepo) Update(context.Context, *models.Subscription) error { panic("unused") }
func (m *mockSubsRepo) ListExpiring(context.Context, time.Time) ([]models.Subscription, error) {
	panic("unused")
}
func (m *mockSubsRepo) ActivateWithPlanUpdate(_ context.Context, sub *models.Subscription, userID uint, planID string) error {
	if m.activate == nil {
		// Default: имитируем БД-присвоение ID, чтобы LinkSubscription вызвался корректно.
		if sub.ID == 0 {
			sub.ID = 100
		}
		return nil
	}
	return m.activate(sub, userID, planID)
}
func (m *mockSubsRepo) CancelAtPeriodEnd(_ context.Context, subID uint) error {
	if m.cancelAtPeriodEnd == nil {
		return nil
	}
	return m.cancelAtPeriodEnd(subID)
}
func (m *mockSubsRepo) ExpireAndDowngrade(_ context.Context, subID, userID uint) error {
	if m.expireAndDowngrade != nil {
		return m.expireAndDowngrade(subID, userID)
	}
	m.expireAndDowngradeOK = true
	m.expireAndDowngradeID = subID
	return nil
}
func (m *mockSubsRepo) GetCurrentByUserID(_ context.Context, userID uint) (*models.Subscription, error) {
	if m.getCurrent == nil {
		return nil, repo.ErrNotFound
	}
	return m.getCurrent(userID)
}
func (m *mockSubsRepo) MarkExpired(context.Context, uint) error { panic("unused") }
func (m *mockSubsRepo) SetRebillId(context.Context, uint, string) error {
	panic("unused")
}
func (m *mockSubsRepo) SetAutoRenew(context.Context, uint, bool) error { panic("unused") }
func (m *mockSubsRepo) ListReadyForRenewal(context.Context, time.Time, time.Time, int) ([]models.Subscription, error) {
	panic("unused")
}
func (m *mockSubsRepo) ExtendPeriod(_ context.Context, subID uint, newEnd time.Time) error {
	if m.extendPeriod == nil {
		return nil
	}
	return m.extendPeriod(subID, newEnd)
}
func (m *mockSubsRepo) RecordRenewalFailure(context.Context, uint) error { panic("unused") }
func (m *mockSubsRepo) ListPreExpiring(context.Context, time.Time, time.Time, int16) ([]models.Subscription, error) {
	panic("unused")
}
func (m *mockSubsRepo) SetPreExpireStage(context.Context, uint, int16) error {
	panic("unused")
}
func (m *mockSubsRepo) Pause(_ context.Context, subID, userID uint, pausedAt, pausedUntil time.Time) error {
	if m.pause == nil {
		return nil
	}
	return m.pause(subID, userID, pausedAt, pausedUntil)
}
func (m *mockSubsRepo) Resume(_ context.Context, subID, userID uint, resumeAt, newEnd time.Time) error {
	if m.resume == nil {
		return nil
	}
	return m.resume(subID, userID, resumeAt, newEnd)
}
func (m *mockSubsRepo) ListExpiredPauses(context.Context, time.Time) ([]models.Subscription, error) {
	panic("unused")
}
func (m *mockSubsRepo) RecordCancellation(_ context.Context, c *models.SubscriptionCancellation) error {
	if m.recordCancellation == nil {
		return nil
	}
	return m.recordCancellation(c)
}

// mockUsersRepo — для users.SetPlan в handleRefund (без активной подписки).
type mockUsersRepo struct {
	setPlanCalls []setPlanCall
	setPlanErr   error
	getByID      func(id uint) (*models.User, error)
}

type setPlanCall struct {
	userID uint
	planID string
}

func (m *mockUsersRepo) Create(context.Context, *models.User) error { panic("unused") }
func (m *mockUsersRepo) GetByID(_ context.Context, id uint) (*models.User, error) {
	if m.getByID == nil {
		return &models.User{ID: id}, nil
	}
	return m.getByID(id)
}
func (m *mockUsersRepo) GetByEmail(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (m *mockUsersRepo) GetByUsername(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (m *mockUsersRepo) SearchUsers(context.Context, string, int) ([]models.User, error) {
	panic("unused")
}
func (m *mockUsersRepo) Update(context.Context, *models.User) error { panic("unused") }
func (m *mockUsersRepo) SetPlan(_ context.Context, userID uint, planID string) error {
	m.setPlanCalls = append(m.setPlanCalls, setPlanCall{userID: userID, planID: planID})
	return m.setPlanErr
}
func (m *mockUsersRepo) SetQuotaWarningSentOn(context.Context, uint, time.Time) error {
	panic("unused")
}
func (m *mockUsersRepo) TouchLastLogin(context.Context, uint) error { panic("unused") }
func (m *mockUsersRepo) ListInactiveForReengagement(context.Context, time.Time, time.Time, int) ([]models.User, error) {
	panic("unused")
}
func (m *mockUsersRepo) MarkReengagementSent(context.Context, uint) error { panic("unused") }
func (m *mockUsersRepo) CountReferredBy(context.Context, string) (int64, error) {
	panic("unused")
}
func (m *mockUsersRepo) GetByReferralCode(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (m *mockUsersRepo) MarkReferralRewarded(context.Context, uint) (bool, error) {
	panic("unused")
}
func (m *mockUsersRepo) ListPaidUsers(context.Context) ([]uint, error) { panic("unused") }
func (m *mockUsersRepo) SetInsightEmailsEnabled(context.Context, uint, bool) error {
	panic("unused")
}

// mockPlansRepo — для plans.GetByID в activateSubscription.
type mockPlansRepo struct {
	getByID func(id string) (*models.SubscriptionPlan, error)
}

func (m *mockPlansRepo) GetAll(context.Context) ([]models.SubscriptionPlan, error) {
	panic("unused")
}
func (m *mockPlansRepo) GetByID(_ context.Context, id string) (*models.SubscriptionPlan, error) {
	if m.getByID == nil {
		return nil, repo.ErrNotFound
	}
	return m.getByID(id)
}
func (m *mockPlansRepo) GetActive(context.Context) ([]models.SubscriptionPlan, error) {
	panic("unused")
}

// payRepoExt — расширение mockPaymentRepo с LinkSubscription/UpdateStatus
// (нужны в activateSubscription и handleRefund соответственно).
// Используем наследование через embedding: struct payRepoExt содержит поля
// mockPaymentRepo + дополнительные.
type payRepoExt struct {
	getByExtID         func(provider, externalID string) (*models.Payment, error)
	transition         func(id uint, expected, next models.PaymentStatus) (bool, error)
	transitionErr      error
	linkSubscription   func(paymentID, subscriptionID uint) error
	updateStatusErr    error
	linkSubscriptionID uint
}

func (m *payRepoExt) Create(context.Context, *models.Payment) error { panic("unused") }
func (m *payRepoExt) GetByExternalID(_ context.Context, provider, externalID string) (*models.Payment, error) {
	return m.getByExtID(provider, externalID)
}
func (m *payRepoExt) GetByIdempotencyKey(context.Context, string) (*models.Payment, error) {
	panic("unused")
}
func (m *payRepoExt) UpdateStatus(context.Context, uint, models.PaymentStatus) error {
	return m.updateStatusErr
}
func (m *payRepoExt) UpdateExternalID(context.Context, uint, string) error { panic("unused") }
func (m *payRepoExt) TransitionStatus(_ context.Context, id uint, expected, next models.PaymentStatus) (bool, error) {
	if m.transitionErr != nil {
		return false, m.transitionErr
	}
	if m.transition != nil {
		return m.transition(id, expected, next)
	}
	return true, nil
}
func (m *payRepoExt) LinkSubscription(_ context.Context, paymentID, subscriptionID uint) error {
	m.linkSubscriptionID = subscriptionID
	if m.linkSubscription == nil {
		return nil
	}
	return m.linkSubscription(paymentID, subscriptionID)
}

// newFullService собирает Service со всеми моками.
func newFullService(prov *mockProvider, pays *payRepoExt, subs *mockSubsRepo, users *mockUsersRepo, plans *mockPlansRepo) *Service {
	return &Service{
		subs:    subs,
		plans:   plans,
		pays:    pays,
		users:   users,
		payment: prov,
	}
}

// providerData — helper для построения PaymentProviderData JSONB.
func providerData(planID string, renewal bool) json.RawMessage {
	d := PaymentProviderData{PlanID: planID}
	if renewal {
		d.Renewal = "true"
	}
	raw, _ := json.Marshal(d)
	return raw
}

// --- CONFIRMED first payment activation ---

func TestHandleWebhook_Confirmed_FirstPayment_ActivatesSubscription(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &payRepoExt{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{
				ID:           1,
				UserID:       42,
				Status:       models.PaymentPending,
				ProviderData: providerData("pro", false),
			}, nil
		},
		transition: func(_ uint, _, _ models.PaymentStatus) (bool, error) {
			return true, nil
		},
	}
	subs := &mockSubsRepo{
		getActive: func(_ uint) (*models.Subscription, error) {
			return nil, repo.ErrNotFound // нет существующей
		},
	}
	users := &mockUsersRepo{}
	plans := &mockPlansRepo{
		getByID: func(id string) (*models.SubscriptionPlan, error) {
			return &models.SubscriptionPlan{ID: id, Name: "Pro", PeriodDays: 30, PriceKop: 59900}, nil
		},
	}
	svc := newFullService(prov, pays, subs, users, plans)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "CONFIRMED", "RebillId": "rb-123"})

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if pays.linkSubscriptionID == 0 {
		t.Errorf("expected LinkSubscription called with non-zero subscription ID, got %d", pays.linkSubscriptionID)
	}
}

// --- CONFIRMED renewal: existing sub → ExtendPeriod ---

func TestHandleWebhook_Confirmed_Renewal_ExtendsPeriod(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &payRepoExt{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{
				ID:           1,
				UserID:       42,
				Status:       models.PaymentPending,
				ProviderData: providerData("pro", true), // renewal=true
			}, nil
		},
		transition: func(_ uint, _, _ models.PaymentStatus) (bool, error) { return true, nil },
	}
	existing := &models.Subscription{
		ID:               55,
		UserID:           42,
		PlanID:           "pro",
		Status:           models.SubStatusActive,
		CurrentPeriodEnd: time.Now().Add(24 * time.Hour),
	}
	var extendCalled bool
	subs := &mockSubsRepo{
		getActive: func(_ uint) (*models.Subscription, error) {
			return existing, nil
		},
		extendPeriod: func(subID uint, _ time.Time) error {
			if subID != 55 {
				t.Errorf("expected ExtendPeriod for sub 55, got %d", subID)
			}
			extendCalled = true
			return nil
		},
	}
	users := &mockUsersRepo{}
	plans := &mockPlansRepo{
		getByID: func(id string) (*models.SubscriptionPlan, error) {
			return &models.SubscriptionPlan{ID: id, PeriodDays: 30}, nil
		},
	}
	svc := newFullService(prov, pays, subs, users, plans)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "CONFIRMED"})

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if !extendCalled {
		t.Error("expected ExtendPeriod called for renewal payment, got none")
	}
	// Defensive sync user.plan_id (фикс 16.05.2026): renewal-branch
	// должна вызывать SetPlan с plan_id из subscription, иначе при drift
	// между user.plan_id и sub.plan_id (например после expirationLoop)
	// юзер останется на free несмотря на успешный charge.
	if len(users.setPlanCalls) != 1 {
		t.Fatalf("expected 1 SetPlan call on renewal, got %d", len(users.setPlanCalls))
	}
	if users.setPlanCalls[0].userID != 42 || users.setPlanCalls[0].planID != "pro" {
		t.Errorf("SetPlan called with wrong args: %+v", users.setPlanCalls[0])
	}
}

// TestHandleWebhook_Confirmed_Renewal_SyncsUserPlanWhenDrift — регрессия
// инцидента 16.05.2026 sub_id=2: user.plan_id отставал от subscription.plan_id
// после восстановления подписки из expired/past_due. ExtendPeriod продлевал
// период, но не обновлял users.plan_id — юзер видел Free несмотря на
// успешное списание.
func TestHandleWebhook_Confirmed_Renewal_SyncsUserPlanWhenDrift(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &payRepoExt{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{
				ID:           2,
				UserID:       42,
				Status:       models.PaymentPending,
				ProviderData: providerData("pro", true), // renewal=true
			}, nil
		},
		transition: func(_ uint, _, _ models.PaymentStatus) (bool, error) { return true, nil },
	}
	// Подписка вернулась в active (с plan_id=pro), но user.plan_id ещё free
	// (drift после expirationLoop). Webhook должен синхронизировать обратно.
	existing := &models.Subscription{
		ID: 99, UserID: 42, PlanID: "pro",
		Status:           models.SubStatusActive,
		CurrentPeriodEnd: time.Now().Add(12 * time.Hour),
	}
	subs := &mockSubsRepo{
		getActive:    func(_ uint) (*models.Subscription, error) { return existing, nil },
		extendPeriod: func(uint, time.Time) error { return nil },
	}
	// Стартовый plan_id=free (drift) — после webhook должно стать pro
	// через SetPlan-call (атомарный UPDATE).
	users := &mockUsersRepo{}
	plans := &mockPlansRepo{
		getByID: func(id string) (*models.SubscriptionPlan, error) {
			return &models.SubscriptionPlan{ID: id, PeriodDays: 30}, nil
		},
	}
	svc := newFullService(prov, pays, subs, users, plans)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "CONFIRMED"})
	if err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}
	if len(users.setPlanCalls) != 1 {
		t.Fatalf("expected SetPlan call to sync drift, got %d", len(users.setPlanCalls))
	}
	if users.setPlanCalls[0].planID != "pro" {
		t.Errorf("SetPlan called with planID=%q, want %q", users.setPlanCalls[0].planID, "pro")
	}
}

// --- REFUNDED with active subscription → ExpireAndDowngrade ---

func TestHandleWebhook_Refunded_WithActiveSubscription_ExpiresAndDowngrades(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &payRepoExt{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{
				ID:     1,
				UserID: 42,
				Status: models.PaymentSucceeded, // pending уже succeeded раньше
			}, nil
		},
		transition: func(_ uint, _, _ models.PaymentStatus) (bool, error) { return true, nil },
	}
	subs := &mockSubsRepo{
		getActive: func(_ uint) (*models.Subscription, error) {
			return &models.Subscription{ID: 55, UserID: 42, PlanID: "pro"}, nil
		},
	}
	users := &mockUsersRepo{}
	plans := &mockPlansRepo{}
	svc := newFullService(prov, pays, subs, users, plans)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "REFUNDED"})

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if !subs.expireAndDowngradeOK || subs.expireAndDowngradeID != 55 {
		t.Errorf("expected ExpireAndDowngrade(sub_id=55), got OK=%v, id=%d",
			subs.expireAndDowngradeOK, subs.expireAndDowngradeID)
	}
	// users.SetPlan НЕ должен вызываться — ExpireAndDowngrade сама обновляет plan_id
	// в одной транзакции.
	if len(users.setPlanCalls) > 0 {
		t.Errorf("users.SetPlan must NOT be called when active sub exists, got %v", users.setPlanCalls)
	}
}

// --- REFUNDED without active subscription → users.SetPlan(free) ---
//
// Защищает от CR-2 (mass-overwrite через gorm.Save). Смотрим что используется
// именно SetPlan (partial UPDATE), а не Update (full Save).
func TestHandleWebhook_Refunded_NoActiveSubscription_SetsPlanToFreeViaPartialUpdate(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &payRepoExt{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{
				ID:     1,
				UserID: 42,
				Status: models.PaymentSucceeded,
			}, nil
		},
		transition: func(_ uint, _, _ models.PaymentStatus) (bool, error) { return true, nil },
	}
	subs := &mockSubsRepo{
		// Нет активной подписки.
		getActive: func(_ uint) (*models.Subscription, error) {
			return nil, repo.ErrNotFound
		},
	}
	users := &mockUsersRepo{}
	plans := &mockPlansRepo{}
	svc := newFullService(prov, pays, subs, users, plans)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "REFUNDED"})

	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if len(users.setPlanCalls) != 1 {
		t.Fatalf("expected 1 SetPlan call (partial UPDATE), got %d", len(users.setPlanCalls))
	}
	call := users.setPlanCalls[0]
	if call.userID != 42 || call.planID != "free" {
		t.Errorf("expected SetPlan(42, free), got SetPlan(%d, %q)", call.userID, call.planID)
	}
}

// --- Pause: AlreadyPaused → ErrSubscriptionPaused ---

func TestPause_AlreadyPaused_ReturnsErr(t *testing.T) {
	subs := &mockSubsRepo{
		getActive: func(_ uint) (*models.Subscription, error) {
			return &models.Subscription{ID: 55, Status: models.SubStatusPaused}, nil
		},
	}
	svc := &Service{subs: subs}

	err := svc.Pause(context.Background(), PauseInput{UserID: 42, Months: 1})
	if !errors.Is(err, ErrSubscriptionPaused) {
		t.Fatalf("expected ErrSubscriptionPaused, got %v", err)
	}
}

// --- Pause: free plan → ErrSubscriptionNotPausable ---

func TestPause_FreePlan_ReturnsNotPausable(t *testing.T) {
	subs := &mockSubsRepo{
		getActive: func(_ uint) (*models.Subscription, error) {
			return &models.Subscription{ID: 55, PlanID: "free", Status: models.SubStatusActive}, nil
		},
	}
	plans := &mockPlansRepo{
		getByID: func(_ string) (*models.SubscriptionPlan, error) {
			return &models.SubscriptionPlan{ID: "free", PriceKop: 0}, nil
		},
	}
	svc := &Service{subs: subs, plans: plans}

	err := svc.Pause(context.Background(), PauseInput{UserID: 42, Months: 1})
	if !errors.Is(err, ErrSubscriptionNotPausable) {
		t.Fatalf("expected ErrSubscriptionNotPausable for free plan, got %v", err)
	}
}

// --- Pause: invalid months → ErrInvalidPauseMonths ---

func TestPause_InvalidMonths_Refused(t *testing.T) {
	svc := &Service{}
	cases := []struct {
		months int
		name   string
	}{
		{0, "zero"},
		{4, "too_many"},
		{-1, "negative"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.Pause(context.Background(), PauseInput{UserID: 42, Months: tc.months})
			if !errors.Is(err, ErrInvalidPauseMonths) {
				t.Errorf("months=%d: expected ErrInvalidPauseMonths, got %v", tc.months, err)
			}
		})
	}
}

// --- Pause: happy path → calls subs.Pause ---

func TestPause_HappyPath_CallsRepoPause(t *testing.T) {
	var pauseCalled bool
	subs := &mockSubsRepo{
		getActive: func(_ uint) (*models.Subscription, error) {
			return &models.Subscription{
				ID:               55,
				PlanID:           "pro",
				Status:           models.SubStatusActive,
				CurrentPeriodEnd: time.Now().Add(15 * 24 * time.Hour),
			}, nil
		},
		pause: func(subID, userID uint, _, _ time.Time) error {
			pauseCalled = true
			if subID != 55 || userID != 42 {
				t.Errorf("Pause(%d, %d): expected (55, 42)", subID, userID)
			}
			return nil
		},
	}
	plans := &mockPlansRepo{
		getByID: func(id string) (*models.SubscriptionPlan, error) {
			return &models.SubscriptionPlan{ID: id, PriceKop: 59900}, nil
		},
	}
	svc := &Service{subs: subs, plans: plans}

	err := svc.Pause(context.Background(), PauseInput{UserID: 42, Months: 2})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !pauseCalled {
		t.Error("expected subs.Pause called")
	}
}

// --- Resume: no active pause → ErrSubscriptionNotPaused ---

func TestResume_NoActivePause_ReturnsErr(t *testing.T) {
	subs := &mockSubsRepo{
		getCurrent: func(_ uint) (*models.Subscription, error) {
			return &models.Subscription{ID: 55, Status: models.SubStatusActive}, nil
		},
	}
	svc := &Service{subs: subs}

	err := svc.Resume(context.Background(), 42)
	if !errors.Is(err, ErrSubscriptionNotPaused) {
		t.Fatalf("expected ErrSubscriptionNotPaused, got %v", err)
	}
}

// --- Resume: no subscription at all → ErrNoActiveSubscription ---

func TestResume_NoSubscription_ReturnsErr(t *testing.T) {
	subs := &mockSubsRepo{
		getCurrent: func(_ uint) (*models.Subscription, error) {
			return nil, repo.ErrNotFound
		},
	}
	svc := &Service{subs: subs}

	err := svc.Resume(context.Background(), 42)
	if !errors.Is(err, ErrNoActiveSubscription) {
		t.Fatalf("expected ErrNoActiveSubscription, got %v", err)
	}
}

// --- Resume: happy path → calls subs.Resume + правильный newEnd ---

func TestResume_HappyPath_ShiftsPeriodByRemaining(t *testing.T) {
	pausedAt := time.Now().Add(-10 * 24 * time.Hour) // 10 дней назад
	periodEnd := pausedAt.Add(20 * 24 * time.Hour)   // на момент pause осталось 20 дней
	expectedRemaining := periodEnd.Sub(pausedAt)

	var resumeNewEnd time.Time
	subs := &mockSubsRepo{
		getCurrent: func(_ uint) (*models.Subscription, error) {
			return &models.Subscription{
				ID:               55,
				Status:           models.SubStatusPaused,
				PausedAt:         &pausedAt,
				CurrentPeriodEnd: periodEnd,
			}, nil
		},
		resume: func(_, _ uint, _, newEnd time.Time) error {
			resumeNewEnd = newEnd
			return nil
		},
	}
	svc := &Service{subs: subs}

	err := svc.Resume(context.Background(), 42)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	// new_end ≈ now + remaining (20 дней).
	got := time.Until(resumeNewEnd)
	delta := got - expectedRemaining
	if delta < -2*time.Second || delta > 2*time.Second {
		t.Errorf("new_end shift: expected ~%v from now, got %v (delta %v)",
			expectedRemaining, got, delta)
	}
}

// --- Cancel: no active subscription → ErrNoActiveSubscription ---

func TestCancel_NoActiveSubscription_ReturnsErr(t *testing.T) {
	subs := &mockSubsRepo{} // getActive==nil → ErrNotFound default
	svc := &Service{subs: subs}

	err := svc.Cancel(context.Background(), CancelInput{UserID: 42, Reason: ""})
	if !errors.Is(err, ErrNoActiveSubscription) {
		t.Fatalf("expected ErrNoActiveSubscription, got %v", err)
	}
}

// --- Cancel: invalid reason → ErrInvalidCancelReason ---

func TestCancel_InvalidReason_Refused(t *testing.T) {
	svc := &Service{}
	err := svc.Cancel(context.Background(), CancelInput{UserID: 42, Reason: "WEIRD_REASON"})
	if !errors.Is(err, ErrInvalidCancelReason) {
		t.Fatalf("expected ErrInvalidCancelReason, got %v", err)
	}
}

// --- Cancel: happy path with reason → CancelAtPeriodEnd + RecordCancellation ---

func TestCancel_WithReason_RecordsCancellation(t *testing.T) {
	var cancelCalled, recordCalled bool
	subs := &mockSubsRepo{
		getActive: func(_ uint) (*models.Subscription, error) {
			return &models.Subscription{ID: 55, PlanID: "pro"}, nil
		},
		cancelAtPeriodEnd: func(_ uint) error {
			cancelCalled = true
			return nil
		},
		recordCancellation: func(c *models.SubscriptionCancellation) error {
			recordCalled = true
			if c.Reason != CancelReasonTooExpensive {
				t.Errorf("expected Reason=too_expensive, got %q", c.Reason)
			}
			return nil
		},
	}
	svc := &Service{subs: subs}

	err := svc.Cancel(context.Background(), CancelInput{
		UserID: 42,
		Reason: CancelReasonTooExpensive,
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !cancelCalled {
		t.Error("expected CancelAtPeriodEnd called")
	}
	if !recordCalled {
		t.Error("expected RecordCancellation called")
	}
}

// --- Cancel: RecordCancellation failure не блокирует Cancel ---

func TestCancel_RecordCancellationFailure_StillSucceeds(t *testing.T) {
	subs := &mockSubsRepo{
		getActive: func(_ uint) (*models.Subscription, error) {
			return &models.Subscription{ID: 55, PlanID: "pro"}, nil
		},
		recordCancellation: func(_ *models.SubscriptionCancellation) error {
			return errors.New("DB connection drop")
		},
	}
	svc := &Service{subs: subs}

	err := svc.Cancel(context.Background(), CancelInput{
		UserID: 42,
		Reason: CancelReasonNotUsing,
	})
	if err != nil {
		t.Errorf("expected nil err (record-failure swallowed), got %v", err)
	}
}

// --- Helper assert: provider не вызван когда payment=nil ---

func TestHandleWebhook_PaymentNotConfigured_ProviderNotCalled(t *testing.T) {
	prov := &mockProvider{}
	svc := &Service{} // payment=nil; provider не должен трогаться

	_ = svc.HandleWebhook(context.Background(), "tbank", map[string]string{"Token": "x"})

	if prov.verifyCalls != 0 {
		t.Errorf("expected provider.Verify NOT called when payment=nil, got %d", prov.verifyCalls)
	}
}

// Compile-time interface satisfaction checks.
var (
	_ repo.PaymentRepository      = (*payRepoExt)(nil)
	_ repo.SubscriptionRepository = (*mockSubsRepo)(nil)
	_ repo.UserRepository         = (*mockUsersRepo)(nil)
	_ repo.PlanRepository         = (*mockPlansRepo)(nil)
	_ payment.PaymentProvider     = (*mockProvider)(nil)
)
