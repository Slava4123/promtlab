package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- helpers ---

func createTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
	t.Helper()
	u := &models.User{
		Email:        email,
		Name:         "Test User",
		PasswordHash: "irrelevant",
	}
	require.NoError(t, db.Create(u).Error)
	return u
}

func createTestPrompt(t *testing.T, db *gorm.DB, userID uint, teamID *uint, title, content string, usageCount int) *models.Prompt {
	t.Helper()
	p := &models.Prompt{
		UserID:     userID,
		TeamID:     teamID,
		Title:      title,
		Content:    content,
		UsageCount: usageCount,
	}
	require.NoError(t, db.Create(p).Error)
	return p
}

func createTestCollection(t *testing.T, db *gorm.DB, userID uint, teamID *uint, name string) *models.Collection {
	t.Helper()
	c := &models.Collection{
		UserID: userID,
		TeamID: teamID,
		Name:   name,
		Color:  "#8b5cf6",
	}
	require.NoError(t, db.Create(c).Error)
	return c
}

func createTestTeam(t *testing.T, db *gorm.DB, creatorID uint) *models.Team {
	t.Helper()
	team := &models.Team{
		Slug:      "team-" + t.Name(),
		Name:      "Test Team",
		CreatedBy: creatorID,
	}
	require.NoError(t, db.Create(team).Error)
	return team
}

func newBadgeRepoTest(t *testing.T) (repo.BadgeRepository, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	// UserBadge уже должна быть в AutoMigrate списке testhelper, но вызываем
	// ещё раз явно для защиты от рассинхронизации. AutoMigrate идемпотентен.
	require.NoError(t, db.AutoMigrate(&models.UserBadge{}))
	return NewBadgeRepository(db), db
}

// --- Unlock / UnlockedIDs / ListByUser ---

func TestBadgeRepo_Unlock_Insert(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "unlock-insert@test.local")

	err := r.Unlock(ctx, u.ID, "first_prompt")
	require.NoError(t, err)

	var stored models.UserBadge
	require.NoError(t, db.Where("user_id = ? AND badge_id = ?", u.ID, "first_prompt").First(&stored).Error)
	assert.Equal(t, u.ID, stored.UserID)
	assert.Equal(t, "first_prompt", stored.BadgeID)
	assert.False(t, stored.UnlockedAt.IsZero(), "UnlockedAt должен быть выставлен")
	assert.WithinDuration(t, time.Now(), stored.UnlockedAt, 5*time.Second)
}

func TestBadgeRepo_Unlock_Duplicate_ReturnsErrAlreadyUnlocked(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "unlock-dup@test.local")

	require.NoError(t, r.Unlock(ctx, u.ID, "architect"))

	err := r.Unlock(ctx, u.ID, "architect")
	assert.ErrorIs(t, err, repo.ErrBadgeAlreadyUnlocked)

	var count int64
	require.NoError(t, db.Model(&models.UserBadge{}).
		Where("user_id = ? AND badge_id = ?", u.ID, "architect").
		Count(&count).Error)
	assert.Equal(t, int64(1), count, "в БД должна остаться ровно одна запись")
}

func TestBadgeRepo_Unlock_ConcurrentRace(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "unlock-race@test.local")

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make(chan error, goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			results <- r.Unlock(ctx, u.ID, "expert")
		}()
	}
	wg.Wait()
	close(results)

	var ok, duplicate int
	for err := range results {
		switch err {
		case nil:
			ok++
		case repo.ErrBadgeAlreadyUnlocked:
			duplicate++
		default:
			t.Fatalf("unexpected error from Unlock: %v", err)
		}
	}
	assert.Equal(t, 1, ok, "ровно один Unlock должен успеть")
	assert.Equal(t, goroutines-1, duplicate, "остальные — ErrBadgeAlreadyUnlocked")

	var count int64
	require.NoError(t, db.Model(&models.UserBadge{}).
		Where("user_id = ?", u.ID).
		Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestBadgeRepo_UnlockedIDs_Empty(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "unlocked-empty@test.local")

	ids, err := r.UnlockedIDs(ctx, u.ID)
	require.NoError(t, err)
	assert.NotNil(t, ids)
	assert.Empty(t, ids)
}

func TestBadgeRepo_UnlockedIDs_MultipleBadges(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "unlocked-many@test.local")

	for _, b := range []string{"first_prompt", "architect", "collector"} {
		require.NoError(t, r.Unlock(ctx, u.ID, b))
	}

	ids, err := r.UnlockedIDs(ctx, u.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 3)
	assert.Contains(t, ids, "first_prompt")
	assert.Contains(t, ids, "architect")
	assert.Contains(t, ids, "collector")
}

func TestBadgeRepo_ListByUser_OrderByUnlockedAtDesc(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "list-order@test.local")

	require.NoError(t, r.Unlock(ctx, u.ID, "first_prompt"))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, r.Unlock(ctx, u.ID, "architect"))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, r.Unlock(ctx, u.ID, "expert"))

	list, err := r.ListByUser(ctx, u.ID)
	require.NoError(t, err)
	require.Len(t, list, 3)
	assert.Equal(t, "expert", list[0].BadgeID)
	assert.Equal(t, "architect", list[1].BadgeID)
	assert.Equal(t, "first_prompt", list[2].BadgeID)
}

