package badge

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- fakeBadgeRepo — in-memory реализация BadgeRepository для unit-тестов ---
// Проще чем testify/mock: мы проверяем конечное состояние, не expectation chain.
type fakeBadgeRepo struct {
	unlocked []models.UserBadge // запись в порядке Unlock; UnlockedAt — now при insert
	// Статические значения для aggregation-методов. Тесты сами выставляют.
	soloPrompts      int64
	teamPrompts      int64
	allPrompts       int64
	soloCollections  int64
	teamCollections  int64
	totalUsage       int64
	versionedPrompts map[int]int64 // minVersions → count (тесты выставляют заранее)

	// Инъекция ошибок для негативных сценариев.
	unlockedIDsErr error
	unlockErr      error

	// Счётчики вызовов для race-проверок.
	unlockCalls   int
	unlockedCalls int
}

func newFake() *fakeBadgeRepo {
	return &fakeBadgeRepo{
		versionedPrompts: make(map[int]int64),
	}
}

func (f *fakeBadgeRepo) Unlock(_ context.Context, userID uint, badgeID string) error {
	f.unlockCalls++
	if f.unlockErr != nil {
		return f.unlockErr
	}
	for _, b := range f.unlocked {
		if b.UserID == userID && b.BadgeID == badgeID {
			return repo.ErrBadgeAlreadyUnlocked
		}
	}
	f.unlocked = append(f.unlocked, models.UserBadge{
		ID:         uint(len(f.unlocked) + 1),
		UserID:     userID,
		BadgeID:    badgeID,
		UnlockedAt: time.Now(),
	})
	return nil
}

func (f *fakeBadgeRepo) UnlockedIDs(_ context.Context, userID uint) (map[string]struct{}, error) {
	f.unlockedCalls++
	if f.unlockedIDsErr != nil {
		return nil, f.unlockedIDsErr
	}
	out := make(map[string]struct{})
	for _, b := range f.unlocked {
		if b.UserID == userID {
			out[b.BadgeID] = struct{}{}
		}
	}
	return out, nil
}

func (f *fakeBadgeRepo) ListByUser(_ context.Context, userID uint) ([]models.UserBadge, error) {
	var out []models.UserBadge
	for _, b := range f.unlocked {
		if b.UserID == userID {
			out = append(out, b)
		}
	}
	return out, nil
}

func (f *fakeBadgeRepo) DeleteByUserAndBadge(_ context.Context, userID uint, badgeID string) error {
	for i, b := range f.unlocked {
		if b.UserID == userID && b.BadgeID == badgeID {
			f.unlocked = slices.Delete(f.unlocked, i, i+1)
			return nil
		}
	}
	return nil
}

func (f *fakeBadgeRepo) CountSoloPrompts(_ context.Context, _ uint) (int64, error) {
	return f.soloPrompts, nil
}

func (f *fakeBadgeRepo) CountTeamPrompts(_ context.Context, _ uint) (int64, error) {
	return f.teamPrompts, nil
}

func (f *fakeBadgeRepo) CountAllPrompts(_ context.Context, _ uint) (int64, error) {
	return f.allPrompts, nil
}

func (f *fakeBadgeRepo) CountSoloCollections(_ context.Context, _ uint) (int64, error) {
	return f.soloCollections, nil
}

func (f *fakeBadgeRepo) CountTeamCollections(_ context.Context, _ uint) (int64, error) {
	return f.teamCollections, nil
}

func (f *fakeBadgeRepo) SumUsage(_ context.Context, _ uint) (int64, error) {
	return f.totalUsage, nil
}

func (f *fakeBadgeRepo) CountVersionedPrompts(_ context.Context, _ uint, minVersions int) (int64, error) {
	return f.versionedPrompts[minVersions], nil
}

// --- helpers ---

func newTestService(t *testing.T, fake *fakeBadgeRepo) *Service {
	t.Helper()
	svc, err := NewService(fake, nil) // streaks=nil — streak-бейджи не участвуют в unit-тестах
	require.NoError(t, err)
	return svc
}

