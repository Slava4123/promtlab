package prompt_insights

import "time"

// PromptInsightRow — row для list-style insight endpoints
// (unused / trending / declining / most-edited).
// Поле UpdatedAt опционально (omitempty) — некоторые SQL возвращают только
// PromptID/Title/Uses без timestamp.
type PromptInsightRow struct {
	PromptID  uint      `json:"prompt_id"`
	Title     string    `json:"title"`
	Uses      int       `json:"uses"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// DuplicatePair — пара похожих промптов из possible_duplicates SQL.
type DuplicatePair struct {
	PromptA    PromptInsightRow `json:"prompt_a"`
	PromptB    PromptInsightRow `json:"prompt_b"`
	Similarity float64          `json:"similarity"`
}
