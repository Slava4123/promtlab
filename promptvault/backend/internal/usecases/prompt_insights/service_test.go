package prompt_insights

import (
	"context"
	"errors"
	"testing"
	"time"

	repo "promptvault/internal/interface/repository"
)

type fakeAnalyticsRepo struct {
	unused []repo.PromptUsageRow
}

func (f *fakeAnalyticsRepo) UnusedPrompts(ctx context.Context, userID uint, teamID *uint, before time.Time, limit int) ([]repo.PromptUsageRow, error) {
	return f.unused, nil
}

type fakePlanLookup struct{ plan string }

func (p *fakePlanLookup) InsightsForPlan(planID string) []string {
	switch planID {
	case "max", "max_yearly":
		return []string{"unused_prompts", "possible_duplicates", "trending", "declining", "most_edited", "orphan_tags", "empty_collections"}
	case "pro", "pro_yearly":
		return []string{"unused_prompts", "possible_duplicates"}
	}
	return nil
}

func (p *fakePlanLookup) LookupPlanID(ctx context.Context, userID uint) (string, error) {
	return p.plan, nil
}

func TestListUnusedMaxPlan(t *testing.T) {
	ar := &fakeAnalyticsRepo{
		unused: []repo.PromptUsageRow{
			{PromptID: 1, Title: "A", Uses: 0},
			{PromptID: 2, Title: "B", Uses: 0},
		},
	}
	svc := NewService(ar, nil, &fakePlanLookup{plan: "max"}, time.Now)

	rows, err := svc.ListUnused(context.Background(), 100, nil, 50)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].PromptID != 1 || rows[0].Title != "A" {
		t.Fatalf("row[0] mismatch: %+v", rows[0])
	}
}

func TestListUnusedFreePlanBlocked(t *testing.T) {
	svc := NewService(&fakeAnalyticsRepo{}, nil, &fakePlanLookup{plan: "free"}, time.Now)
	_, err := svc.ListUnused(context.Background(), 100, nil, 50)
	if !errors.Is(err, ErrProRequired) {
		t.Fatalf("expected ErrProRequired, got %v", err)
	}
}