// --- catalog tests ---

func TestLoadCatalog_Valid(t *testing.T) {
	catalog, err := LoadCatalog()
	require.NoError(t, err)
	require.Len(t, catalog, 11, "каталог должен содержать ровно 11 бейджей")

	// Проверим, что все ожидаемые ID присутствуют.
	ids := make(map[string]struct{})
	for _, b := range catalog {
		ids[b.ID] = struct{}{}
	}
	expected := []string{
		"first_prompt", "architect", "team_player", "team_lead", "prompt_master",
		"collector", "team_librarian", "advanced", "refactorer", "on_fire", "expert",
	}
	for _, id := range expected {
		assert.Contains(t, ids, id, "missing badge %q", id)
	}
}

func TestLoadCatalog_RefactorerHasMinVersions(t *testing.T) {
	catalog, err := LoadCatalog()
	require.NoError(t, err)
	for _, b := range catalog {
		if b.ID == "refactorer" {
			assert.Equal(t, CondVersionedPromptCount, b.Condition.Type)
			assert.Equal(t, 3, b.Condition.MinVersions)
			return
		}
	}
	t.Fatal("refactorer badge not found")
}

// --- evaluate tests: basic flow ---

func TestEvaluate_NoCandidatesForEvent(t *testing.T) {
	fake := newFake()
	svc := newTestService(t, fake)

	// Event без зарегистрированных бейджей (выдуманный тип) — no-op.
	newly := svc.Evaluate(context.Background(), 1, Event{Type: "nonexistent_event"})
	assert.Empty(t, newly)
	assert.Equal(t, 0, fake.unlockedCalls, "UnlockedIDs не должен вызываться, если нет кандидатов")
}

func TestEvaluate_FirstPrompt_UnlocksOnFirst(t *testing.T) {
	fake := newFake()
	fake.soloPrompts = 1
	svc := newTestService(t, fake)

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated})

	var gotIDs []string
	for _, b := range newly {
		gotIDs = append(gotIDs, b.ID)
	}
	assert.Contains(t, gotIDs, "first_prompt")
	assert.NotContains(t, gotIDs, "architect", "10 промптов не должно быть разблокировано на первом")
}

func TestEvaluate_Architect_UnlocksOnTen(t *testing.T) {
	fake := newFake()
	fake.soloPrompts = 10
	svc := newTestService(t, fake)

	// Сначала first_prompt уже разблокирован — симулируем.
	require.NoError(t, fake.Unlock(context.Background(), 1, "first_prompt"))

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated})

	var gotIDs []string
	for _, b := range newly {
		gotIDs = append(gotIDs, b.ID)
	}
	assert.Contains(t, gotIDs, "architect")
	assert.NotContains(t, gotIDs, "first_prompt", "first_prompt уже был разблокирован")
}

func TestEvaluate_AlreadyUnlocked_Skips(t *testing.T) {
	fake := newFake()
	fake.soloPrompts = 1
	svc := newTestService(t, fake)

	require.NoError(t, fake.Unlock(context.Background(), 1, "first_prompt"))

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated})
	assert.Empty(t, newly, "уже разблокированный бейдж не должен попадать в newly")
}

func TestEvaluate_ConditionNotMet(t *testing.T) {
	fake := newFake()
	fake.soloPrompts = 5 // меньше 10 для architect, но ≥ 1 для first_prompt
	svc := newTestService(t, fake)

	require.NoError(t, fake.Unlock(context.Background(), 1, "first_prompt"))

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated})
	assert.Empty(t, newly, "5 < 10, architect не должен быть разблокирован")
}

// --- evaluate tests: error handling / best-effort ---

