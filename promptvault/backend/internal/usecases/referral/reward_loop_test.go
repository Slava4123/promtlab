// backend/internal/usecases/referral/reward_loop_test.go
package referral

import (
	"context"
	"errors"
	"testing"
	"time"

	"promptvault/internal/models"
)

// Reusable helper: создаёт Service + RewardLoop с зафиксированным временем
// (тем же, что использует newTestService — 2026-05-17 12:00 UTC). Возвращает
// loop, чтобы тесты могли подменить SetNowFn если нужно.
func newTestLoop(svc *Service, pending *fakePendingRepo) *RewardLoop {
	loop := NewRewardLoop(svc, pending, time.Hour, 100)
	loop.SetNowFn(func() time.Time {
		return time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	})
	return loop
}

// TestRewardLoop_TickGrantsEligible — happy path: pending с eligible_at в
// прошлом, GrantReward возвращает nil → summary.Granted=1, pending удалён.
func TestRewardLoop_TickGrantsEligible(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()
	pending := newFakePendingRepo()

	// Pro-референт с активной подпиской — extend ветка.
	users.users[10] = &models.User{ID: 10, PlanID: "pro"}
	originalEnd := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	subs.activeByUser[10] = &models.Subscription{
		ID:               55,
		UserID:           10,
		PlanID:           "pro",
		Status:           models.SubStatusActive,
		CurrentPeriodEnd: originalEnd,
	}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	// pending eligible_at = 2026-05-01 (раньше "now" = 2026-05-17).
	pending.rows = []models.ReferralPendingReward{
		{
			ID:         101,
			ReferrerID: 10,
			RefereeID:  20,
			PaymentID:  7,
			EligibleAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	svc := newTestService(users, subs, pays)
	// Подменяем pending — newTestService создаёт фейковый, нам нужен наш.
	svc.pending = pending
	loop := newTestLoop(svc, pending)

	summary := loop.tickOnce(context.Background())

	if summary.Granted != 1 {
		t.Errorf("summary.Granted = %d, want 1", summary.Granted)
	}
	if summary.Total() != 1 {
		t.Errorf("summary.Total() = %d, want 1", summary.Total())
	}
	if len(pending.deletedIDs) != 1 || pending.deletedIDs[0] != 101 {
		t.Errorf("pending.deletedIDs = %v, want [101]", pending.deletedIDs)
	}
	if len(subs.updatePeriodC) != 1 {
		t.Errorf("expected 1 UpdatePeriodEnd call, got %d", len(subs.updatePeriodC))
	}
}

// TestRewardLoop_SkipRefundedAndDeletePending — payment отрефанжен →
// ErrPaymentRefunded — terminal skip, pending удалён (нет смысла retry'ить
// refunded payment). summary.SkippedRefund=1, pending удалён.
func TestRewardLoop_SkipRefundedAndDeletePending(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()
	pending := newFakePendingRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "pro"}
	// Payment отрефанжен после payment.succeeded webhook'а — eligibility
	// окно как раз для этого (14d > T-Bank refund window).
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentRefunded}

	pending.rows = []models.ReferralPendingReward{
		{
			ID:         102,
			ReferrerID: 10,
			RefereeID:  20,
			PaymentID:  7,
			EligibleAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	svc := newTestService(users, subs, pays)
	svc.pending = pending
	loop := newTestLoop(svc, pending)

	summary := loop.tickOnce(context.Background())

	if summary.SkippedRefund != 1 {
		t.Errorf("summary.SkippedRefund = %d, want 1", summary.SkippedRefund)
	}
	if summary.Granted != 0 {
		t.Errorf("summary.Granted = %d, want 0", summary.Granted)
	}
	if summary.Errors != 0 {
		t.Errorf("summary.Errors = %d, want 0 (refund — terminal skip, не error)", summary.Errors)
	}
	if len(pending.deletedIDs) != 1 || pending.deletedIDs[0] != 102 {
		t.Errorf("pending.deletedIDs = %v, want [102] (terminal skip удаляет pending)",
			pending.deletedIDs)
	}
	// Не должно быть никаких side-effects на подписке.
	if len(subs.updatePeriodC) != 0 {
		t.Errorf("no UpdatePeriodEnd expected on refunded pending")
	}
	if users.users[10].ReferralRewardedAt != nil {
		t.Errorf("ReferralRewardedAt should NOT be set on refund")
	}
}

// TestRewardLoop_TransientErrorKeepsPending — generic ошибка GrantReward
// (например, DB-сбой при extendActiveSubscription) → summary.Errors=1,
// pending НЕ удалён (retry на следующем тике, "at-least-once" семантика).
func TestRewardLoop_TransientErrorKeepsPending(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()
	pending := newFakePendingRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "pro"}
	// Подписка есть, но GetActiveByUserID возвращает transient error.
	subs.activeByUser[10] = &models.Subscription{
		ID:               55,
		UserID:           10,
		PlanID:           "pro",
		Status:           models.SubStatusActive,
		CurrentPeriodEnd: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	subs.getActiveErr = errors.New("db: connection refused")
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	pending.rows = []models.ReferralPendingReward{
		{
			ID:         103,
			ReferrerID: 10,
			RefereeID:  20,
			PaymentID:  7,
			EligibleAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	svc := newTestService(users, subs, pays)
	svc.pending = pending
	loop := newTestLoop(svc, pending)

	summary := loop.tickOnce(context.Background())

	if summary.Errors != 1 {
		t.Errorf("summary.Errors = %d, want 1", summary.Errors)
	}
	if summary.Granted != 0 {
		t.Errorf("summary.Granted = %d, want 0", summary.Granted)
	}
	// Critical: pending НЕ удалён — следующий тик retry'нет.
	if len(pending.deletedIDs) != 0 {
		t.Errorf("pending.deletedIDs = %v, want [] (transient error — keep for retry)",
			pending.deletedIDs)
	}
	if len(pending.rows) != 1 {
		t.Errorf("pending.rows count = %d, want 1 (row survives transient error)", len(pending.rows))
	}
}

// TestRewardLoop_EmptyEligibleNoOp — pending'ов нет (или все в будущем) →
// summary all zeros, никаких Delete'ов.
func TestRewardLoop_EmptyEligibleNoOp(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()
	pending := newFakePendingRepo()

	// pending.rows пуст — ListEligible вернёт [].
	svc := newTestService(users, subs, pays)
	svc.pending = pending
	loop := newTestLoop(svc, pending)

	summary := loop.tickOnce(context.Background())

	if summary.Total() != 0 {
		t.Errorf("summary.Total() = %d, want 0 (empty eligible queue)", summary.Total())
	}
	if len(pending.deletedIDs) != 0 {
		t.Errorf("pending.deletedIDs = %v, want []", pending.deletedIDs)
	}
}

// TestRewardLoop_EligibleAtInFutureNotProcessed — pending с eligible_at >
// nowFn() → ListEligible его не возвращает (фильтр по eligible_at <= now).
// Защита от слишком ранней выдачи (до истечения refund-окна).
func TestRewardLoop_EligibleAtInFutureNotProcessed(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()
	pending := newFakePendingRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "pro"}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	// eligible_at = 2026-06-01 (после nowFn = 2026-05-17).
	pending.rows = []models.ReferralPendingReward{
		{
			ID:         104,
			ReferrerID: 10,
			RefereeID:  20,
			PaymentID:  7,
			EligibleAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	svc := newTestService(users, subs, pays)
	svc.pending = pending
	loop := newTestLoop(svc, pending)

	summary := loop.tickOnce(context.Background())

	if summary.Total() != 0 {
		t.Errorf("summary.Total() = %d, want 0 (pending not yet eligible)", summary.Total())
	}
	if len(pending.deletedIDs) != 0 {
		t.Errorf("pending.deletedIDs = %v, want [] (don't delete future pendings)",
			pending.deletedIDs)
	}
}
