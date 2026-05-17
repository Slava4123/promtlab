// backend/internal/usecases/referral/types.go
package referral

// RewardSummary — результат одного тика ReferralRewardLoop.
// Используется для observability (logging + metrics).
type RewardSummary struct {
	Granted        int
	SkippedRefund  int // payment.refunded к моменту eligibility
	SkippedActive  int // referee subscription уже не active
	SkippedDeleted int // referrer был удалён
	Errors         int // grant fail'нул по другой причине
}

func (s RewardSummary) Total() int {
	return s.Granted + s.SkippedRefund + s.SkippedActive + s.SkippedDeleted + s.Errors
}

// Длительности reward'а и eligibility window.
const (
	// RewardDays — длительность Pro-периода, которым награждаем пригласившего.
	RewardDays = 30
	// EligibilityDays — задержка между первым платежом реферри и grant'ом.
	// Должна превышать T-Bank refund-окно (14 дней), иначе arbitrage-риск.
	EligibilityDays = 14
)
