package payment

import "context"

// InitRequest — параметры инициализации платежа.
//
// Recurrent + CustomerKey используются для подключения автопродления:
// при Recurrent=true T-Bank вернёт RebillId в webhook первого успешного
// платежа; этот RebillId потом используется в Charge для безакцептных
// списаний (без 3DS) на последующие периоды подписки.
//
// Receipt — чек по 54-ФЗ. Если задан, T-Bank сам сформирует фискальный чек
// через «Чеки Т-Бизнеса» / подключённую кассу и передаст в ОФД. Receipt
// НЕ участвует в подписи Token (исключается на стороне T-Bank).
type InitRequest struct {
	OrderID     string
	Amount      int // копейки
	Description string
	SuccessURL  string
	FailURL     string
	WebhookURL  string
	Recurrent   bool   // true — попросить T-Bank выдать RebillId для будущих Charge
	CustomerKey string // обязателен при Recurrent=true; обычно user_id магазина
	Receipt     *Receipt
}

// Receipt — фискальный чек 54-ФЗ. Email или Phone обязательны (хотя бы один).
type Receipt struct {
	Email    string        // email клиента — куда придёт чек
	Phone    string        // альтернатива email; формат +79991234567
	Taxation string        // система налогообложения: usn_income | usn_income_outcome | osn | patent | esn | envd
	Items    []ReceiptItem // позиции чека (для подписки — обычно одна)
}

// ReceiptItem — одна позиция чека.
type ReceiptItem struct {
	Name          string // наименование, ≤128 символов
	PriceKop      int    // цена за единицу в копейках
	Quantity      int    // количество (для услуг обычно 1)
	AmountKop     int    // итого по позиции = Price * Quantity
	Tax           string // ставка НДС: none | vat0 | vat10 | vat20 | vat110 | vat120
	PaymentMethod string // способ расчёта: full_payment | full_prepayment | prepayment | credit | ...
	PaymentObject string // предмет расчёта: service | commodity | excise | ...
}

// InitResult — результат инициализации платежа от провайдера.
type InitResult struct {
	PaymentURL string
	ExternalID string
}

// ChargeRequest — параметры безакцептного списания (рекуррент).
// PaymentID — ID платежа, созданного через Init с CustomerKey (без Recurrent),
// RebillID — взятый из webhook первого платежа с Recurrent=true.
type ChargeRequest struct {
	PaymentID string
	RebillID  string
}

// ChargeResult — результат Charge. Терминальный статус приходит асинхронно
// через webhook; здесь возвращается то, что T-Bank ответил синхронно.
type ChargeResult struct {
	ExternalID string
	Status     string // например "CONFIRMED" / "REJECTED" / "AUTHORIZED"
}

// PaymentProvider — абстракция платёжного провайдера.
// Реализации: tbank.Provider.
type PaymentProvider interface {
	Init(ctx context.Context, req InitRequest) (*InitResult, error)
	Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error)
	VerifyWebhookSignature(params map[string]string, token string) bool
}
