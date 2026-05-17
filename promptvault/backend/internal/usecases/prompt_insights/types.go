package prompt_insights

import "time"

// PromptInsightRow — row для list-style insight endpoints
// (unused / trending / declining / most-edited).
// UpdatedAt помечен omitzero (Go 1.24+) — omitempty для time.Time не работает,
// потому что struct никогда не равен своему zero "interface" значению.
type PromptInsightRow struct {
	PromptID  uint      `json:"prompt_id"`
	Title     string    `json:"title"`
	Uses      int       `json:"uses"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
}

// DuplicatePair — пара похожих промптов из possible_duplicates SQL.
type DuplicatePair struct {
	PromptA    PromptInsightRow `json:"prompt_a"`
	PromptB    PromptInsightRow `json:"prompt_b"`
	Similarity float64          `json:"similarity"`
}
