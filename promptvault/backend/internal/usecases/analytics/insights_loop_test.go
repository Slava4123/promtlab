package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// fakeUsersForLoop — минимальный fake UserRepository, нужен ListPaidUsers +
// GetByID (Task 7: per-plan dispatch).
type fakeUsersForLoop struct {
	ids []uint
	err error
	// plans — мапа uid → plan_id для GetByID. По умолчанию (если nil) — "max"
	// (тесты до Task 7 покрывали Max-only loop, сохраняем backward-compat).
	plans     map[uint]string
	getByErr  error
	getByMiss bool // если true — возвращаем repo.ErrNotFound
}

func (f *fakeUsersForLoop) ListPaidUsers(_ context.Context) ([]uint, error) {
	return f.ids, f.err
}

func (f *fakeUsersForLoop) GetByID(_ context.Context, uid uint) (*models.User, error) {
	if f.getByMiss {
		return nil, repo.ErrNotFound
	}
	if f.getByErr != nil {
		return nil, f.getByErr
	}
	plan := "max"
	if f.plans != nil {
		if p, ok := f.plans[uid]; ok {
			plan = p
		}
	}
	return &models.User{ID: uid, PlanID: plan}, nil
}

// Остальные методы UserRepository не используются loop'ом — паника.
func (f *fakeUsersForLoop) Create(context.Context, *models.User) error { panic("unused") }
func (f *fakeUsersForLoop) GetByEmail(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) GetByUsername(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) SearchUsers(context.Context, string, int) ([]models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) Update(context.Context, *models.User) error  { panic("unused") }
func (f *fakeUsersForLoop) SetPlan(context.Context, uint, string) error { panic("unused") }
func (f *fakeUsersForLoop) SetQuotaWarningSentOn(context.Context, uint, time.Time) error {
	panic("unused")
}
func (f *fakeUsersForLoop) TouchLastLogin(context.Context, uint) error { panic("unused") }
func (f *fakeUsersForLoop) ListInactiveForReengagement(context.Context, time.Time, time.Time, int) ([]models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) MarkReengagementSent(context.Context, uint) error { panic("unused") }
func (f *fakeUsersForLoop) CountReferredBy(context.Context, string) (int64, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) GetByReferralCode(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) MarkReferralRewarded(context.Context, uint) (bool, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) SetInsightEmailsEnabled(context.Context, uint, bool) error {
	panic("unused")
}

// fakeTeamsForLoop — fake TeamRepository, реализует только ListOwnedTeams.
type fakeTeamsForLoop struct {
	ownedByUser map[uint][]models.Team
	err         error
	errOnUser   uint // если != 0, возвращает err только для этого userID
}

func (f *fakeTeamsForLoop) ListOwnedTeams(_ context.Context, userID uint) ([]models.Team, error) {
	if f.err != nil && (f.errOnUser == 0 || f.errOnUser == userID) {
		return nil, f.err
	}
	return f.ownedByUser[userID], nil
}

// Остальные методы TeamRepository — паника.
func (f *fakeTeamsForLoop) CreateWithOwner(context.Context, *models.Team, uint) error {
	panic("unused")
}
func (f *fakeTeamsForLoop) GetBySlug(context.Context, string) (*models.Team, error) {
	panic("unused")
}
func (f *fakeTeamsForLoop) GetByID(context.Context, uint) (*models.Team, error) {
	panic("unused")
}
func (f *fakeTeamsForLoop) ListByUserID(context.Context, uint) ([]models.Team, error) {
	panic("unused")
}
func (f *fakeTeamsForLoop) ListByUserIDWithRolesAndCounts(context.Context, uint) ([]models.TeamWithRoleAndCount, error) {
	panic("unused")
}
func (f *fakeTeamsForLoop) Update(context.Context, *models.Team) error { panic("unused") }
func (f *fakeTeamsForLoop) Delete(context.Context, uint) error         { panic("unused") }
func (f *fakeTeamsForLoop) GetMember(context.Context, uint, uint) (*models.TeamMember, error) {
	panic("unused")
}
func (f *fakeTeamsForLoop) UpdateMemberRole(context.Context, uint, uint, models.TeamRole) error {
	panic("unused")
}
func (f *fakeTeamsForLoop) RemoveMember(context.Context, uint, uint) error { panic("unused") }
func (f *fakeTeamsForLoop) ListMembers(context.Context, uint) ([]models.TeamMember, error) {
	panic("unused")
}
func (f *fakeTeamsForLoop) CountMembers(context.Context, uint) (int, error) { panic("unused") }
func (f *fakeTeamsForLoop) CreateInvitation(context.Context, *models.TeamInvitation) error {
	panic("unused")
}
func (f *fakeTeamsForLoop) GetInvitationByID(context.Context, uint) (*models.TeamInvitation, error) {
	panic("unused")
}
func (f *fakeTeamsForLoop) GetPendingInvitation(context.Context, uint, uint) (*models.TeamInvitation, error) {
	panic("unused")
}
func (f *fakeTeamsForLoop) ListPendingByUserID(context.Context, uint) ([]models.TeamInvitation, error) {
	panic("unused")
}
func (f *fakeTeamsForLoop) ListPendingByTeamID(context.Context, uint) ([]models.TeamInvitation, error) {
	panic("unused")
}
func (f *fakeTeamsForLoop) UpdateInvitationStatus(context.Context, uint, models.InvitationStatus) error {
	panic("unused")
}
func (f *fakeTeamsForLoop) DeleteInvitation(context.Context, uint) error { panic("unused") }
func (f *fakeTeamsForLoop) AcceptInvitationTx(context.Context, uint, *models.TeamMember) error {
	panic("unused")
}
func (f *fakeTeamsForLoop) UpdateBranding(context.Context, uint, string, string, string, string) error {
	panic("unused")
}
func (f *fakeTeamsForLoop) UpdateBrandLogoSource(context.Context, uint, string) error {
	panic("unused")
}

