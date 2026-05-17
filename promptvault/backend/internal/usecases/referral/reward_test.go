// backend/internal/usecases/referral/reward_test.go
package referral

import (
	"context"
	"errors"
	"testing"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ===== Fakes — partial implementations =====
//
// Все методы интерфейса репозитория должны быть на receiver'е чтобы fake
// удовлетворял interface. Методы, которые НЕ используются в GrantReward,
// panic'ят при вызове — это сразу покажет регресс если кто-то начнёт их
// дёргать без обновления тестов.

type fakeUserRepo struct {
	users         map[uint]*models.User
	markRewardedF func(userID uint) (bool, error)
	setPlanCalls  []setPlanCall
}

type setPlanCall struct {
	userID uint
	planID string
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[uint]*models.User)}
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uint) (*models.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) SetPlan(_ context.Context, userID uint, planID string) error {
	u, ok := f.users[userID]
	if !ok {
		return repo.ErrNotFound
	}
	u.PlanID = planID
	f.setPlanCalls = append(f.setPlanCalls, setPlanCall{userID: userID, planID: planID})
	return nil
}

func (f *fakeUserRepo) MarkReferralRewarded(_ context.Context, userID uint) (bool, error) {
	if f.markRewardedF != nil {
		return f.markRewardedF(userID)
	}
	u, ok := f.users[userID]
	if !ok {
		return false, repo.ErrNotFound
	}
	if u.ReferralRewardedAt != nil {
		return false, nil
	}
	now := time.Now()
	u.ReferralRewardedAt = &now
	return true, nil
}

// --- unused methods (panic to flag accidental usage) ---

func (f *fakeUserRepo) Create(context.Context, *models.User) error { panic("unused") }
func (f *fakeUserRepo) GetByEmail(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUserRepo) GetByUsername(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUserRepo) SearchUsers(context.Context, string, int) ([]models.User, error) {
	panic("unused")
}
func (f *fakeUserRepo) Update(context.Context, *models.User) error               { panic("unused") }
func (f *fakeUserRepo) SetQuotaWarningSentOn(context.Context, uint, time.Time) error {
	panic("unused")
}
func (f *fakeUserRepo) TouchLastLogin(context.Context, uint) error { panic("unused") }
func (f *fakeUserRepo) ListInactiveForReengagement(context.Context, time.Time, time.Time, int) ([]models.User, error) {
	panic("unused")
}
func (f *fakeUserRepo) MarkReengagementSent(context.Context, uint) error { panic("unused") }
func (f *fakeUserRepo) CountReferredBy(context.Context, string) (int64, error) {
	panic("unused")
}
func (f *fakeUserRepo) GetByReferralCode(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUserRepo) ListPaidUsers(context.Context) ([]uint, error) { panic("unused") }
func (f *fakeUserRepo) SetInsightEmailsEnabled(context.Context, uint, bool) error {
	panic("unused")
}

// ===== fakeSubRepo =====

type updatePeriodCall struct {
	subID  uint
	newEnd time.Time
}

type fakeSubRepo struct {
	activeByUser   map[uint]*models.Subscription
	created        []*models.Subscription
	updatePeriodC  []updatePeriodCall
	getActiveErr   error
	updatePeriodErr error
	createErr      error
}

func newFakeSubRepo() *fakeSubRepo {
	return &fakeSubRepo{activeByUser: make(map[uint]*models.Subscription)}
}

func (f *fakeSubRepo) Create(_ context.Context, sub *models.Subscription) error {
	if f.createErr != nil {
		return f.createErr
	}
	if sub.ID == 0 {
		sub.ID = uint(len(f.created) + 1000)
	}
	f.created = append(f.created, sub)
	// Также положим в active, чтобы тесты на side-effects могли проверить.
	f.activeByUser[sub.UserID] = sub
	return nil
}

func (f *fakeSubRepo) GetActiveByUserID(_ context.Context, userID uint) (*models.Subscription, error) {
	if f.getActiveErr != nil {
		return nil, f.getActiveErr
	}
	s, ok := f.activeByUser[userID]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return s, nil
}

