package prompt_insights

import (
	"testing"
	"time"
)

func TestPromptInsightRowJSON(t *testing.T) {
	r := PromptInsightRow{PromptID: 42, Title: "X", Uses: 10, UpdatedAt: time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC)}
	if r.Title != "X" {
		t.Fatalf("Title mismatch: %v", r.Title)
	}
}

func TestErrSentinels(t *testing.T) {
	for _, e := range []error{ErrUnknownInsightType, ErrPromptsNotOwned, ErrSamePrompt, ErrProRequired} {
		if e == nil {
			t.Fatalf("expected non-nil sentinel error")
		}
	}
}
