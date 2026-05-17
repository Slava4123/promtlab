// backend/internal/usecases/referral/errors.go
package referral

import "errors"

// Доменные ошибки. HTTP-маппинг — в delivery/http/referral/errors.go
// (если фича будет иметь endpoint'ы; на MVP — только background grant).
var (
	// ErrAlreadyRewarded — referrer уже получил награду за этого referee
	// (или раньше за другого — referral_rewarded_at != NULL).
	ErrAlreadyRewarded = errors.New("referral: уже награждено")
	// ErrPaymentRefunded — pending был создан, но payment refunded до eligibility.
	ErrPaymentRefunded = errors.New("referral: payment refunded")
	// ErrRefereeInactive — referee subscription больше не active (cancelled/expired).
	ErrRefereeInactive = errors.New("referral: referee inactive")
	// ErrReferrerMissing — referrer был удалён (ON DELETE CASCADE pending удалит, но edge cases).
	ErrReferrerMissing = errors.New("referral: referrer missing")
)
