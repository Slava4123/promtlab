package analytics

import (
	"slices"
	"testing"
)

func TestInsightsForPlanPublic(t *testing.T) {
	s := &Service{proInsightsTeaserEnabled: true}
	free := s.InsightsForPlan("free")
	pro := s.InsightsForPlan("pro")
	max := s.InsightsForPlan("max")

	if len(free) != 0 {
		t.Fatalf("Free should get [], got %v", free)
	}
	if !slices.Equal(pro, []string{"unused_prompts", "possible_duplicates"}) {
		t.Fatalf("Pro teaser: got %v", pro)
	}
	if len(max) != 7 {
		t.Fatalf("Max should get 7 types, got %d", len(max))
	}
}
