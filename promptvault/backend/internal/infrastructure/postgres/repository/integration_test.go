package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ─── User Repo ──────────────────────────────────────────────────────────────

func TestUserRepo_CreateAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	r := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:        "alice@example.com",
		Name:         "Alice",
		PasswordHash: "hash123",
	}
	require.NoError(t, r.Create(ctx, user))
	assert.NotZero(t, user.ID)

	got, err := r.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", got.Email)
	assert.Equal(t, "Alice", got.Name)
	assert.Equal(t, "hash123", got.PasswordHash)
	assert.False(t, got.CreatedAt.IsZero())
}

func TestUserRepo_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	r := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "bob@example.com", Name: "Bob"}
	require.NoError(t, r.Create(ctx, user))

	got, err := r.GetByEmail(ctx, "bob@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)
	assert.Equal(t, "Bob", got.Name)
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	r := NewUserRepository(db)
	ctx := context.Background()

	_, err := r.GetByID(ctx, 99999)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestUserRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	r := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "carol@example.com", Name: "Carol"}
	require.NoError(t, r.Create(ctx, user))

	user.Name = "Carol Updated"
	require.NoError(t, r.Update(ctx, user))

	got, err := r.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Carol Updated", got.Name)
}

// ─── Prompt Repo ────────────────────────────────────────────────────────────

func TestPromptRepo_CreateAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	promptRepo := NewPromptRepository(db)
	tagRepo := NewTagRepository(db)
	collRepo := NewCollectionRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "prompt-user@test.com", Name: "Prompt User"}
	require.NoError(t, userRepo.Create(ctx, user))

	tag, err := tagRepo.GetOrCreate(ctx, "golang", "#00ADD8", user.ID, nil)
	require.NoError(t, err)

	coll := &models.Collection{UserID: user.ID, Name: "My Collection"}
	require.NoError(t, collRepo.Create(ctx, coll))

	prompt := &models.Prompt{
		UserID:      user.ID,
		Title:       "Test Prompt",
		Content:     "Hello world",
		Model:       "gpt-4",
		Tags:        []models.Tag{*tag},
		Collections: []models.Collection{*coll},
	}
	require.NoError(t, promptRepo.Create(ctx, prompt))
	assert.NotZero(t, prompt.ID)

	got, err := promptRepo.GetByID(ctx, prompt.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test Prompt", got.Title)
	assert.Equal(t, "Hello world", got.Content)
	assert.Len(t, got.Tags, 1)
	assert.Equal(t, "golang", got.Tags[0].Name)
	assert.Len(t, got.Collections, 1)
	assert.Equal(t, "My Collection", got.Collections[0].Name)
}

func TestPromptRepo_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "del-user@test.com", Name: "Del User"}
	require.NoError(t, userRepo.Create(ctx, user))

	prompt := &models.Prompt{UserID: user.ID, Title: "To Delete", Content: "bye"}
	require.NoError(t, promptRepo.Create(ctx, prompt))

	require.NoError(t, promptRepo.SoftDelete(ctx, prompt.ID))

	_, err := promptRepo.GetByID(ctx, prompt.ID)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestPromptRepo_SetFavorite(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "fav-user@test.com", Name: "Fav User"}
	require.NoError(t, userRepo.Create(ctx, user))

	prompt := &models.Prompt{UserID: user.ID, Title: "Fav Prompt", Content: "star me"}
	require.NoError(t, promptRepo.Create(ctx, prompt))
	assert.False(t, prompt.Favorite)

	require.NoError(t, promptRepo.SetFavorite(ctx, prompt.ID, true))

	got, err := promptRepo.GetByID(ctx, prompt.ID)
	require.NoError(t, err)
	assert.True(t, got.Favorite)
}

func TestPromptRepo_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "list-user@test.com", Name: "List User"}
	require.NoError(t, userRepo.Create(ctx, user))

	for i := 0; i < 5; i++ {
		p := &models.Prompt{
			UserID:  user.ID,
			Title:   fmt.Sprintf("Prompt %d", i),
			Content: fmt.Sprintf("Content %d", i),
		}
		require.NoError(t, promptRepo.Create(ctx, p))
	}

	prompts, total, err := promptRepo.List(ctx, repo.PromptListFilter{
		UserID:   user.ID,
		Page:     1,
		PageSize: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, prompts, 2)
}

