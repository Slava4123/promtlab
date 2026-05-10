package audit

import "promptvault/internal/models"

// MN-30: типы Action / TargetType переехали в models/audit_log.go вместе
// с самой моделью AuditLog. Здесь оставлены type-aliases для backward
// compatibility callers'ов и константы — единый источник правды для значений.
type (
	Action     = models.AuditAction
	TargetType = models.AuditTargetType
)

const (
	ActionGrantBadge           Action = "grant_badge"
	ActionRevokeBadge          Action = "revoke_badge"
	ActionFreezeUser           Action = "freeze_user"
	ActionUnfreezeUser         Action = "unfreeze_user"
	ActionResetPassword        Action = "reset_password"
	ActionChangeTier           Action = "change_tier"
	ActionPromoteAdmin         Action = "promote_admin"
	ActionDemoteAdmin          Action = "demote_admin"
	ActionEnrollTOTP           Action = "enroll_totp"
	ActionDisableTOTP          Action = "disable_totp"
	ActionRegenBackupCodes     Action = "regen_backup_codes"
	ActionUpdateFeedbackStatus Action = "update_feedback_status"
	ActionDeleteFeedback       Action = "delete_feedback"
)

const (
	TargetUser       TargetType = "user"
	TargetPrompt     TargetType = "prompt"
	TargetCollection TargetType = "collection"
	TargetBadge      TargetType = "badge"
	TargetFeedback   TargetType = "feedback"
)
