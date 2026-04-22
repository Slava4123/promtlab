package app

import "time"

// Lifecycle helpers для фоновых loops. Вынесены из app.go в отдельный файл,
// чтобы основной DI-файл не размывался Start/Stop-boilerplate'ом.

// StartBackground запускает все ежесуточные/периодические loops.
// Идемпотентен по смыслу: каждый loop внутри проверяет ticker + stopCh.
func (a *App) StartBackground() {
	a.purgeLoop.Start()
	a.expirationLoop.Start()
	a.renewalLoop.Start()
	a.reminderLoop.Start()
	a.reengagementLoop.Start()
	a.streakReminderLoop.Start()
	a.activityCleanupLoop.Start()
	a.insightsLoop.Start()
}

// Shutdown останавливает все loops и ждёт фоновых задач auth Service
// до указанного timeout. Вызывается при graceful shutdown сервера.
func (a *App) Shutdown(timeout time.Duration) {
	a.purgeLoop.Stop()
	a.expirationLoop.Stop()
	a.renewalLoop.Stop()
	a.reminderLoop.Stop()
	a.reengagementLoop.Stop()
	a.streakReminderLoop.Stop()
	a.activityCleanupLoop.Stop()
	a.insightsLoop.Stop()
	a.feedbackRL.Close()
	a.authSvc.WaitBackground(timeout)
}
