package quota

import "fmt"

// QuotaExceededError — typed error с метаданными для enriched HTTP 402 response.
type QuotaExceededError struct {
	Message   string `json:"error"`
	QuotaType string `json:"quota_type"`
	Used      int    `json:"used"`
	Limit     int    `json:"limit"`
	PlanID    string `json:"plan"`
}

func (e *QuotaExceededError) Error() string {
	return e.Message
}

func newQuotaExceeded(quotaType, planID string, used, limit int, resource string) *QuotaExceededError {
	return &QuotaExceededError{
		Message:   fmt.Sprintf("Лимит %s исчерпан", resource),
		QuotaType: quotaType,
		Used:      used,
		Limit:     limit,
		PlanID:    planID,
	}
}