func (f *fakeSubRepo) UpdatePeriodEnd(_ context.Context, subID uint, newEnd time.Time) error {
	if f.updatePeriodErr != nil {
		return f.updatePeriodErr
	}
	f.updatePeriodC = append(f.updatePeriodC, updatePeriodCall{subID: subID, newEnd: newEnd})
	for _, s := range f.activeByUser {
		if s.ID == subID {
			s.CurrentPeriodEnd = newEnd
			break
		}
	}
	return nil
}

// --- unused ---

func (f *fakeSubRepo) Update(context.Context, *models.Subscription) error { panic("unused") }
func (f *fakeSubRepo) ListExpiring(context.Context, time.Time) ([]models.Subscription, error) {
	panic("unused")
}
func (f *fakeSubRepo) ActivateWithPlanUpdate(context.Context, *models.Subscription, uint, string) error {
	panic("unused")
}
func (f *fakeSubRepo) CancelAtPeriodEnd(context.Context, uint) error           { panic("unused") }
func (f *fakeSubRepo) ExpireAndDowngrade(context.Context, uint, uint) error    { panic("unused") }
func (f *fakeSubRepo) GetCurrentByUserID(context.Context, uint) (*models.Subscription, error) {
	panic("unused")
}
func (f *fakeSubRepo) MarkExpired(context.Context, uint) error          { panic("unused") }
func (f *fakeSubRepo) SetRebillId(context.Context, uint, string) error  { panic("unused") }
func (f *fakeSubRepo) SetAutoRenew(context.Context, uint, bool) error   { panic("unused") }
func (f *fakeSubRepo) ListReadyForRenewal(context.Context, time.Time, time.Time, int) ([]models.Subscription, error) {
	panic("unused")
}
func (f *fakeSubRepo) ExtendPeriod(context.Context, uint, time.Time) error { panic("unused") }
func (f *fakeSubRepo) RecordRenewalFailure(context.Context, uint) error    { panic("unused") }
func (f *fakeSubRepo) ListPreExpiring(context.Context, time.Time, time.Time, int16) ([]models.Subscription, error) {
	panic("unused")
}
func (f *fakeSubRepo) SetPreExpireStage(context.Context, uint, int16) error { panic("unused") }
func (f *fakeSubRepo) Pause(context.Context, uint, uint, time.Time, time.Time) error {
	panic("unused")
}
func (f *fakeSubRepo) Resume(context.Context, uint, uint, time.Time, time.Time) error {
	panic("unused")
}
func (f *fakeSubRepo) ListExpiredPauses(context.Context, time.Time) ([]models.Subscription, error) {
	panic("unused")
}
func (f *fakeSubRepo) RecordCancellation(context.Context, *models.SubscriptionCancellation) error {
	panic("unused")
}

// ===== fakePayRepo =====

type fakePayRepo struct {
	payments map[uint]*models.Payment
	getErr   error
}

func newFakePayRepo() *fakePayRepo {
	return &fakePayRepo{payments: make(map[uint]*models.Payment)}
}

func (f *fakePayRepo) GetByID(_ context.Context, id uint) (*models.Payment, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	p, ok := f.payments[id]
	if !ok {
		return nil, repo.ErrNotFound
	}
	return p, nil
}

// --- unused ---

func (f *fakePayRepo) Create(context.Context, *models.Payment) error { panic("unused") }
func (f *fakePayRepo) GetByExternalID(context.Context, string, string) (*models.Payment, error) {
	panic("unused")
}
func (f *fakePayRepo) GetByIdempotencyKey(context.Context, string) (*models.Payment, error) {
	panic("unused")
}
func (f *fakePayRepo) UpdateStatus(context.Context, uint, models.PaymentStatus) error {
	panic("unused")
}
func (f *fakePayRepo) UpdateExternalID(context.Context, uint, string) error { panic("unused") }
func (f *fakePayRepo) TransitionStatus(context.Context, uint, models.PaymentStatus, models.PaymentStatus) (bool, error) {
	panic("unused")
}
func (f *fakePayRepo) LinkSubscription(context.Context, uint, uint) error { panic("unused") }

// ===== fakePendingRepo =====
//
// Раньше panic'ил на всё (для GrantReward тестов pending не дёргался). Теперь
// — функциональный fake, потому что reward_loop_test.go вызывает ListEligible
// и Delete. Поведение прозрачно для существующих тестов: они не трогают rows.