func TestBadgeRepo_DeleteByUserAndBadge(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "delete@test.local")

	require.NoError(t, r.Unlock(ctx, u.ID, "first_prompt"))
	require.NoError(t, r.Unlock(ctx, u.ID, "architect"))

	require.NoError(t, r.DeleteByUserAndBadge(ctx, u.ID, "first_prompt"))

	ids, err := r.UnlockedIDs(ctx, u.ID)
	require.NoError(t, err)
	assert.NotContains(t, ids, "first_prompt")
	assert.Contains(t, ids, "architect")

	// Идемпотентность: повторный delete несуществующего — no-op.
	require.NoError(t, r.DeleteByUserAndBadge(ctx, u.ID, "first_prompt"))
}

// --- aggregation tests ---

func TestBadgeRepo_CountSoloPrompts_FiltersTeam(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "count-solo@test.local")

	team := createTestTeam(t, db, u.ID)

	createTestPrompt(t, db, u.ID, nil, "solo-1", "c", 0)
	createTestPrompt(t, db, u.ID, nil, "solo-2", "c", 0)
	createTestPrompt(t, db, u.ID, nil, "solo-3", "c", 0)
	createTestPrompt(t, db, u.ID, &team.ID, "team-1", "c", 0)
	createTestPrompt(t, db, u.ID, &team.ID, "team-2", "c", 0)

	n, err := r.CountSoloPrompts(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)
}

func TestBadgeRepo_CountTeamPrompts_FiltersSolo(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "count-team@test.local")

	team := createTestTeam(t, db, u.ID)

	createTestPrompt(t, db, u.ID, nil, "solo-1", "c", 0)
	createTestPrompt(t, db, u.ID, nil, "solo-2", "c", 0)
	createTestPrompt(t, db, u.ID, &team.ID, "team-1", "c", 0)
	createTestPrompt(t, db, u.ID, &team.ID, "team-2", "c", 0)
	createTestPrompt(t, db, u.ID, &team.ID, "team-3", "c", 0)

	n, err := r.CountTeamPrompts(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)
}

func TestBadgeRepo_CountAllPrompts_IncludesBoth(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "count-all@test.local")

	team := createTestTeam(t, db, u.ID)

	createTestPrompt(t, db, u.ID, nil, "solo-1", "c", 0)
	createTestPrompt(t, db, u.ID, nil, "solo-2", "c", 0)
	createTestPrompt(t, db, u.ID, &team.ID, "team-1", "c", 0)
	createTestPrompt(t, db, u.ID, &team.ID, "team-2", "c", 0)
	createTestPrompt(t, db, u.ID, &team.ID, "team-3", "c", 0)

	n, err := r.CountAllPrompts(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)
}

func TestBadgeRepo_CountAllPrompts_ExcludesSoftDeleted(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "count-soft@test.local")

	createTestPrompt(t, db, u.ID, nil, "keep-1", "c", 0)
	createTestPrompt(t, db, u.ID, nil, "keep-2", "c", 0)
	deleted := createTestPrompt(t, db, u.ID, nil, "deleted", "c", 0)

	// GORM soft-delete — DeletedAt в модели Prompt устанавливается автоматически.
	require.NoError(t, db.Delete(deleted).Error)

	n, err := r.CountAllPrompts(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)
}

func TestBadgeRepo_CountSoloAndTeamCollections(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "count-collections@test.local")

	team := createTestTeam(t, db, u.ID)

	createTestCollection(t, db, u.ID, nil, "solo-c-1")
	createTestCollection(t, db, u.ID, nil, "solo-c-2")
	createTestCollection(t, db, u.ID, &team.ID, "team-c-1")

	solo, err := r.CountSoloCollections(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), solo)

	teamN, err := r.CountTeamCollections(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), teamN)
}

func TestBadgeRepo_SumUsage(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "sum-usage@test.local")

	// Пустая база → 0 (COALESCE защищает от NULL).
	n, err := r.SumUsage(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)

	createTestPrompt(t, db, u.ID, nil, "p1", "c", 10)
	createTestPrompt(t, db, u.ID, nil, "p2", "c", 15)
	createTestPrompt(t, db, u.ID, nil, "p3", "c", 25)

	n, err = r.SumUsage(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(50), n)
}

func TestBadgeRepo_CountVersionedPrompts_MinVersionsFilter(t *testing.T) {
	r, db := newBadgeRepoTest(t)
	ctx := context.Background()
	u := createTestUser(t, db, "count-versioned@test.local")

	p1 := createTestPrompt(t, db, u.ID, nil, "p1", "c", 0)
	p2 := createTestPrompt(t, db, u.ID, nil, "p2", "c", 0)
	p3 := createTestPrompt(t, db, u.ID, nil, "p3", "c", 0)
	p4 := createTestPrompt(t, db, u.ID, nil, "p4", "c", 0)

	createVersions := func(promptID uint, count int) {
		for i := 1; i <= count; i++ {
			v := &models.PromptVersion{
				PromptID:      promptID,
				VersionNumber: uint(i),
				Title:         "v",
				Content:       "c",
			}
			require.NoError(t, db.Create(v).Error)
		}
	}

	createVersions(p1.ID, 1)
	createVersions(p2.ID, 2)
	createVersions(p3.ID, 3)
	createVersions(p4.ID, 4)

	// minVersions = 3 → 2 (p3, p4).
	n, err := r.CountVersionedPrompts(ctx, u.ID, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)

	// minVersions = 5 → 0.
	n, err = r.CountVersionedPrompts(ctx, u.ID, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)
}
