package tbank

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"promptvault/internal/infrastructure/payment"
)

// httpTimeout — общий timeout на запросы к T-Bank API.
// T-Bank обычно отвечает за 1-3 сек; 30 сек — запас на сетевые задержки.
const httpTimeout = 30 * time.Second

// skipSignatureKeys — поля, которые T-Bank не включает в расчёт подписи webhook:
// Token (сама подпись), Receipt и DATA (вложенные объекты/массивы в JSON).
// Также защищает от вредоносных полей-объектов в map[string]string.
var skipSignatureKeys = map[string]struct{}{
	"Token":   {},
	"Receipt": {},
	"DATA":    {},
}

// Config — конфигурация T-Bank платёжного терминала.
type Config struct {
	TerminalKey string
	Password    string
	BaseURL     string // например https://securepay.tinkoff.ru/v2
}

// Provider — реализация PaymentProvider для T-Bank (Tinkoff Acquiring API v2).
type Provider struct {
	cfg    Config
	client *http.Client
}

func NewProvider(cfg Config) *Provider {
	return &Provider{
		cfg:    cfg,
		client: &http.Client{Timeout: httpTimeout},
	}
}

// initRequest — JSON body для POST /Init.
// Recurrent="Y" + CustomerKey запускают рекуррентный flow: T-Bank в webhook
// первого CONFIRMED платежа вернёт RebillId, который нужно сохранить и потом
// использовать в Charge для безакцептных списаний.
type initRequest struct {
	TerminalKey     string         `json:"TerminalKey"`
	Amount          int            `json:"Amount"`
	OrderID         string         `json:"OrderId"`
	Description     string         `json:"Description"`
	Token           string         `json:"Token"`
	SuccessURL      string         `json:"SuccessURL,omitempty"`
	FailURL         string         `json:"FailURL,omitempty"`
	NotificationURL string         `json:"NotificationURL,omitempty"`
	Recurrent       string         `json:"Recurrent,omitempty"`
	CustomerKey     string         `json:"CustomerKey,omitempty"`
	Receipt         *receiptDTO    `json:"Receipt,omitempty"`
}

// receiptDTO — фискальный чек 54-ФЗ в формате T-Bank.
type receiptDTO struct {
	Email    string            `json:"Email,omitempty"`
	Phone    string            `json:"Phone,omitempty"`
	Taxation string            `json:"Taxation"`
	Items    []receiptItemDTO  `json:"Items"`
}

type receiptItemDTO struct {
	Name          string `json:"Name"`
	Price         int    `json:"Price"`
	Quantity      int    `json:"Quantity"`
	Amount        int    `json:"Amount"`
	Tax           string `json:"Tax"`
	PaymentMethod string `json:"PaymentMethod,omitempty"`
	PaymentObject string `json:"PaymentObject,omitempty"`
}

// toReceiptDTO конвертирует доменный Receipt в формат T-Bank.
func toReceiptDTO(r *payment.Receipt) *receiptDTO {
	if r == nil {
		return nil
	}
	items := make([]receiptItemDTO, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, receiptItemDTO{
			Name:          it.Name,
			Price:         it.PriceKop,
			Quantity:      it.Quantity,
			Amount:        it.AmountKop,
			Tax:           it.Tax,
			PaymentMethod: it.PaymentMethod,
			PaymentObject: it.PaymentObject,
		})
	}
	return &receiptDTO{
		Email:    r.Email,
		Phone:    r.Phone,
		Taxation: r.Taxation,
		Items:    items,
	}
}

// generateToken создаёт Token (SHA-256 подпись) для T-Bank API запроса.
func (p *Provider) generateToken(params map[string]string) string {
	merged := make(map[string]string, len(params)+1)
	for k, v := range params {
		merged[k] = v
	}
	merged["Password"] = p.cfg.Password

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(merged[k])
	}

	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

// initResponse — ответ от POST /Init.
type initResponse struct {
	Success    bool   `json:"Success"`
	ErrorCode  string `json:"ErrorCode"`
	Message    string `json:"Message"`
	Details    string `json:"Details"`
	PaymentID  string `json:"PaymentId"`
	PaymentURL string `json:"PaymentURL"`
}