type fakePendingRepo struct {
	rows           []models.ReferralPendingReward
	listEligibleErr error
	deleteErr      error
	deletedIDs     []uint
}

func newFakePendingRepo() *fakePendingRepo {
	return &fakePendingRepo{}
}

func (f *fakePendingRepo) Create(context.Context, *models.ReferralPendingReward) error {
	panic("unused")
}

func (f *fakePendingRepo) ListEligible(_ context.Context, ts time.Time, limit int) ([]models.ReferralPendingReward, error) {
	if f.listEligibleErr != nil {
		return nil, f.listEligibleErr
	}
	out := make([]models.ReferralPendingReward, 0, len(f.rows))
	for _, r := range f.rows {
		if !r.EligibleAt.After(ts) { // r.EligibleAt <= ts
			out = append(out, r)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (f *fakePendingRepo) FindByReferee(context.Context, uint) (*models.ReferralPendingReward, error) {
	panic("unused")
}

func (f *fakePendingRepo) Delete(_ context.Context, id uint) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deletedIDs = append(f.deletedIDs, id)
	// Также вычищаем row из rows, чтобы повторный ListEligible его не вернул.
	kept := f.rows[:0]
	for _, r := range f.rows {
		if r.ID != id {
			kept = append(kept, r)
		}
	}
	f.rows = kept
	return nil
}

// ===== Helpers =====

func newTestService(users *fakeUserRepo, subs *fakeSubRepo, pays *fakePayRepo) *Service {
	svc := NewService(subs, users, pays, &fakePendingRepo{})
	// Зафиксированный момент времени для предсказуемости тестов.
	fixedNow := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	svc.SetNowFn(func() time.Time { return fixedNow })
	return svc
}

// ===== Tests =====

func TestService_GrantReward_ProReferrer_ExtendsPeriodEnd(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	referrer := &models.User{ID: 10, PlanID: "pro"}
	users.users[10] = referrer

	originalEnd := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	subs.activeByUser[10] = &models.Subscription{
		ID:               55,
		UserID:           10,
		PlanID:           "pro",
		Status:           models.SubStatusActive,
		CurrentPeriodEnd: originalEnd,
	}

	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	svc := newTestService(users, subs, pays)
	if err := svc.GrantReward(context.Background(), 10, 20, 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(subs.updatePeriodC) != 1 {
		t.Fatalf("expected 1 UpdatePeriodEnd call, got %d", len(subs.updatePeriodC))
	}
	got := subs.updatePeriodC[0]
	if got.subID != 55 {
		t.Errorf("UpdatePeriodEnd subID = %d, want 55", got.subID)
	}
	wantEnd := originalEnd.Add(30 * 24 * time.Hour)
	if !got.newEnd.Equal(wantEnd) {
		t.Errorf("UpdatePeriodEnd newEnd = %v, want %v", got.newEnd, wantEnd)
	}

	// SetPlan не должен вызываться для Pro — план не меняется.
	if len(users.setPlanCalls) != 0 {
		t.Errorf("SetPlan should NOT be called for Pro referrer, got %v", users.setPlanCalls)
	}
	// referral_rewarded_at должен быть set.
	if referrer.ReferralRewardedAt == nil {
		t.Errorf("ReferralRewardedAt should be set after grant")
	}
	// Не создаём новых подписок.
	if len(subs.created) != 0 {
		t.Errorf("should not Create subscription for Pro referrer, got %d", len(subs.created))
	}
}

func TestService_GrantReward_ProYearlyReferrer_ExtendsPeriodEnd(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "pro_yearly"}
	originalEnd := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	subs.activeByUser[10] = &models.Subscription{
		ID: 56, UserID: 10, PlanID: "pro_yearly",
		Status: models.SubStatusActive, CurrentPeriodEnd: originalEnd,
	}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	svc := newTestService(users, subs, pays)
	if err := svc.GrantReward(context.Background(), 10, 20, 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subs.updatePeriodC) != 1 {
		t.Fatalf("expected 1 UpdatePeriodEnd call")
	}
	if !subs.updatePeriodC[0].newEnd.Equal(originalEnd.Add(30 * 24 * time.Hour)) {
		t.Errorf("expected +30d on pro_yearly subscription")
	}
}

func TestService_GrantReward_FreeReferrer_CreatesSyntheticPro(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "free"}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	svc := newTestService(users, subs, pays)
	if err := svc.GrantReward(context.Background(), 10, 20, 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(subs.created) != 1 {
		t.Fatalf("expected 1 Create call, got %d", len(subs.created))
	}
	created := subs.created[0]
	if created.UserID != 10 {
		t.Errorf("Subscription.UserID = %d, want 10", created.UserID)
	}
	if created.PlanID != "pro" {
		t.Errorf("Subscription.PlanID = %q, want 'pro'", created.PlanID)
	}
	if created.Status != models.SubStatusActive {
		t.Errorf("Subscription.Status = %q, want 'active'", created.Status)
	}
	if created.AutoRenew {
		t.Errorf("Subscription.AutoRenew = true, want false (trial — без авторекуррента)")
	}
	if created.RebillId != "" {
		t.Errorf("Subscription.RebillId = %q, want empty (нет карты)", created.RebillId)
	}
	expectedEnd := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC).Add(30 * 24 * time.Hour)
	if !created.CurrentPeriodEnd.Equal(expectedEnd) {
		t.Errorf("CurrentPeriodEnd = %v, want %v", created.CurrentPeriodEnd, expectedEnd)
	}

	// user.plan_id должен стать pro.
	if len(users.setPlanCalls) != 1 {
		t.Fatalf("expected 1 SetPlan call, got %d", len(users.setPlanCalls))
	}
	if users.setPlanCalls[0].planID != "pro" {
		t.Errorf("SetPlan planID = %q, want 'pro'", users.setPlanCalls[0].planID)
	}
	if users.users[10].PlanID != "pro" {
		t.Errorf("after grant user.PlanID = %q, want 'pro'", users.users[10].PlanID)
	}
	if users.users[10].ReferralRewardedAt == nil {
		t.Errorf("ReferralRewardedAt should be set")
	}
}

func TestService_GrantReward_MaxReferrer_ExtendsButDoesNotDowngrade(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "max"}
	originalEnd := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	subs.activeByUser[10] = &models.Subscription{
		ID: 77, UserID: 10, PlanID: "max",
		Status: models.SubStatusActive, CurrentPeriodEnd: originalEnd,
	}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	svc := newTestService(users, subs, pays)
	if err := svc.GrantReward(context.Background(), 10, 20, 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(subs.updatePeriodC) != 1 {
		t.Fatalf("expected 1 UpdatePeriodEnd call (extend Max period), got %d", len(subs.updatePeriodC))
	}
	if subs.updatePeriodC[0].subID != 77 {
		t.Errorf("UpdatePeriodEnd subID = %d, want 77 (Max-подписка)", subs.updatePeriodC[0].subID)
	}
	wantEnd := originalEnd.Add(30 * 24 * time.Hour)
	if !subs.updatePeriodC[0].newEnd.Equal(wantEnd) {
		t.Errorf("UpdatePeriodEnd newEnd = %v, want %v", subs.updatePeriodC[0].newEnd, wantEnd)
	}

	// Crucial: НЕ должно быть downgrade на Pro!
	if len(users.setPlanCalls) != 0 {
		t.Errorf("Max referrer не должен downgrade'иться: SetPlan calls = %v", users.setPlanCalls)
	}
	if users.users[10].PlanID != "max" {
		t.Errorf("after grant user.PlanID = %q, want 'max' (без downgrade)", users.users[10].PlanID)
	}
	// Не создавать новую trial-подписку — у Max уже есть.
	if len(subs.created) != 0 {
		t.Errorf("should not Create subscription for Max referrer")
	}
}

func TestService_GrantReward_MaxYearlyReferrer_Extends(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "max_yearly"}
	subs.activeByUser[10] = &models.Subscription{
		ID: 78, UserID: 10, PlanID: "max_yearly",
		Status:           models.SubStatusActive,
		CurrentPeriodEnd: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	svc := newTestService(users, subs, pays)
	if err := svc.GrantReward(context.Background(), 10, 20, 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subs.updatePeriodC) != 1 {
		t.Fatalf("expected 1 UpdatePeriodEnd, got %d", len(subs.updatePeriodC))
	}
}

func TestService_GrantReward_Idempotent_AlreadyRewarded(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	rewardedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	users.users[10] = &models.User{ID: 10, PlanID: "pro", ReferralRewardedAt: &rewardedAt}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	svc := newTestService(users, subs, pays)
	err := svc.GrantReward(context.Background(), 10, 20, 7)
	if !errors.Is(err, ErrAlreadyRewarded) {
		t.Fatalf("expected ErrAlreadyRewarded, got %v", err)
	}

	// Не должно быть никаких побочных эффектов.
	if len(subs.updatePeriodC) != 0 {
		t.Errorf("no UpdatePeriodEnd expected, got %d", len(subs.updatePeriodC))
	}
	if len(subs.created) != 0 {
		t.Errorf("no Create expected, got %d", len(subs.created))
	}
}

func TestService_GrantReward_PaymentRefunded(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "pro"}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentRefunded}

	svc := newTestService(users, subs, pays)
	err := svc.GrantReward(context.Background(), 10, 20, 7)
	if !errors.Is(err, ErrPaymentRefunded) {
		t.Fatalf("expected ErrPaymentRefunded, got %v", err)
	}

	if len(subs.updatePeriodC) != 0 {
		t.Errorf("no UpdatePeriodEnd expected on refund")
	}
	if users.users[10].ReferralRewardedAt != nil {
		t.Errorf("ReferralRewardedAt should NOT be set on refund")
	}
}

