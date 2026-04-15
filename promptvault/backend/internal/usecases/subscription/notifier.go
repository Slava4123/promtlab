package subscription

import "time"

// EmailSender — минимальный контракт для отправки email. Позволяет unit-тестам
// внедрять мок без тяжёлого SMTP-сервиса и держит usecase независимым от пакета email.
type EmailSender interface {
	SendRenewalFailed(to, planName string, endsAt time.Time, frontendURL string) error
	SendSubscriptionExpired(to, planName, frontendURL string) error
}

// EmailNotifier реализует RenewalNotifier и ExpirationNotifier поверх EmailSender.
// FrontendURL подставляется во все уведомления чтобы ссылки работали независимо
// от окружения (dev/prod). Может использоваться nil-значением: в этом случае
// usecase пропускает отправку (адаптер сам обрабатывает nil в app layer).
type EmailNotifier struct {
	Sender      EmailSender
	FrontendURL string
}

func NewEmailNotifier(sender EmailSender, frontendURL string) *EmailNotifier {
	return &EmailNotifier{Sender: sender, FrontendURL: frontendURL}
}

func (n *EmailNotifier) NotifyRenewalFailed(to, planName string, endsAt time.Time) error {
	return n.Sender.SendRenewalFailed(to, planName, endsAt, n.FrontendURL)
}

func (n *EmailNotifier) NotifySubscriptionExpired(to, planName string) error {
	return n.Sender.SendSubscriptionExpired(to, planName, n.FrontendURL)
}