func TestEvaluate_UnlockedIDsFails_ReturnsNilSilently(t *testing.T) {
	fake := newFake()
	fake.unlockedIDsErr = errors.New("db connection lost")
	fake.soloPrompts = 1
	svc := newTestService(t, fake)

	// Не должен panic, не должен возвращать error — это void method.
	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated})
	assert.Nil(t, newly)
	assert.Equal(t, 0, fake.unlockCalls, "Unlock не должен вызываться при провале UnlockedIDs")
}

func TestEvaluate_UnlockFails_SkipsButContinues(t *testing.T) {
	fake := newFake()
	fake.unlockErr = errors.New("disk full")
	fake.soloPrompts = 1
	svc := newTestService(t, fake)

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated})
	assert.Empty(t, newly, "если Unlock упал — бейдж не в newly, но main flow не ломается")
}

func TestEvaluate_RaceCondition_ErrAlreadyUnlocked_NotInNewly(t *testing.T) {
	fake := newFake()
	// Симулируем гонку: другой вызов Evaluate уже разблокировал first_prompt,
	// но наш ещё не успел получить UnlockedIDs. fake.Unlock вернёт
	// ErrBadgeAlreadyUnlocked, Service должен тихо пропустить.
	fake.unlockErr = repo.ErrBadgeAlreadyUnlocked
	fake.soloPrompts = 1
	svc := newTestService(t, fake)

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated})
	assert.Empty(t, newly)
}

// --- evaluate tests: team badges ---

func TestEvaluate_TeamBadge_NilTeamID_Skipped(t *testing.T) {
	fake := newFake()
	fake.teamPrompts = 1
	svc := newTestService(t, fake)

	// Event без TeamID — team_player не должен быть проверен (short-circuit).
	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated})

	var gotIDs []string
	for _, b := range newly {
		gotIDs = append(gotIDs, b.ID)
	}
	assert.NotContains(t, gotIDs, "team_player")
}

func TestEvaluate_TeamBadge_WithTeamID_Unlocks(t *testing.T) {
	fake := newFake()
	fake.teamPrompts = 1
	svc := newTestService(t, fake)

	teamID := uint(42)
	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated, TeamID: &teamID})

	var gotIDs []string
	for _, b := range newly {
		gotIDs = append(gotIDs, b.ID)
	}
	assert.Contains(t, gotIDs, "team_player")
}

func TestEvaluate_TeamLead_UnlocksOnTen(t *testing.T) {
	fake := newFake()
	fake.teamPrompts = 10
	svc := newTestService(t, fake)

	// team_player уже разблокирован.
	require.NoError(t, fake.Unlock(context.Background(), 1, "team_player"))

	teamID := uint(42)
	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated, TeamID: &teamID})

	var gotIDs []string
	for _, b := range newly {
		gotIDs = append(gotIDs, b.ID)
	}
	assert.Contains(t, gotIDs, "team_lead")
}

// --- evaluate tests: milestone badges ---

func TestEvaluate_PromptMaster_UnlocksOnTwentyFive(t *testing.T) {
	fake := newFake()
	fake.allPrompts = 25
	svc := newTestService(t, fake)

	// Все другие prompt_created бейджи уже разблокированы.
	for _, id := range []string{"first_prompt", "architect"} {
		require.NoError(t, fake.Unlock(context.Background(), 1, id))
	}

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptCreated})

	var gotIDs []string
	for _, b := range newly {
		gotIDs = append(gotIDs, b.ID)
	}
	assert.Contains(t, gotIDs, "prompt_master")
}

func TestEvaluate_Advanced_UnlocksOnFiftyUsage(t *testing.T) {
	fake := newFake()
	fake.totalUsage = 50
	svc := newTestService(t, fake)

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptUsed})

	var gotIDs []string
	for _, b := range newly {
		gotIDs = append(gotIDs, b.ID)
	}
	assert.Contains(t, gotIDs, "advanced")
}

