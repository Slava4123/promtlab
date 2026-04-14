package subscription

import "errors"

var (
	ErrAlreadySubscribed       = errors.New("у вас уже есть активная подписка")
	ErrNoActiveSubscription    = errors.New("активная подписка не найдена")
	ErrPlanNotFound            = errors.New("тарифный план не найден")
	ErrPaymentNotConfigured    = errors.New("платёжная система не настроена")
	ErrPaymentFailed           = errors.New("ошибка создания платежа")
	ErrInvalidWebhookSignature = errors.New("невалидная подпись webhook")
)