func TestPromptRepo_List_Search(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "search-user@test.com", Name: "Search User"}
	require.NoError(t, userRepo.Create(ctx, user))

	prompts := []*models.Prompt{
		{UserID: user.ID, Title: "Kubernetes deployment", Content: "kubectl apply"},
		{UserID: user.ID, Title: "Docker compose", Content: "docker-compose up"},
		{UserID: user.ID, Title: "Go testing", Content: "go test ./..."},
	}
	for _, p := range prompts {
		require.NoError(t, promptRepo.Create(ctx, p))
	}

	results, total, err := promptRepo.List(ctx, repo.PromptListFilter{
		UserID:   user.ID,
		Query:    "docker",
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "Docker compose", results[0].Title)
}

// ─── Collection Repo ────────────────────────────────────────────────────────

func TestCollectionRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	collRepo := NewCollectionRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "coll-user@test.com", Name: "Coll User"}
	require.NoError(t, userRepo.Create(ctx, user))

	// Create
	coll := &models.Collection{UserID: user.ID, Name: "Work", Description: "Work prompts", Color: "#ff0000"}
	require.NoError(t, collRepo.Create(ctx, coll))
	assert.NotZero(t, coll.ID)

	// Get
	got, err := collRepo.GetByID(ctx, coll.ID)
	require.NoError(t, err)
	assert.Equal(t, "Work", got.Name)
	assert.Equal(t, "#ff0000", got.Color)

	// Update
	got.Name = "Work Updated"
	require.NoError(t, collRepo.Update(ctx, got))

	got2, err := collRepo.GetByID(ctx, coll.ID)
	require.NoError(t, err)
	assert.Equal(t, "Work Updated", got2.Name)

	// Delete
	require.NoError(t, collRepo.Delete(ctx, coll.ID))

	_, err = collRepo.GetByID(ctx, coll.ID)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestCollectionRepo_DeleteByID(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	collRepo := NewCollectionRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "owner@test.com", Name: "Owner"}
	require.NoError(t, userRepo.Create(ctx, user))

	coll := &models.Collection{UserID: user.ID, Name: "ToDelete"}
	require.NoError(t, collRepo.Create(ctx, coll))

	// Delete by ID only — authorization is enforced at the usecase layer
	require.NoError(t, collRepo.Delete(ctx, coll.ID))

	// Collection should be gone
	_, err := collRepo.GetByID(ctx, coll.ID)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestCollectionRepo_ListWithCounts(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	collRepo := NewCollectionRepository(db)
	promptRepo := NewPromptRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "counts-user@test.com", Name: "Counts User"}
	require.NoError(t, userRepo.Create(ctx, user))

	coll := &models.Collection{UserID: user.ID, Name: "Counted"}
	require.NoError(t, collRepo.Create(ctx, coll))

	// Create 3 prompts linked to the collection
	for i := 0; i < 3; i++ {
		p := &models.Prompt{
			UserID:      user.ID,
			Title:       fmt.Sprintf("Linked %d", i),
			Content:     "content",
			Collections: []models.Collection{*coll},
		}
		require.NoError(t, promptRepo.Create(ctx, p))
	}

	results, err := collRepo.ListWithCounts(ctx, user.ID, nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Counted", results[0].Name)
	assert.Equal(t, int64(3), results[0].PromptCount)
}

// ─── Tag Repo ───────────────────────────────────────────────────────────────

func TestTagRepo_GetOrCreate(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "tag-user@test.com", Name: "Tag User"}
	require.NoError(t, userRepo.Create(ctx, user))

	// First call — creates
	tag1, err := tagRepo.GetOrCreate(ctx, "python", "#3776AB", user.ID, nil)
	require.NoError(t, err)
	assert.NotZero(t, tag1.ID)
	assert.Equal(t, "python", tag1.Name)

	// Second call — returns existing
	tag2, err := tagRepo.GetOrCreate(ctx, "python", "#3776AB", user.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, tag1.ID, tag2.ID)
}

func TestTagRepo_List(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "taglist-user@test.com", Name: "Tag List User"}
	require.NoError(t, userRepo.Create(ctx, user))

	team := &models.Team{Slug: "tag-team", Name: "Tag Team", CreatedBy: user.ID}
	require.NoError(t, teamRepo.CreateWithOwner(ctx, team, user.ID))

	// Personal tag
	_, err := tagRepo.GetOrCreate(ctx, "personal-tag", "#111111", user.ID, nil)
	require.NoError(t, err)

	// Team tag
	_, err = tagRepo.GetOrCreate(ctx, "team-tag", "#222222", user.ID, &team.ID)
	require.NoError(t, err)

	// List personal — should only see personal tag
	personalTags, err := tagRepo.List(ctx, user.ID, nil)
	require.NoError(t, err)
	assert.Len(t, personalTags, 1)
	assert.Equal(t, "personal-tag", personalTags[0].Name)

	// List team — should only see team tag
	teamTags, err := tagRepo.List(ctx, user.ID, &team.ID)
	require.NoError(t, err)
	assert.Len(t, teamTags, 1)
	assert.Equal(t, "team-tag", teamTags[0].Name)
}

func TestTagRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	tagRepo := NewTagRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "tagdel-user@test.com", Name: "Tag Del User"}
	require.NoError(t, userRepo.Create(ctx, user))

	tag, err := tagRepo.GetOrCreate(ctx, "to-delete", "#ff0000", user.ID, nil)
	require.NoError(t, err)

	require.NoError(t, tagRepo.Delete(ctx, tag.ID))

	_, err = tagRepo.GetByID(ctx, tag.ID)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

// ─── Team Repo ──────────────────────────────────────────────────────────────

func TestTeamRepo_CreateWithOwner(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "team-owner@test.com", Name: "Team Owner"}
	require.NoError(t, userRepo.Create(ctx, user))

	team := &models.Team{Slug: "my-team", Name: "My Team", CreatedBy: user.ID}
	require.NoError(t, teamRepo.CreateWithOwner(ctx, team, user.ID))
	assert.NotZero(t, team.ID)

	// Verify owner member was created
	member, err := teamRepo.GetMember(ctx, team.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, models.RoleOwner, member.Role)
}

func TestTeamRepo_GetMember(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	ctx := context.Background()

	owner := &models.User{Email: "gm-owner@test.com", Name: "Owner"}
	require.NoError(t, userRepo.Create(ctx, owner))

	stranger := &models.User{Email: "gm-stranger@test.com", Name: "Stranger"}
	require.NoError(t, userRepo.Create(ctx, stranger))

	team := &models.Team{Slug: "gm-team", Name: "GM Team", CreatedBy: owner.ID}
	require.NoError(t, teamRepo.CreateWithOwner(ctx, team, owner.ID))

	// Existing member
	member, err := teamRepo.GetMember(ctx, team.ID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, models.RoleOwner, member.Role)

	// Non-member returns ErrNotFound
	_, err = teamRepo.GetMember(ctx, team.ID, stranger.ID)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestTeamRepo_InvitationFlow(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	teamRepo := NewTeamRepository(db)
	ctx := context.Background()

	owner := &models.User{Email: "inv-owner@test.com", Name: "Owner"}
	invitee := &models.User{Email: "inv-invitee@test.com", Name: "Invitee"}
	require.NoError(t, userRepo.Create(ctx, owner))
	require.NoError(t, userRepo.Create(ctx, invitee))

	team := &models.Team{Slug: "inv-team", Name: "Inv Team", CreatedBy: owner.ID}
	require.NoError(t, teamRepo.CreateWithOwner(ctx, team, owner.ID))

	// Create invitation
	inv := &models.TeamInvitation{
		TeamID:    team.ID,
		UserID:    invitee.ID,
		InviterID: owner.ID,
		Role:      models.RoleEditor,
		Status:    models.InvitationPending,
	}
	require.NoError(t, teamRepo.CreateInvitation(ctx, inv))
	assert.NotZero(t, inv.ID)

	// Accept invitation (creates member atomically)
	member := &models.TeamMember{
		TeamID: team.ID,
		UserID: invitee.ID,
		Role:   inv.Role,
	}
	require.NoError(t, teamRepo.AcceptInvitationTx(ctx, inv.ID, member))

	// Verify invitation status updated
	gotInv, err := teamRepo.GetInvitationByID(ctx, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, models.InvitationAccepted, gotInv.Status)

	// Verify member created
	gotMember, err := teamRepo.GetMember(ctx, team.ID, invitee.ID)
	require.NoError(t, err)
	assert.Equal(t, models.RoleEditor, gotMember.Role)
}

// ─── Version Repo ───────────────────────────────────────────────────────────

func TestVersionRepo_CreateWithNextVersion_Atomicity(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	promptRepo := NewPromptRepository(db)
	versionRepo := NewVersionRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "ver-user@test.com", Name: "Ver User"}
	require.NoError(t, userRepo.Create(ctx, user))

	prompt := &models.Prompt{UserID: user.ID, Title: "Versioned", Content: "v0"}
	require.NoError(t, promptRepo.Create(ctx, prompt))

	// Create 3 versions sequentially
	for i := 0; i < 3; i++ {
		v := &models.PromptVersion{
			PromptID:   prompt.ID,
			Title:      fmt.Sprintf("Version %d", i+1),
			Content:    fmt.Sprintf("Content v%d", i+1),
			ChangeNote: fmt.Sprintf("Change %d", i+1),
		}
		require.NoError(t, versionRepo.CreateWithNextVersion(ctx, v))
		assert.Equal(t, uint(i+1), v.VersionNumber)
	}

	// Verify all 3 versions exist with correct numbers
	versions, total, err := versionRepo.ListByPromptID(ctx, prompt.ID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, versions, 3)
	// ListByPromptID orders DESC so first is version 3
	assert.Equal(t, uint(3), versions[0].VersionNumber)
	assert.Equal(t, uint(1), versions[2].VersionNumber)
}

