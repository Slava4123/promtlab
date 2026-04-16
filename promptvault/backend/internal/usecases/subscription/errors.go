package subscription

import "errors"

var (
	ErrAlreadySubscribed       = errors.New("у вас уже есть активная подписка")
	ErrNoActiveSubscription    = errors.New("активная подписка не найдена")
	ErrPlanNotFound            = errors.New("тарифный план не найден")
	ErrPaymentNotConfigured    = errors.New("платёжная система не настроена")
	ErrPaymentFailed           = errors.New("ошибка создания платежа")
	ErrInvalidWebhookSignature = errors.New("невалидная подпись webhook")

	// M-6 / M-6b
	ErrSubscriptionNotPausable = errors.New("эту подписку нельзя поставить на паузу")
	ErrSubscriptionPaused      = errors.New("подписка уже на паузе")
	ErrSubscriptionNotPaused   = errors.New("подписка не на паузе")
	ErrInvalidPauseMonths      = errors.New("pause months должно быть 1, 2 или 3")
	ErrInvalidCancelReason     = errors.New("неизвестная причина отмены")
)
