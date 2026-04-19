package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

func newOAuthRepos(t *testing.T) (repo.OAuthClientRepository, repo.OAuthAuthorizationCodeRepository, repo.OAuthTokenRepository, *models.User) {
	t.Helper()
	db := setupTestDB(t)
	u := createTestUser(t, db, "oauth-test@example.com")
	return NewOAuthClientRepository(db),
		NewOAuthAuthorizationCodeRepository(db),
		NewOAuthTokenRepository(db),
		u
}

func createTestClient(t *testing.T, clients repo.OAuthClientRepository) *models.OAuthClient {
	t.Helper()
	ctx := context.Background()
	c := &models.OAuthClient{
		ClientID:                "pvoci_test_" + t.Name(),
		ClientName:              "Test Client",
		RedirectURIs:            pq.StringArray{"https://claude.ai/api/mcp/auth_callback"},
		GrantTypes:              pq.StringArray{"authorization_code", "refresh_token"},
		ResponseTypes:           pq.StringArray{"code"},
		TokenEndpointAuthMethod: "none",
		Scope:                   "mcp:read mcp:write",
		IsDynamic:               true,
	}
	require.NoError(t, clients.Create(ctx, c))
	return c
}

// ---------------------------------------------------------------------------
// OAuthClientRepository
// ---------------------------------------------------------------------------

func TestOAuthClientRepo_CreateAndGet(t *testing.T) {
	clients, _, _, _ := newOAuthRepos(t)
	ctx := context.Background()

	c := createTestClient(t, clients)

	got, err := clients.GetByClientID(ctx, c.ClientID)
	require.NoError(t, err)
	assert.Equal(t, c.ClientName, got.ClientName)
	assert.Equal(t, []string(c.RedirectURIs), []string(got.RedirectURIs))
}