func TestVersionRepo_GetByIDForPrompt(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	promptRepo := NewPromptRepository(db)
	versionRepo := NewVersionRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "verget-user@test.com", Name: "VerGet User"}
	require.NoError(t, userRepo.Create(ctx, user))

	prompt1 := &models.Prompt{UserID: user.ID, Title: "Prompt1", Content: "c1"}
	prompt2 := &models.Prompt{UserID: user.ID, Title: "Prompt2", Content: "c2"}
	require.NoError(t, promptRepo.Create(ctx, prompt1))
	require.NoError(t, promptRepo.Create(ctx, prompt2))

	v := &models.PromptVersion{PromptID: prompt1.ID, Title: "V1", Content: "cv1"}
	require.NoError(t, versionRepo.CreateWithNextVersion(ctx, v))

	// Get existing version for correct prompt
	got, err := versionRepo.GetByIDForPrompt(ctx, v.ID, prompt1.ID)
	require.NoError(t, err)
	assert.Equal(t, "V1", got.Title)

	// Get same version but for wrong prompt → NotFound
	_, err = versionRepo.GetByIDForPrompt(ctx, v.ID, prompt2.ID)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

// ─── Verification Repo ──────────────────────────────────────────────────────

func TestVerificationRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	verRepo := NewVerificationRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "verify-user@test.com", Name: "Verify User"}
	require.NoError(t, userRepo.Create(ctx, user))

	// Create
	v := &models.EmailVerification{
		UserID:    user.ID,
		Code:      "123456",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	require.NoError(t, verRepo.Create(ctx, v))
	assert.NotZero(t, v.ID)

	// Get by user ID
	got, err := verRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "123456", got.Code)
	assert.Equal(t, 0, got.Attempts)

	// Increment attempts
	require.NoError(t, verRepo.IncrementAttempts(ctx, got.ID))

	got2, err := verRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got2.Attempts)

	// Delete
	require.NoError(t, verRepo.DeleteByUserID(ctx, user.ID))

	// After delete, GetByUserID returns repo.ErrNotFound
	_, err = verRepo.GetByUserID(ctx, user.ID)
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestVerificationRepo_GetByUserID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	verRepo := NewVerificationRepository(db)
	ctx := context.Background()

	_, err := verRepo.GetByUserID(ctx, 99999)
	assert.Error(t, err)
}

// ─── LinkedAccount Repo ─────────────────────────────────────────────────────

func TestLinkedAccountRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db)
	laRepo := NewLinkedAccountRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "linked-user@test.com", Name: "Linked User"}
	require.NoError(t, userRepo.Create(ctx, user))

	// Create
	la := &models.LinkedAccount{
		UserID:     user.ID,
		Provider:   "github",
		ProviderID: "gh-12345",
	}
	require.NoError(t, laRepo.Create(ctx, la))

	// Get by provider ID
	got, err := laRepo.GetByProviderID(ctx, "github", "gh-12345")
	require.NoError(t, err)
	assert.Equal(t, user.ID, got.UserID)
	assert.Equal(t, "github", got.Provider)

	// Get by user ID
	accounts, err := laRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, accounts, 1)

	// Count
	count, err := laRepo.CountByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete
	require.NoError(t, laRepo.Delete(ctx, user.ID, "github"))

	accounts2, err := laRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, accounts2, 0)
}

func TestLinkedAccountRepo_GetByProviderID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	laRepo := NewLinkedAccountRepository(db)
	ctx := context.Background()

	_, err := laRepo.GetByProviderID(ctx, "github", "nonexistent")
	assert.ErrorIs(t, err, repo.ErrNotFound)
}
