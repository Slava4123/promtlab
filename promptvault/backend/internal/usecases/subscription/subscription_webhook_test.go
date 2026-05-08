// CR-8 partial — webhook handler tests. Полное покрытие 15+ сценариев
// (CONFIRMED первый платёж/duplicate/REFUNDED и т.д.) требует объёмные
// mocks (5 репозиториев + provider). Этот файл закрывает критические
// гарантии HandleWebhook без активации/refund подписки:
//
//   1. invalid signature → ErrInvalidWebhookSignature (sentinel, чтобы
//      handler ответил 400 — T-Bank не должен ретраить);
//   2. payment не configured → ErrPaymentNotConfigured;
//   3. payment по external_id не найден → wrapped error;
//   4. terminal state (refunded/failed) → idempotent return nil;
//   5. unknown T-Bank статус → return nil, log info;
//   6. same status (succeeded → succeeded) → no-op return nil;
//   7. transition race (transitioned=false) → return nil без активации.
//
// Полные scenarios (CONFIRMED + activation, REFUNDED + handleRefund,
// duplicate webhook idempotency) — отложено как тех.долг (CR-8 part 2,
// требует SubscriptionRepository + UserRepository + PlanRepository mocks).
package subscription

import (
	"context"
	"errors"
	"testing"

	"promptvault/internal/infrastructure/payment"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- Mocks ---

// mockProvider — payment.PaymentProvider stub.
type mockProvider struct {
	verifyResult bool
	verifyCalls  int
	verifyParams map[string]string
}

func (m *mockProvider) Init(_ context.Context, _ payment.InitRequest) (*payment.InitResult, error) {
	panic("unused in webhook tests")
}
func (m *mockProvider) Charge(_ context.Context, _ payment.ChargeRequest) (*payment.ChargeResult, error) {
	panic("unused in webhook tests")
}
func (m *mockProvider) VerifyWebhookSignature(params map[string]string, _ string) bool {
	m.verifyCalls++
	m.verifyParams = params
	return m.verifyResult
}

// mockPaymentRepo — для webhook тестов нужны GetByExternalID, TransitionStatus.
type mockPaymentRepo struct {
	getByExtID    func(provider, externalID string) (*models.Payment, error)
	transition    func(id uint, expected, next models.PaymentStatus) (bool, error)
	transitionErr error
}

func (m *mockPaymentRepo) Create(context.Context, *models.Payment) error { panic("unused") }
func (m *mockPaymentRepo) GetByExternalID(_ context.Context, provider, externalID string) (*models.Payment, error) {
	return m.getByExtID(provider, externalID)
}
func (m *mockPaymentRepo) GetByIdempotencyKey(context.Context, string) (*models.Payment, error) {
	panic("unused")
}
func (m *mockPaymentRepo) UpdateStatus(context.Context, uint, models.PaymentStatus) error {
	panic("unused")
}
func (m *mockPaymentRepo) UpdateExternalID(context.Context, uint, string) error { panic("unused") }
func (m *mockPaymentRepo) TransitionStatus(_ context.Context, id uint, expected, next models.PaymentStatus) (bool, error) {
	if m.transitionErr != nil {
		return false, m.transitionErr
	}
	if m.transition != nil {
		return m.transition(id, expected, next)
	}
	return false, nil
}
func (m *mockPaymentRepo) LinkSubscription(context.Context, uint, uint) error { panic("unused") }

// newWebhookService собирает Service с минимальным набором mocks для
// webhook-тестов. subs/users/plans = nil — мы не доходим до их вызовов
// в покрытых сценариях.
func newWebhookService(t *testing.T, prov *mockProvider, pays *mockPaymentRepo) *Service {
	t.Helper()
	return &Service{
		pays:    pays,
		payment: prov,
	}
}

// --- Tests ---

func TestHandleWebhook_PaymentNotConfigured(t *testing.T) {
	svc := &Service{} // payment=nil
	err := svc.HandleWebhook(context.Background(), "tbank", map[string]string{"Token": "x"})
	if !errors.Is(err, ErrPaymentNotConfigured) {
		t.Fatalf("expected ErrPaymentNotConfigured, got %v", err)
	}
}

func TestHandleWebhook_InvalidSignature(t *testing.T) {
	prov := &mockProvider{verifyResult: false}
	pays := &mockPaymentRepo{}
	svc := newWebhookService(t, prov, pays)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "abc", "PaymentId": "p1", "Status": "CONFIRMED"})

	if !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Fatalf("expected ErrInvalidWebhookSignature, got %v", err)
	}
	if prov.verifyCalls != 1 {
		t.Fatalf("expected 1 VerifyWebhookSignature call, got %d", prov.verifyCalls)
	}
}