func TestOAuthClientRepo_GetNonExistent(t *testing.T) {
	clients, _, _, _ := newOAuthRepos(t)
	ctx := context.Background()

	_, err := clients.GetByClientID(ctx, "pvoci_missing")
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestOAuthClientRepo_UpdateLastUsed(t *testing.T) {
	clients, _, _, _ := newOAuthRepos(t)
	ctx := context.Background()
	c := createTestClient(t, clients)

	before, err := clients.GetByClientID(ctx, c.ClientID)
	require.NoError(t, err)
	assert.Nil(t, before.LastUsedAt)

	require.NoError(t, clients.UpdateLastUsed(ctx, c.ClientID))

	after, err := clients.GetByClientID(ctx, c.ClientID)
	require.NoError(t, err)
	require.NotNil(t, after.LastUsedAt)
	assert.WithinDuration(t, time.Now(), *after.LastUsedAt, 5*time.Second)
}

// ---------------------------------------------------------------------------
// OAuthAuthorizationCodeRepository
// ---------------------------------------------------------------------------

func TestOAuthCodeRepo_Consume_OneTime(t *testing.T) {
	clients, codes, _, user := newOAuthRepos(t)
	ctx := context.Background()
	c := createTestClient(t, clients)

	code := &models.OAuthAuthorizationCode{
		CodeHash:            "hash-onetime",
		ClientID:            c.ClientID,
		UserID:              user.ID,
		RedirectURI:         c.RedirectURIs[0],
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		Scope:               "mcp:read",
		Resource:            "https://promtlabs.ru/mcp",
		Policy:              json.RawMessage(`{}`),
		ExpiresAt:           time.Now().Add(60 * time.Second),
	}
	require.NoError(t, codes.Create(ctx, code))

	// First consume succeeds.
	got, err := codes.Consume(ctx, "hash-onetime")
	require.NoError(t, err)
	assert.Equal(t, user.ID, got.UserID)
	require.NotNil(t, got.UsedAt)

	// Second consume must fail — one-time.
	_, err = codes.Consume(ctx, "hash-onetime")
	assert.ErrorIs(t, err, repo.ErrNotFound, "код уже использован → второй Consume должен вернуть ErrNotFound")
}

func TestOAuthCodeRepo_Consume_Expired(t *testing.T) {
	clients, codes, _, user := newOAuthRepos(t)
	ctx := context.Background()
	c := createTestClient(t, clients)

	code := &models.OAuthAuthorizationCode{
		CodeHash:            "hash-expired",
		ClientID:            c.ClientID,
		UserID:              user.ID,
		RedirectURI:         c.RedirectURIs[0],
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		Scope:               "mcp:read",
		Resource:            "https://promtlabs.ru/mcp",
		Policy:              json.RawMessage(`{}`),
		ExpiresAt:           time.Now().Add(-1 * time.Second), // already expired
	}
	require.NoError(t, codes.Create(ctx, code))

	_, err := codes.Consume(ctx, "hash-expired")
	assert.ErrorIs(t, err, repo.ErrNotFound, "истёкший код не должен извлекаться")
}

func TestOAuthCodeRepo_DeleteExpired(t *testing.T) {
	clients, codes, _, user := newOAuthRepos(t)
	ctx := context.Background()
	c := createTestClient(t, clients)

	// Один истёкший, один валидный
	expired := &models.OAuthAuthorizationCode{
		CodeHash: "expired", ClientID: c.ClientID, UserID: user.ID,
		RedirectURI: c.RedirectURIs[0], CodeChallenge: "x", CodeChallengeMethod: "S256",
		Scope: "mcp:read", Resource: "https://promtlabs.ru/mcp",
		Policy: json.RawMessage(`{}`), ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	valid := &models.OAuthAuthorizationCode{
		CodeHash: "valid", ClientID: c.ClientID, UserID: user.ID,
		RedirectURI: c.RedirectURIs[0], CodeChallenge: "x", CodeChallengeMethod: "S256",
		Scope: "mcp:read", Resource: "https://promtlabs.ru/mcp",
		Policy: json.RawMessage(`{}`), ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	require.NoError(t, codes.Create(ctx, expired))
	require.NoError(t, codes.Create(ctx, valid))

	deleted, err := codes.DeleteExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted, "должен удалить 1 истёкший код")
}

// ---------------------------------------------------------------------------
// OAuthTokenRepository
// ---------------------------------------------------------------------------

func TestOAuthTokenRepo_CreateAndGet(t *testing.T) {
	clients, _, tokens, user := newOAuthRepos(t)
	ctx := context.Background()
	c := createTestClient(t, clients)

	tok := &models.OAuthToken{
		TokenType: "access",
		TokenHash: "access-hash-1",
		ClientID:  c.ClientID,
		UserID:    user.ID,
		Scope:     "mcp:read mcp:write",
		Resource:  "https://promtlabs.ru/mcp",
		Policy:    json.RawMessage(`{"read_only":false}`),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	require.NoError(t, tokens.Create(ctx, tok))

	got, err := tokens.GetByHash(ctx, "access-hash-1")
	require.NoError(t, err)
	assert.Equal(t, "access", got.TokenType)
	assert.Equal(t, user.ID, got.UserID)
}

func TestOAuthTokenRepo_Revoke(t *testing.T) {
	clients, _, tokens, user := newOAuthRepos(t)
	ctx := context.Background()
	c := createTestClient(t, clients)

	tok := &models.OAuthToken{
		TokenType: "refresh", TokenHash: "refresh-hash",
		ClientID: c.ClientID, UserID: user.ID,
		Scope: "mcp:read", Resource: "https://promtlabs.ru/mcp",
		Policy: json.RawMessage(`{}`), ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	require.NoError(t, tokens.Create(ctx, tok))

	require.NoError(t, tokens.Revoke(ctx, "refresh-hash"))

	got, err := tokens.GetByHash(ctx, "refresh-hash")
	require.NoError(t, err)
	require.NotNil(t, got.RevokedAt, "revoked_at должен быть установлен после Revoke")
}

func TestOAuthTokenRepo_Revoke_NotFound(t *testing.T) {
	_, _, tokens, _ := newOAuthRepos(t)
	ctx := context.Background()
	err := tokens.Revoke(ctx, "never-existed")
	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestOAuthTokenRepo_RevokeChain(t *testing.T) {
	clients, _, tokens, user := newOAuthRepos(t)
	ctx := context.Background()
	c := createTestClient(t, clients)

	// Цепочка refresh-rotation: root → child1 → child2
	root := &models.OAuthToken{
		TokenType: "refresh", TokenHash: "root",
		ClientID: c.ClientID, UserID: user.ID,
		Scope: "mcp:read", Resource: "https://promtlabs.ru/mcp",
		Policy: json.RawMessage(`{}`), ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	require.NoError(t, tokens.Create(ctx, root))

	child1 := &models.OAuthToken{
		TokenType: "refresh", TokenHash: "child1",
		ClientID: c.ClientID, UserID: user.ID,
		Scope: "mcp:read", Resource: "https://promtlabs.ru/mcp",
		Policy: json.RawMessage(`{}`), ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		ParentTokenID: &root.ID,
	}
	require.NoError(t, tokens.Create(ctx, child1))

	child2 := &models.OAuthToken{
		TokenType: "refresh", TokenHash: "child2",
		ClientID: c.ClientID, UserID: user.ID,
		Scope: "mcp:read", Resource: "https://promtlabs.ru/mcp",
		Policy: json.RawMessage(`{}`), ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		ParentTokenID: &child1.ID,
	}
	require.NoError(t, tokens.Create(ctx, child2))

	// RevokeChain от root должен revoke всех трёх.
	require.NoError(t, tokens.RevokeChain(ctx, root.ID))

	for _, hash := range []string{"root", "child1", "child2"} {
		got, err := tokens.GetByHash(ctx, hash)
		require.NoError(t, err)
		assert.NotNil(t, got.RevokedAt, "token %s должен быть revoked", hash)
	}
}
