// Тесты для RenewalLoop.renewOne.
//
// Контекст: 14 мая 2026 T-Bank отбил Init рекуррентного платежа кодом 309
// «Неверные параметры. {request.validate.expected.receipt}» — терминал
// настроен на обязательную фискализацию 54-ФЗ, а renewOne не передавал
// Receipt (см. subscription.go:152 — там Receipt передаётся, а в renewal.go
// был пропущен). Подписка истекла, юзер задаунгрейжен.
//
// Этот тест-файл фиксирует контракт: при включённой фискализации
// (cfg.ReceiptEnabled=true) renewOne обязан собрать Receipt с email
// юзера и передать в Init — иначе бизнес-сценарий автопродления падает.
package subscription

import (
	"context"
	"testing"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/payment"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- Stubs для repository-интерфейсов ---

// Полные интерфейсы крупные (~30 методов на subs), но happy-path renewOne
// дёргает только перечисленные ниже. Остальные методы embeddedного
// интерфейса вызовут nil-panic — это явный сигнал «тест ходит не туда».

type renewalPlansStub struct {
	repo.PlanRepository
	plan *models.SubscriptionPlan
}

func (s *renewalPlansStub) GetByID(_ context.Context, _ string) (*models.SubscriptionPlan, error) {
	return s.plan, nil
}

type renewalUsersStub struct {
	repo.UserRepository
	user      *models.User
	callCount int
}

func (s *renewalUsersStub) GetByID(_ context.Context, _ uint) (*models.User, error) {
	s.callCount++
	return s.user, nil
}

type renewalPaysStub struct {
	repo.PaymentRepository
	created           *models.Payment
	updatedExternalID string
}

func (s *renewalPaysStub) Create(_ context.Context, p *models.Payment) error {
	p.ID = 999 // имитируем autoincrement
	s.created = p
	return nil
}
func (s *renewalPaysStub) UpdateExternalID(_ context.Context, _ uint, externalID string) error {
	s.updatedExternalID = externalID
	return nil
}
func (s *renewalPaysStub) UpdateStatus(_ context.Context, _ uint, _ models.PaymentStatus) error {
	return nil
}

// renewalSubsStub — subs не должен дергаться в happy-path renewOne. Если дернётся
// (RecordRenewalFailure при ошибке), это сигнал что путь пошёл не туда.
type renewalSubsStub struct {
	repo.SubscriptionRepository
}

// --- Tests ---

// Главный регрешн-тест: Receipt из Init не должен быть nil, и Email должен
// быть взят из User. Без этого T-Bank возвращает 309 (см. инциндент 14.05.2026).
func TestRenewOne_PassesReceiptWithUserEmail(t *testing.T) {
	user := &models.User{ID: 42, Email: "user@example.com"}
	plan := &models.SubscriptionPlan{ID: "pro", Name: "Pro", PriceKop: 59900}
	sub := &models.Subscription{ID: 7, UserID: 42, PlanID: "pro", RebillId: "rb-12345"}

	var capturedInit payment.InitRequest
	prov := &mockProvider{
		initFunc: func(_ context.Context, req payment.InitRequest) (*payment.InitResult, error) {
			capturedInit = req
			return &payment.InitResult{ExternalID: "tbank-ext-1"}, nil
		},
		chargeFunc: func(_ context.Context, _ payment.ChargeRequest) (*payment.ChargeResult, error) {
			return &payment.ChargeResult{ExternalID: "tbank-ext-1", Status: "AUTHORIZING"}, nil
		},
	}

	usersStub := &renewalUsersStub{user: user}
	loop := &RenewalLoop{
		subs:    &renewalSubsStub{},
		plans:   &renewalPlansStub{plan: plan},
		pays:    &renewalPaysStub{},
		users:   usersStub,
		payment: prov,
		cfg: &config.PaymentConfig{
			Enabled:        true,
			ReceiptEnabled: true,
			Taxation:       "usn_income",
			WebhookBaseURL: "https://test.local",
		},
	}

	if err := loop.renewOne(context.Background(), sub); err != nil {
		t.Fatalf("renewOne: unexpected error: %v", err)
	}

	if usersStub.callCount == 0 {
		t.Fatal("renewOne должен загружать юзера через users.GetByID для email Receipt")
	}
	if capturedInit.Receipt == nil {
		t.Fatal("Init получил Receipt=nil — T-Bank вернёт 309 при включённой фискализации")
	}
	if capturedInit.Receipt.Email != user.Email {
		t.Errorf("Receipt.Email = %q, want %q", capturedInit.Receipt.Email, user.Email)
	}
	if capturedInit.Receipt.Taxation != "usn_income" {
		t.Errorf("Receipt.Taxation = %q, want %q", capturedInit.Receipt.Taxation, "usn_income")
	}
	if len(capturedInit.Receipt.Items) != 1 {
		t.Fatalf("Receipt.Items: len = %d, want 1", len(capturedInit.Receipt.Items))
	}
	item := capturedInit.Receipt.Items[0]
	if item.AmountKop != plan.PriceKop {
		t.Errorf("Receipt.Items[0].AmountKop = %d, want %d", item.AmountKop, plan.PriceKop)
	}
	if item.PaymentObject != "service" {
		t.Errorf("Receipt.Items[0].PaymentObject = %q, want %q", item.PaymentObject, "service")
	}
}

// Edge-кейс: фискализация выключена (cfg.ReceiptEnabled=false) — Receipt=nil
// допустим. На таких терминалах T-Bank не требует чек и Init проходит без него.
func TestRenewOne_NoReceiptWhenFiscalizationDisabled(t *testing.T) {
	user := &models.User{ID: 1, Email: "u@x.ru"}
	plan := &models.SubscriptionPlan{ID: "pro", Name: "Pro", PriceKop: 59900}
	sub := &models.Subscription{ID: 1, UserID: 1, PlanID: "pro", RebillId: "rb-1"}

	var capturedInit payment.InitRequest
	prov := &mockProvider{
		initFunc: func(_ context.Context, req payment.InitRequest) (*payment.InitResult, error) {
			capturedInit = req
			return &payment.InitResult{ExternalID: "ext"}, nil
		},
		chargeFunc: func(_ context.Context, _ payment.ChargeRequest) (*payment.ChargeResult, error) {
			return &payment.ChargeResult{ExternalID: "ext", Status: "AUTHORIZING"}, nil
		},
	}

	loop := &RenewalLoop{
		subs:    &renewalSubsStub{},
		plans:   &renewalPlansStub{plan: plan},
		pays:    &renewalPaysStub{},
		users:   &renewalUsersStub{user: user},
		payment: prov,
		cfg: &config.PaymentConfig{
			Enabled:        true,
			ReceiptEnabled: false, // фискализация выключена
			WebhookBaseURL: "https://test.local",
		},
	}

	if err := loop.renewOne(context.Background(), sub); err != nil {
		t.Fatalf("renewOne: unexpected error: %v", err)
	}
	if capturedInit.Receipt != nil {
		t.Errorf("Receipt должен быть nil при ReceiptEnabled=false, получили %+v", capturedInit.Receipt)
	}
}
