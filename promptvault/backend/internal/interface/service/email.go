package service

// EmailSender defines the interface for sending emails.
type EmailSender interface {
	Configured() bool
	SendVerificationCode(to, code string) error
	SendPasswordResetCode(to, code string) error
	SendSetPasswordCode(to, code string) error
	SendPasswordChangedNotification(to string) error
	SendTeamInvitation(to, teamName, inviterName string) error
	// SendWelcome — приветственное письмо после verify. name может быть пустым
	// (юзер не указал имя) — тогда используется generic greeting.
	SendWelcome(to, name, frontendURL string) error
	// SendQuotaWarning — предупреждение при достижении 80% AI-квоты (M-5c).
	// quotaType: "ai_total" (Free одноразово) или "ai_daily" (Pro/Max сегодня).
	SendQuotaWarning(to, name, quotaType string, used, limit int, frontendURL string) error
	// SendReengagement — letter для юзеров неактивных 14+ дней (M-5d).
	SendReengagement(to, name, frontendURL string) error
}