// Init инициализирует платёж через T-Bank API.
func (p *Provider) Init(ctx context.Context, req payment.InitRequest) (*payment.InitResult, error) {
	// Генерируем Token для подписи — все поля запроса + Password
	tokenParams := map[string]string{
		"TerminalKey": p.cfg.TerminalKey,
		"Amount":      fmt.Sprintf("%d", req.Amount),
		"OrderId":     req.OrderID,
		"Description": req.Description,
	}
	if req.SuccessURL != "" {
		tokenParams["SuccessURL"] = req.SuccessURL
	}
	// FailURL не передаём — T-Bank покажет свой экран ошибки с кнопкой "Повторить".
	// Юзер сам вернётся в приложение.
	if req.WebhookURL != "" {
		tokenParams["NotificationURL"] = req.WebhookURL
	}
	if req.Recurrent {
		tokenParams["Recurrent"] = "Y"
	}
	if req.CustomerKey != "" {
		tokenParams["CustomerKey"] = req.CustomerKey
	}
	token := p.generateToken(tokenParams)

	body := initRequest{
		TerminalKey:     p.cfg.TerminalKey,
		Amount:          req.Amount,
		OrderID:         req.OrderID,
		Description:     req.Description,
		Token:           token,
		SuccessURL:      req.SuccessURL,
		NotificationURL: req.WebhookURL,
		CustomerKey:     req.CustomerKey,
		Receipt:         toReceiptDTO(req.Receipt),
	}
	if req.Recurrent {
		body.Recurrent = "Y"
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("tbank: не удалось сериализовать запрос: %w", err)
	}

	slog.Info("tbank.init.request", "order_id", req.OrderID, "amount", req.Amount)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.BaseURL+"/Init", bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("tbank: не удалось создать HTTP-запрос: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tbank: ошибка HTTP-запроса: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tbank: не удалось прочитать ответ: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("tbank.init.http_error", "status", resp.StatusCode, "body", string(respBody))
		return nil, fmt.Errorf("tbank: HTTP %d", resp.StatusCode)
	}

	var result initResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("tbank: не удалось десериализовать ответ: %w", err)
	}

	if !result.Success {
		slog.Error("tbank.init.failed", "error_code", result.ErrorCode, "message", result.Message, "details", result.Details)
		return nil, fmt.Errorf("tbank: ошибка %s — %s", result.ErrorCode, result.Message)
	}

	return &payment.InitResult{
		PaymentURL: result.PaymentURL,
		ExternalID: result.PaymentID,
	}, nil
}

// chargeRequest — JSON body для POST /Charge.
type chargeRequest struct {
	TerminalKey string `json:"TerminalKey"`
	PaymentID   string `json:"PaymentId"`
	RebillID    string `json:"RebillId"`
	Token       string `json:"Token"`
}

// chargeResponse — ответ от POST /Charge. Status — терминальный или промежуточный.
type chargeResponse struct {
	Success   bool   `json:"Success"`
	ErrorCode string `json:"ErrorCode"`
	Message   string `json:"Message"`
	Details   string `json:"Details"`
	PaymentID string `json:"PaymentId"`
	Status    string `json:"Status"`
}

// Charge выполняет безакцептное списание по ранее сохранённому RebillId.
// Используется для автопродления подписки: PaymentId — свежий Init без 3DS,
// RebillId — из webhook первого Recurrent=Y платежа.
func (p *Provider) Charge(ctx context.Context, req payment.ChargeRequest) (*payment.ChargeResult, error) {
	tokenParams := map[string]string{
		"TerminalKey": p.cfg.TerminalKey,
		"PaymentId":   req.PaymentID,
		"RebillId":    req.RebillID,
	}
	token := p.generateToken(tokenParams)

	body := chargeRequest{
		TerminalKey: p.cfg.TerminalKey,
		PaymentID:   req.PaymentID,
		RebillID:    req.RebillID,
		Token:       token,
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("tbank.charge: marshal: %w", err)
	}

	slog.Info("tbank.charge.request", "payment_id", req.PaymentID, "rebill_id", req.RebillID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.BaseURL+"/Charge", bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("tbank.charge: new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tbank.charge: http: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tbank.charge: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("tbank.charge.http_error", "status", resp.StatusCode, "body", string(respBody))
		return nil, fmt.Errorf("tbank.charge: HTTP %d", resp.StatusCode)
	}

	var result chargeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("tbank.charge: unmarshal: %w", err)
	}

	if !result.Success {
		slog.Error("tbank.charge.failed", "error_code", result.ErrorCode, "message", result.Message, "details", result.Details)
		return nil, fmt.Errorf("tbank.charge: %s — %s", result.ErrorCode, result.Message)
	}

	return &payment.ChargeResult{
		ExternalID: result.PaymentID,
		Status:     result.Status,
	}, nil
}

// VerifyWebhookSignature проверяет подпись webhook-уведомления от T-Bank.
//
// Алгоритм:
//  1. Исключить Token, Receipt и DATA (T-Bank не включает их в подпись)
//  2. Добавить Password под ключом "Password"
//  3. Отсортировать пары по ключу (UTF-8 по возрастанию)
//  4. Конкатенировать значения (без разделителей)
//  5. SHA-256 от результата, сравнить hex с переданным token
//
// Значения должны приходить уже в строковом формате T-Bank: bool → "true"/"false",
// числа → десятичная запись без пробелов. Объекты/массивы не должны попадать сюда
// (фильтруются на уровне HTTP-handler'а).
func (p *Provider) VerifyWebhookSignature(params map[string]string, token string) bool {
	merged := make(map[string]string, len(params)+1)
	for k, v := range params {
		if _, skip := skipSignatureKeys[k]; skip {
			continue
		}
		merged[k] = v
	}
	merged["Password"] = p.cfg.Password

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(merged[k])
	}

	hash := sha256.Sum256([]byte(sb.String()))
	computed := hex.EncodeToString(hash[:])

	return strings.EqualFold(computed, token)
}
