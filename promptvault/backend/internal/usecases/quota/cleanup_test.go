package quota

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeCleanupRepo — минимальный mock только под DeleteOldDailyUsage,
// чтобы не тащить полный fakeQuotaRepo из quota_test.go (он использует
// counters для других методов, не нужных здесь).
type fakeCleanupRepo struct {
	receivedDays int
	returnRows   int64
	returnErr    error
	calls        int
}

func (r *fakeCleanupRepo) CountPersonalPrompts(context.Context, uint) (int64, error)     { return 0, nil }
func (r *fakeCleanupRepo) CountPersonalCollections(context.Context, uint) (int64, error) { return 0, nil }
func (r *fakeCleanupRepo) CountPersonalChains(context.Context, uint) (int64, error)      { return 0, nil }
func (r *fakeCleanupRepo) CountTeamPrompts(context.Context, uint) (int64, error)         { return 0, nil }
func (r *fakeCleanupRepo) CountTeamCollections(context.Context, uint) (int64, error)     { return 0, nil }
func (r *fakeCleanupRepo) CountTeamChains(context.Context, uint) (int64, error)          { return 0, nil }
func (r *fakeCleanupRepo) CountTeamsOwned(context.Context, uint) (int64, error)          { return 0, nil }
func (r *fakeCleanupRepo) CountActiveShareLinks(context.Context, uint) (int64, error)    { return 0, nil }
func (r *fakeCleanupRepo) CountTeamMembers(context.Context, uint) (int, error)           { return 0, nil }
func (r *fakeCleanupRepo) GetDailyUsage(context.Context, uint, time.Time, string) (int, error) {
	return 0, nil
}
func (r *fakeCleanupRepo) GetTotalUsage(context.Context, uint, string) (int, error) { return 0, nil }
func (r *fakeCleanupRepo) IncrementDailyUsage(context.Context, uint, time.Time, string) error {
	return nil
}
func (r *fakeCleanupRepo) CountStepsByChain(context.Context, uint) (int64, error) { return 0, nil }
func (r *fakeCleanupRepo) DeleteOldDailyUsage(_ context.Context, days int) (int64, error) {
	r.calls++
	r.receivedDays = days
	return r.returnRows, r.returnErr
}

func TestCleanupLoop_Cleanup_PassesRetention(t *testing.T) {
	repo := &fakeCleanupRepo{returnRows: 42}
	loop := NewCleanupLoop(repo, time.Hour)

	loop.cleanup()

	if repo.calls != 1 {
		t.Fatalf("expected 1 call to DeleteOldDailyUsage, got %d", repo.calls)
	}
	if repo.receivedDays != DefaultRetentionDays {
		t.Errorf("expected retention=%d, got %d", DefaultRetentionDays, repo.receivedDays)
	}
}

func TestCleanupLoop_Cleanup_LogsErrorWithoutPanic(t *testing.T) {
	repo := &fakeCleanupRepo{returnErr: errors.New("simulated db error")}
	loop := NewCleanupLoop(repo, time.Hour)

	// Не должно паниковать; ошибка просто логгируется и cleanup() возвращается.
	// Регрессия-тест: ранее loop падал, если ошибка возвращалась без recover.
	loop.cleanup()

	if repo.calls != 1 {
		t.Errorf("expected DeleteOldDailyUsage called once, got %d", repo.calls)
	}
}