func TestService_GrantReward_PaymentFailedAlsoBlocks(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "pro"}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentFailed}

	svc := newTestService(users, subs, pays)
	err := svc.GrantReward(context.Background(), 10, 20, 7)
	if !errors.Is(err, ErrPaymentRefunded) {
		t.Fatalf("expected ErrPaymentRefunded on failed payment, got %v", err)
	}
}

func TestService_GrantReward_ReferrerMissing(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	// referrer не в map — GetByID вернёт repo.ErrNotFound.
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	svc := newTestService(users, subs, pays)
	err := svc.GrantReward(context.Background(), 999, 20, 7)
	if !errors.Is(err, ErrReferrerMissing) {
		t.Fatalf("expected ErrReferrerMissing, got %v", err)
	}
}

func TestService_GrantReward_RaceOnMarkRewarded(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "pro"}
	subs.activeByUser[10] = &models.Subscription{
		ID: 55, UserID: 10, PlanID: "pro",
		Status: models.SubStatusActive, CurrentPeriodEnd: time.Now().Add(time.Hour),
	}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	// MarkReferralRewarded симулирует параллельный grant: CAS не прошёл.
	users.markRewardedF = func(_ uint) (bool, error) { return false, nil }

	svc := newTestService(users, subs, pays)
	err := svc.GrantReward(context.Background(), 10, 20, 7)
	if !errors.Is(err, ErrAlreadyRewarded) {
		t.Fatalf("expected ErrAlreadyRewarded on race, got %v", err)
	}
}

