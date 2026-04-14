package config

import "fmt"

// PaymentConfig — конфигурация платёжного провайдера T-Bank.
//
// Taxation — система налогообложения для фискальных чеков 54-ФЗ:
//
//	usn_income           — УСН доходы 6% (наиболее частый для SaaS)
//	usn_income_outcome   — УСН доходы минус расходы 15%
//	osn                  — общая система
//	patent               — патент
//	esn                  — ЕСХН
//	envd                 — ЕНВД (отменён с 2021)
//
// Передаётся в Receipt.Taxation. ReceiptEnabled=false полностью отключает
// формирование чека (для случаев когда онлайн-касса ещё не подключена).
type PaymentConfig struct {
	Enabled          bool   `koanf:"enabled"`
	TBankTerminalKey string `koanf:"tbank_terminal_key"`
	TBankPassword    string `koanf:"tbank_password"`
	TBankBaseURL     string `koanf:"tbank_base_url"`
	WebhookBaseURL   string `koanf:"webhook_base_url"`
	SuccessURL       string `koanf:"success_url"`
	FailURL          string `koanf:"fail_url"`
	ReceiptEnabled   bool   `koanf:"receipt_enabled"`
	Taxation         string `koanf:"taxation"`
	// RecurrentEnabled — если true, Checkout передаёт Recurrent=Y + CustomerKey
	// для подключения автопродления. Отключить временно только для прохождения
	// теста 1 «Успешная оплата» в тестовом терминале T-Bank (их проверка падает
	// на Recurrent=Y). В prod всегда true.
	RecurrentEnabled bool `koanf:"recurrent_enabled"`
}

// Validate fail-fast проверяет конфиг. Если Enabled=false — разрешаем любые
// значения (биллинг отключён). Если Enabled=true — все T-Bank ключи и
// webhook URL обязательны, иначе сервис стартовать не должен (иначе Checkout
// вернёт 501 для юзера, что выглядит как сломанный продукт).
func (c PaymentConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.TBankTerminalKey == "" {
		return fmt.Errorf("PAYMENT_ENABLED=true but PAYMENT_TBANK_TERMINAL_KEY is empty")
	}
	if c.TBankPassword == "" {
		return fmt.Errorf("PAYMENT_ENABLED=true but PAYMENT_TBANK_PASSWORD is empty")
	}
	if c.TBankBaseURL == "" {
		return fmt.Errorf("PAYMENT_ENABLED=true but PAYMENT_TBANK_BASE_URL is empty")
	}
	if c.WebhookBaseURL == "" {
		return fmt.Errorf("PAYMENT_ENABLED=true but PAYMENT_WEBHOOK_BASE_URL is empty (T-Bank не сможет доставить webhook)")
	}
	return nil
}
