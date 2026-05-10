package subscription

import (
	"os"
	"time"
)

// addPlanPeriod — t + N "дней" подписки. Базовый расчёт для всех мест где
// проставляется CurrentPeriodEnd (initial activation + renewal extend).
//
// Dev-only override: SERVER_ENVIRONMENT=development И BILLING_FAST_DEV=true →
// каждый "день" подписки превращается в МИНУТУ. Используется для QA-проверки
// flow expire/renewal без ожидания суток (вместе с CRON_FAST_DEV для loops).
//
// Двойная защита (env + IsDev) предотвращает случайное срабатывание в prod
// при ошибке выкатки конфигов.
func addPlanPeriod(t time.Time, days int) time.Time {
	if os.Getenv("SERVER_ENVIRONMENT") == "development" && os.Getenv("BILLING_FAST_DEV") == "true" {
		return t.Add(time.Duration(days) * time.Minute)
	}
	return t.AddDate(0, 0, days)
}