func TestService_GrantReward_ProReferrer_NoActiveSubscription_Error(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	// Pro юзер, но активной подписки нет (например, expired между webhook'ом и eligibility).
	users.users[10] = &models.User{ID: 10, PlanID: "pro"}
	pays.payments[7] = &models.Payment{ID: 7, Status: models.PaymentSucceeded}

	svc := newTestService(users, subs, pays)
	err := svc.GrantReward(context.Background(), 10, 20, 7)
	if err == nil {
		t.Fatalf("expected error when Pro referrer has no active subscription")
	}
	if errors.Is(err, ErrAlreadyRewarded) || errors.Is(err, ErrPaymentRefunded) || errors.Is(err, ErrReferrerMissing) {
		t.Fatalf("expected generic 'no active subscription' error, got domain error: %v", err)
	}

	// reward не выдан.
	if users.users[10].ReferralRewardedAt != nil {
		t.Errorf("ReferralRewardedAt should NOT be set on extend failure")
	}
}

func TestService_GrantReward_PaymentNotFound(t *testing.T) {
	users := newFakeUserRepo()
	subs := newFakeSubRepo()
	pays := newFakePayRepo()

	users.users[10] = &models.User{ID: 10, PlanID: "pro"}
	// payments[7] не существует — ErrNotFound.

	svc := newTestService(users, subs, pays)
	err := svc.GrantReward(context.Background(), 10, 20, 7)
	if err == nil {
		t.Fatalf("expected error when payment not found")
	}
}