// ===== Тесты =====

func TestInsightsComputeLoop_TeamScope_OwnerGetsTeamInsights(t *testing.T) {
	r := newTrackingRepo()
	svc := newServiceForTest(r, true, true)

	users := &fakeUsersForLoop{ids: []uint{42}}
	teams := &fakeTeamsForLoop{
		ownedByUser: map[uint][]models.Team{
			42: {{ID: 100}, {ID: 200}},
		},
	}

	loop := NewInsightsComputeLoop(svc, users, teams, time.Hour)
	loop.compute()

	// Personal + 2 team scopes = 3 раза вызвался ComputeInsights, значит 3 раза
	// дёрнулся UnusedPrompts (он первый в insights.go).
	assert.Equal(t, 3, r.calls["UnusedPrompts"], "1 personal + 2 team")
}

func TestInsightsComputeLoop_TeamScope_NoOwnedTeams_OnlyPersonal(t *testing.T) {
	r := newTrackingRepo()
	svc := newServiceForTest(r, true, true)

	users := &fakeUsersForLoop{ids: []uint{42}}
	teams := &fakeTeamsForLoop{ownedByUser: map[uint][]models.Team{}} // пусто

	loop := NewInsightsComputeLoop(svc, users, teams, time.Hour)
	loop.compute()

	assert.Equal(t, 1, r.calls["UnusedPrompts"], "только personal scope")
}

func TestInsightsComputeLoop_TeamScope_ListOwnedTeamsFails_ContinuesNextUser(t *testing.T) {
	r := newTrackingRepo()
	svc := newServiceForTest(r, true, true)

	users := &fakeUsersForLoop{ids: []uint{1, 2}}
	teams := &fakeTeamsForLoop{
		ownedByUser: map[uint][]models.Team{
			2: {{ID: 200}},
		},
		err:       errors.New("db down"),
		errOnUser: 1, // ошибка только для user 1
	}

	loop := NewInsightsComputeLoop(svc, users, teams, time.Hour)
	loop.compute()

	// Personal scope для обоих юзеров (2) + team scope для user 2 (1) = 3
	assert.Equal(t, 3, r.calls["UnusedPrompts"])
}

// MN-13: resilience при ошибке ComputeInsights для одного юзера.
// Ожидание: остальные юзеры всё равно обработаются, loop не зависает,
// слабые ошибки не пробрасываются наружу.
func TestInsightsComputeLoop_PersonalComputeFails_ContinuesNextUser(t *testing.T) {
	r := newTrackingRepo()
	r.failOnUserID = 1 // ComputeInsights для user 1 упадёт
	svc := newServiceForTest(r, true, true)

	users := &fakeUsersForLoop{ids: []uint{1, 2, 3}}
	teams := &fakeTeamsForLoop{ownedByUser: map[uint][]models.Team{}}

	loop := NewInsightsComputeLoop(svc, users, teams, time.Hour)
	loop.compute()

	// User 1 — ошибка, user 2 и 3 — успешно. UnusedPrompts должна быть вызвана
	// для всех трёх (errgroup parallelism=4 — все запускаются параллельно).
	assert.Equal(t, 3, r.calls["UnusedPrompts"], "loop должен дойти до всех юзеров несмотря на ошибку")
}

