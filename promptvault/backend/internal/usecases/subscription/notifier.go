package subscription

import "time"

// EmailSender — минимальный контракт для отправки email. Позволяет unit-тестам
// внедрять мок без тяжёлого SMTP-сервиса и держит usecase независимым от пакета email.
type EmailSender interface {
	// SendRenewalFailed — уведомление о неудачной попытке автопродления.
	// attempt/maxAttempts — прогресс retry (текст письма может отличаться).
	// graceUntil — nil если retries ещё будут; задано — если исчерпаны
	// и доступ сохраняется до этой даты (M-9 grace period).
	SendRenewalFailed(to, planName string, attempt, maxAttempts int, endsAt time.Time, graceUntil *time.Time, frontendURL string) error
	SendSubscriptionExpired(to, planName, frontendURL string) error
	// SendPreExpireReminder — pre-expire напоминание за 3/1 день (M-5b) для
	// auto_renew=false подписок. daysLeft — 3 или 1.
	SendPreExpireReminder(to, planName string, daysLeft int, endsAt time.Time, frontendURL string) error
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

func (n *EmailNotifier) NotifyRenewalFailed(to, planName string, attempt, maxAttempts int, endsAt time.Time, graceUntil *time.Time) error {
	return n.Sender.SendRenewalFailed(to, planName, attempt, maxAttempts, endsAt, graceUntil, n.FrontendURL)
}

func (n *EmailNotifier) NotifySubscriptionExpired(to, planName string) error {
	return n.Sender.SendSubscriptionExpired(to, planName, n.FrontendURL)
}

// NotifyPreExpireReminder — pre-expire напоминание (M-5b).
func (n *EmailNotifier) NotifyPreExpireReminder(to, planName string, daysLeft int, endsAt time.Time) error {
	return n.Sender.SendPreExpireReminder(to, planName, daysLeft, endsAt, n.FrontendURL)
}
