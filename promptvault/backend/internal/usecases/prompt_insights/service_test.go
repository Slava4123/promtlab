package prompt_insights

import (
	"context"
	"errors"
	"testing"
	"time"

	repo "promptvault/internal/interface/repository"

	"gorm.io/gorm"
)

type fakeAnalyticsRepo struct {
	unused     []repo.PromptUsageRow
	duplicates []repo.DuplicatePair
	trending   []repo.TrendRow
	declining  []repo.TrendRow
	mostEdited []repo.PromptUsageRow
}

func (f *fakeAnalyticsRepo) UnusedPrompts(ctx context.Context, userID uint, teamID *uint, before time.Time, limit int) ([]repo.PromptUsageRow, error) {
	return f.unused, nil
}

func (f *fakeAnalyticsRepo) PossibleDuplicates(ctx context.Context, userID uint, teamID *uint, threshold float32, limit int) ([]repo.DuplicatePair, error) {
	return f.duplicates, nil
}

func (f *fakeAnalyticsRepo) GetTrendingPrompts(ctx context.Context, userID uint, teamID *uint, factor float64, growing bool, limit int) ([]repo.TrendRow, error) {
	if growing {
		return f.trending, nil
	}
	return f.declining, nil
}

func (f *fakeAnalyticsRepo) MostEditedPrompts(ctx context.Context, userID uint, teamID *uint, limit int) ([]repo.PromptUsageRow, error) {
	return f.mostEdited, nil
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

func TestListTrending(t *testing.T) {
	ar := &fakeAnalyticsRepo{
		trending: []repo.TrendRow{
			{PromptID: 5, Title: "Hot", UsesLast: 20, UsesPrevious: 5},
		},
	}
	svc := NewService(ar, nil, &fakePlanLookup{plan: "max"}, time.Now)
	rows, err := svc.ListTrending(context.Background(), 100, nil, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(rows) != 1 || rows[0].Uses != 20 {
		t.Fatalf("expected 1 row with uses=20 (UsesLast), got %+v", rows)
	}
}

func TestListDeclining(t *testing.T) {
	ar := &fakeAnalyticsRepo{
		declining: []repo.TrendRow{
			{PromptID: 7, Title: "Falling", UsesLast: 2, UsesPrevious: 18},
		},
	}
	svc := NewService(ar, nil, &fakePlanLookup{plan: "max"}, time.Now)
	rows, err := svc.ListDeclining(context.Background(), 100, nil, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(rows) != 1 || rows[0].Uses != 2 {
		t.Fatalf("expected 1 row with uses=2, got %+v", rows)
	}
}

func TestListMostEdited(t *testing.T) {
	ar := &fakeAnalyticsRepo{
		mostEdited: []repo.PromptUsageRow{
			{PromptID: 8, Title: "Churn", Uses: 15},
		},
	}
	svc := NewService(ar, nil, &fakePlanLookup{plan: "max"}, time.Now)
	rows, err := svc.ListMostEdited(context.Background(), 100, nil, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(rows) != 1 || rows[0].Uses != 15 {
		t.Fatalf("expected 1 row, got %+v", rows)
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

type fakePromptMerger struct {
	called  bool
	keepID  uint
	mergeID uint
	userID  uint
	returns error
}

func (f *fakePromptMerger) MergeWith(ctx context.Context, keepID, mergeID, userID uint) error {
	f.called = true
	f.keepID = keepID
	f.mergeID = mergeID
	f.userID = userID
	return f.returns
}

func TestMergePromptsHappy(t *testing.T) {
	m := &fakePromptMerger{}
	svc := NewService(&fakeAnalyticsRepo{}, m, &fakePlanLookup{plan: "max"}, time.Now)
	err := svc.MergePrompts(context.Background(), 100, 1, 2)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !m.called || m.keepID != 1 || m.mergeID != 2 || m.userID != 100 {
		t.Fatalf("merger not called correctly: %+v", m)
	}
}

func TestMergePromptsSameID(t *testing.T) {
	svc := NewService(&fakeAnalyticsRepo{}, &fakePromptMerger{}, &fakePlanLookup{plan: "max"}, time.Now)
	err := svc.MergePrompts(context.Background(), 100, 5, 5)
	if !errors.Is(err, ErrSamePrompt) {
		t.Fatalf("expected ErrSamePrompt, got %v", err)
	}
}

func TestMergePromptsNotOwned(t *testing.T) {
	m := &fakePromptMerger{returns: gorm.ErrRecordNotFound}
	svc := NewService(&fakeAnalyticsRepo{}, m, &fakePlanLookup{plan: "max"}, time.Now)
	err := svc.MergePrompts(context.Background(), 100, 1, 2)
	if !errors.Is(err, ErrPromptsNotOwned) {
		t.Fatalf("expected ErrPromptsNotOwned, got %v", err)
	}
}