func TestHandleWebhook_PaymentNotFound(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &mockPaymentRepo{
		getByExtID: func(_, _ string) (*models.Payment, error) { return nil, repo.ErrNotFound },
	}
	svc := newWebhookService(t, prov, pays)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "ghost", "Status": "CONFIRMED"})

	if err == nil || !errors.Is(err, repo.ErrNotFound) {
		t.Fatalf("expected wrapped repo.ErrNotFound, got %v", err)
	}
}

func TestHandleWebhook_TerminalState_Refunded_Idempotent(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &mockPaymentRepo{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{ID: 1, Status: models.PaymentRefunded}, nil
		},
	}
	svc := newWebhookService(t, prov, pays)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "REFUNDED"})

	if err != nil {
		t.Fatalf("expected nil (idempotent terminal state), got %v", err)
	}
}

func TestHandleWebhook_TerminalState_Failed_Idempotent(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &mockPaymentRepo{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{ID: 1, Status: models.PaymentFailed}, nil
		},
	}
	svc := newWebhookService(t, prov, pays)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "REJECTED"})

	if err != nil {
		t.Fatalf("expected nil for terminal failed state, got %v", err)
	}
}

func TestHandleWebhook_UnknownStatus_Ignored(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &mockPaymentRepo{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{ID: 1, Status: models.PaymentPending}, nil
		},
	}
	svc := newWebhookService(t, prov, pays)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "WEIRD_NEW_STATUS"})

	if err != nil {
		t.Fatalf("expected nil on unknown status (we don't echo back), got %v", err)
	}
}

func TestHandleWebhook_SameStatus_NoOp(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &mockPaymentRepo{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{ID: 1, Status: models.PaymentSucceeded}, nil
		},
	}
	svc := newWebhookService(t, prov, pays)

	// CONFIRMED → succeeded; уже succeeded → no-op.
	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "CONFIRMED"})

	if err != nil {
		t.Fatalf("expected nil on same status, got %v", err)
	}
}

func TestHandleWebhook_TransitionRace_NoOp(t *testing.T) {
	// Два concurrent webhook'а: первый прошёл TransitionStatus (transitioned=true),
	// второй получил transitioned=false и НЕ должен идти к activate/refund.
	prov := &mockProvider{verifyResult: true}
	pays := &mockPaymentRepo{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{ID: 1, Status: models.PaymentPending}, nil
		},
		transition: func(_ uint, _, _ models.PaymentStatus) (bool, error) {
			return false, nil // race-loser
		},
	}
	svc := newWebhookService(t, prov, pays)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "CONFIRMED"})

	if err != nil {
		t.Fatalf("expected nil on transition race-loss, got %v", err)
	}
}

func TestHandleWebhook_TransitionDBError_Wrapped(t *testing.T) {
	prov := &mockProvider{verifyResult: true}
	pays := &mockPaymentRepo{
		getByExtID: func(_, _ string) (*models.Payment, error) {
			return &models.Payment{ID: 1, Status: models.PaymentPending}, nil
		},
		transitionErr: errors.New("connection refused"),
	}
	svc := newWebhookService(t, prov, pays)

	err := svc.HandleWebhook(context.Background(), "tbank",
		map[string]string{"Token": "x", "PaymentId": "p1", "Status": "CONFIRMED"})

	if err == nil {
		t.Fatalf("expected error on DB failure, got nil")
	}
}