func TestEvaluate_Refactorer_UnlocksOnFiveVersionedPrompts(t *testing.T) {
	fake := newFake()
	fake.versionedPrompts[3] = 5 // 5 промптов с >= 3 версиями
	svc := newTestService(t, fake)

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptUpdated})

	var gotIDs []string
	for _, b := range newly {
		gotIDs = append(gotIDs, b.ID)
	}
	assert.Contains(t, gotIDs, "refactorer")
}

func TestEvaluate_Refactorer_NotYet(t *testing.T) {
	fake := newFake()
	fake.versionedPrompts[3] = 4 // почти
	svc := newTestService(t, fake)

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventPromptUpdated})
	assert.Empty(t, newly)
}

// --- evaluate tests: collection badges ---

func TestEvaluate_Collector_UnlocksOnFiveSoloCollections(t *testing.T) {
	fake := newFake()
	fake.soloCollections = 5
	svc := newTestService(t, fake)

	newly := svc.Evaluate(context.Background(), 1, Event{Type: EventCollectionCreated})

	var gotIDs []string
	for _, b := range newly {
		gotIDs = append(gotIDs, b.ID)
	}
	assert.Contains(t, gotIDs, "collector")
}

// --- List tests ---

func TestList_AllLockedForNewUser(t *testing.T) {
	fake := newFake()
	svc := newTestService(t, fake)

	list, err := svc.List(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, list, 11)

	for _, b := range list {
		assert.False(t, b.Unlocked, "бейдж %q должен быть locked для нового юзера", b.ID)
		assert.Nil(t, b.UnlockedAt)
		assert.Equal(t, int64(0), b.Progress)
		assert.Equal(t, b.Condition.Threshold, b.Target)
	}
}

func TestList_UnlockedBadgeHasFullProgress(t *testing.T) {
	fake := newFake()
	fake.soloPrompts = 1
	svc := newTestService(t, fake)
	require.NoError(t, fake.Unlock(context.Background(), 1, "first_prompt"))

	list, err := svc.List(context.Background(), 1)
	require.NoError(t, err)

	var found BadgeWithState
	for _, b := range list {
		if b.ID == "first_prompt" {
			found = b
			break
		}
	}
	require.NotZero(t, found.ID)
	assert.True(t, found.Unlocked)
	assert.NotNil(t, found.UnlockedAt)
	assert.Equal(t, int64(1), found.Progress)
	assert.Equal(t, int64(1), found.Target)
}

func TestList_LockedBadgeShowsActualProgress(t *testing.T) {
	fake := newFake()
	fake.soloPrompts = 3 // < 10, architect locked, но progress=3
	svc := newTestService(t, fake)

	list, err := svc.List(context.Background(), 1)
	require.NoError(t, err)

	var found BadgeWithState
	for _, b := range list {
		if b.ID == "architect" {
			found = b
			break
		}
	}
	require.NotZero(t, found.ID)
	assert.False(t, found.Unlocked)
	assert.Equal(t, int64(3), found.Progress)
	assert.Equal(t, int64(10), found.Target)
}

func TestList_ProgressClampedToTarget(t *testing.T) {
	fake := newFake()
	fake.soloPrompts = 15 // прогресс > target для architect
	svc := newTestService(t, fake)

	// architect НЕ в unlocked — симулируем состояние drift.
	list, err := svc.List(context.Background(), 1)
	require.NoError(t, err)

	var found BadgeWithState
	for _, b := range list {
		if b.ID == "architect" {
			found = b
			break
		}
	}
	require.NotZero(t, found.ID)
	assert.False(t, found.Unlocked)
	assert.Equal(t, int64(10), found.Progress, "progress должен быть clamped до target")
	assert.Equal(t, int64(10), found.Target)
}

func TestList_PreservesCatalogOrder(t *testing.T) {
	fake := newFake()
	svc := newTestService(t, fake)

	list, err := svc.List(context.Background(), 1)
	require.NoError(t, err)

	// Первый бейдж должен быть first_prompt (как в catalog.json).
	require.NotEmpty(t, list)
	assert.Equal(t, "first_prompt", list[0].ID)
}
