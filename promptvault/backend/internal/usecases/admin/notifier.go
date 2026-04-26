package admin

// TierChangeNotifier — узкий интерфейс для уведомления юзера об admin override
// тарифа. Реализуется *email.Service в production, *fakeNotifier в тестах.
//
// Поле notifier в Service nullable: если nil, отправка email пропускается
// (используется в unit-тестах admin без реального SMTP).
//
// Все ошибки notifier — non-blocking: операция ChangeTier завершается
// успешно даже при ошибке email, ошибка логируется через slog.
type TierChangeNotifier interface {
	SendAdminTierChanged(to, name, oldPlan, newPlan, reason, frontendURL string) error
}
