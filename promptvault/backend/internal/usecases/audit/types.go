package audit

// Action — фиксированный список типов административных действий.
// Используется как значение audit_log.action. Новые actions добавляются здесь,
// чтобы избежать опечаток и облегчить full-text поиск по коду.
type Action string

const (
	ActionGrantBadge     Action = "grant_badge"
	ActionRevokeBadge    Action = "revoke_badge"
	ActionFreezeUser     Action = "freeze_user"
	ActionUnfreezeUser   Action = "unfreeze_user"
	ActionResetPassword  Action = "reset_password"
	ActionChangeTier     Action = "change_tier"
	ActionPromoteAdmin   Action = "promote_admin"
	ActionDemoteAdmin    Action = "demote_admin"
	ActionEnrollTOTP     Action = "enroll_totp"
	ActionDisableTOTP    Action = "disable_totp"
	ActionRegenBackupCodes Action = "regen_backup_codes"
)

// TargetType — тип сущности над которой совершается админ-действие.
type TargetType string

const (
	TargetUser       TargetType = "user"
	TargetPrompt     TargetType = "prompt"
	TargetCollection TargetType = "collection"
	TargetBadge      TargetType = "badge"
)
