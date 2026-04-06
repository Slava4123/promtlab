package service

// EmailSender defines the interface for sending emails.
type EmailSender interface {
	Configured() bool
	SendVerificationCode(to, code string) error
	SendPasswordResetCode(to, code string) error
	SendSetPasswordCode(to, code string) error
	SendPasswordChangedNotification(to string) error
}
