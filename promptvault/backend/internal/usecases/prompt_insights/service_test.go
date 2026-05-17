package prompt_insights

import (
	"context"
	"errors"
	"testing"
	"time"

	repo "promptvault/internal/interface/repository"
)

type fakeAnalyticsRepo struct {
	unused     []repo.PromptUsageRow
	duplicates []repo.DuplicatePair
}

func (f *fakeAnalyticsRepo) UnusedPrompts(ctx context.Context, userID uint, teamID *uint, before time.Time, limit int) ([]repo.PromptUsageRow, error) {
	return f.unused, nil
}

func (f *fakeAnalyticsRepo) PossibleDuplicates(ctx context.Context, userID uint, teamID *uint, threshold float32, limit int) ([]repo.DuplicatePair, error) {
	return f.duplicates, nil
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

func TestListDuplicatesProTeaser(t *testing.T) {
	ar := &fakeAnalyticsRepo{
		duplicates: []repo.DuplicatePair{
			{PromptAID: 1, PromptATitle: "A1", PromptBID: 2, PromptBTitle: "A2", Similarity: 0.91},
		},
	}
	svc := NewService(ar, nil, &fakePlanLookup{plan: "pro"}, time.Now)
	pairs, err := svc.ListDuplicates(context.Background(), 100, nil, 20)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].PromptA.PromptID != 1 || pairs[0].PromptB.PromptID != 2 {
		t.Fatalf("pair mismatch: %+v", pairs[0])
	}
	if pairs[0].Similarity < 0.9 || pairs[0].Similarity > 0.95 {
		t.Fatalf("similarity mismatch: %v", pairs[0].Similarity)
	}
}