// MN-13: ListPaidUsers вернул empty — loop корректно завершается без вызовов.
func TestInsightsComputeLoop_NoPaidUsers_NoOp(t *testing.T) {
	r := newTrackingRepo()
	svc := newServiceForTest(r, true, true)

	users := &fakeUsersForLoop{ids: []uint{}}
	teams := &fakeTeamsForLoop{}

	loop := NewInsightsComputeLoop(svc, users, teams, time.Hour)
	loop.compute()

	assert.Zero(t, r.calls["UnusedPrompts"], "пустой список — никаких вызовов не должно быть")
}

// MN-13: ListPaidUsers вернул ошибку — loop тихо завершается, метрика error
// инкрементится, panic нет.
func TestInsightsComputeLoop_ListPaidUsersError_NoPanic(t *testing.T) {
	r := newTrackingRepo()
	svc := newServiceForTest(r, true, true)

	users := &fakeUsersForLoop{ids: nil, err: errors.New("db dropped")}
	teams := &fakeTeamsForLoop{}

	loop := NewInsightsComputeLoop(svc, users, teams, time.Hour)

	// Не должна паниковать на nil-deref — функция должна корректно отдать early return.
	require.NotPanics(t, func() { loop.compute() })
	assert.Zero(t, r.calls["UnusedPrompts"])
}

func TestListOwnedTeams_InterfaceContract(t *testing.T) {
	// Sanity-check: интерфейс TeamRepository содержит ListOwnedTeams.
	teams := &fakeTeamsForLoop{ownedByUser: map[uint][]models.Team{1: {{ID: 99}}}}
	got, err := teams.ListOwnedTeams(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, uint(99), got[0].ID)
}

// Task 7: per-plan dispatch — Pro юзер получает 2 типа (unused +
// duplicates), Max юзер получает все 7. Loop читает plan через GetByID и
// передаёт insightsForPlan(plan) в ComputeInsights. Free (race-window
// после ListPaidUsers) — skip без compute.
func TestInsightsComputeLoop_PerPlanDispatch(t *testing.T) {
	r := newTrackingRepo()
	svc := newServiceForTest(r, true, true)

	users := &fakeUsersForLoop{
		ids: []uint{1, 2, 3},
		plans: map[uint]string{
			1: "pro",
			2: "max_yearly",
			3: "free", // race: downgrade между ListPaidUsers и GetByID
		},
	}
	teams := &fakeTeamsForLoop{ownedByUser: map[uint][]models.Team{}}

	loop := NewInsightsComputeLoop(svc, users, teams, time.Hour)
	loop.compute()

	// Pro юзер 1: unused + duplicates = 2 типа.
	// Max юзер 2: 7 типов (unused + 2×trending + most_edited + duplicates +
	//   orphan_tags + empty_collections).
	// Free юзер 3: skip — никаких compute-вызовов.
	//
	// UnusedPrompts вызвалась 2 раза (Pro 1 + Max 2), не 3.
	assert.Equal(t, 2, r.calls["UnusedPrompts"], "Pro+Max → 2, Free → skip")
	// PossibleDuplicates тоже в обоих наборах (Pro teaser + Max).
	assert.Equal(t, 2, r.calls["PossibleDuplicates"], "Pro+Max → 2, Free → skip")
	// Trending/declining — только Max (1 раз × 2 вызова = 2).
	assert.Equal(t, 2, r.calls["GetTrendingPrompts"], "только Max → trending + declining")
	// MostEdited — только Max (1 раз).
	assert.Equal(t, 1, r.calls["MostEditedPrompts"], "только Max")
	// OrphanTags — только Max (1 раз).
	assert.Equal(t, 1, r.calls["OrphanTags"], "только Max")
	// EmptyCollections — только Max (1 раз).
	assert.Equal(t, 1, r.calls["EmptyCollections"], "только Max")
}

// Task 7: race-window — юзер удалён между ListPaidUsers и GetByID.
// Repo.ErrNotFound → skip без логирования error, остальные юзеры
// обрабатываются.
func TestInsightsComputeLoop_UserDeletedBetweenSnapshotAndCompute_Skip(t *testing.T) {
	r := newTrackingRepo()
	svc := newServiceForTest(r, true, true)

	users := &fakeUsersForLoop{
		ids:       []uint{42},
		getByMiss: true, // GetByID вернёт repo.ErrNotFound
	}
	teams := &fakeTeamsForLoop{}

	loop := NewInsightsComputeLoop(svc, users, teams, time.Hour)
	require.NotPanics(t, func() { loop.compute() })

	// Никаких compute-вызовов: ErrNotFound → skip.
	assert.Zero(t, r.calls["UnusedPrompts"])
}
