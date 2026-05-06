package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"promptvault/internal/models"
)

// fakeUsersForLoop — минимальный fake UserRepository, нужен только ListMaxUsers.
type fakeUsersForLoop struct {
	ids []uint
	err error
}

func (f *fakeUsersForLoop) ListMaxUsers(_ context.Context) ([]uint, error) {
	return f.ids, f.err
}

// Остальные методы UserRepository не используются loop'ом — паника.
func (f *fakeUsersForLoop) Create(context.Context, *models.User) error      { panic("unused") }
func (f *fakeUsersForLoop) GetByID(context.Context, uint) (*models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) GetByEmail(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) GetByUsername(context.Context, string) (*models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) SearchUsers(context.Context, string, int) ([]models.User, error) {
	panic("unused")
}
func (f *fakeUsersForLoop) Update(context.Context, *models.User) error { panic("unused") }
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

func TestListOwnedTeams_InterfaceContract(t *testing.T) {
	// Sanity-check: интерфейс TeamRepository содержит ListOwnedTeams.
	teams := &fakeTeamsForLoop{ownedByUser: map[uint][]models.Team{1: {{ID: 99}}}}
	got, err := teams.ListOwnedTeams(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, uint(99), got[0].ID)
}
