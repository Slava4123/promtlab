package admin

import "time"

// FreshTOTPTTL — окно свежести TOTP verification для destructive actions.
// В течение этого времени после успешного /api/auth/verify-totp юзер может
// выполнять destructive admin actions без повторной verification.
// 12 часов = компромисс между UX (не дёргать TOTP на каждое действие) и
// security (если JWT украли — через 12ч re-verification заблокирует дальнейшее).
const FreshTOTPTTL = 12 * time.Hour
