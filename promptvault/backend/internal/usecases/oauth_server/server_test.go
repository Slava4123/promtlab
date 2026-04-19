package oauth_server

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	"promptvault/internal/pkg/pkce"
	"promptvault/internal/pkg/tokens"
)

const testResource = "https://promtlabs.ru/mcp"

// -----------------------------------------------------------------------------
// Mocks (testify/mock) — тот же паттерн, что в usecases/apikey/apikey_test.go.
// -----------------------------------------------------------------------------

type mockClientRepo struct{ mock.Mock }

func (m *mockClientRepo) Create(ctx context.Context, c *models.OAuthClient) error {
	// GORM на реальной БД устанавливает ID; в тестах имитируем, чтобы последующие
	// вызовы UpdateLastUsed имели корректный FK.
	c.ID = 1
	return m.Called(ctx, c).Error(0)
}
func (m *mockClientRepo) GetByClientID(ctx context.Context, cid string) (*models.OAuthClient, error) {
	a := m.Called(ctx, cid)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*models.OAuthClient), a.Error(1)
}
func (m *mockClientRepo) UpdateLastUsed(ctx context.Context, cid string) error {
	return m.Called(ctx, cid).Error(0)
}
func (m *mockClientRepo) Delete(ctx context.Context, cid string) error {
	return m.Called(ctx, cid).Error(0)
}

type mockCodeRepo struct{ mock.Mock }

func (m *mockCodeRepo) Create(ctx context.Context, c *models.OAuthAuthorizationCode) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockCodeRepo) Consume(ctx context.Context, h string) (*models.OAuthAuthorizationCode, error) {
	a := m.Called(ctx, h)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*models.OAuthAuthorizationCode), a.Error(1)
}
func (m *mockCodeRepo) DeleteExpired(ctx context.Context) (int64, error) {
	a := m.Called(ctx)
	return a.Get(0).(int64), a.Error(1)
}

type mockTokenRepo struct{ mock.Mock }

func (m *mockTokenRepo) Create(ctx context.Context, t *models.OAuthToken) error {
	t.ID = uint(time.Now().UnixNano()) // уникальный ID для parent_token_id chain'ов
	return m.Called(ctx, t).Error(0)
}
func (m *mockTokenRepo) GetByHash(ctx context.Context, h string) (*models.OAuthToken, error) {
	a := m.Called(ctx, h)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*models.OAuthToken), a.Error(1)
}
func (m *mockTokenRepo) Revoke(ctx context.Context, h string) error {
	return m.Called(ctx, h).Error(0)
}
func (m *mockTokenRepo) RevokeChain(ctx context.Context, parentID uint) error {
	return m.Called(ctx, parentID).Error(0)
}
func (m *mockTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	a := m.Called(ctx)
	return a.Get(0).(int64), a.Error(1)
}

func newTestService() (*Service, *mockClientRepo, *mockCodeRepo, *mockTokenRepo) {
	cr, cdr, tr := new(mockClientRepo), new(mockCodeRepo), new(mockTokenRepo)
	return NewService(cr, cdr, tr, testResource), cr, cdr, tr
}

func sampleClient() *models.OAuthClient {
	return &models.OAuthClient{
		ID:                      1,
		ClientID:                "pvoci_test123",
		ClientName:              "Test Client",
		RedirectURIs:            pq.StringArray{"https://claude.ai/api/mcp/auth_callback"},
		GrantTypes:              pq.StringArray{"authorization_code", "refresh_token"},
		ResponseTypes:           pq.StringArray{"code"},
		TokenEndpointAuthMethod: "none",
		Scope:                   "mcp:read mcp:write",
	}
}

// -----------------------------------------------------------------------------
// RegisterClient
// -----------------------------------------------------------------------------

func TestRegisterClient_Success(t *testing.T) {
	svc, cr, _, _ := newTestService()
	ctx := context.Background()
	cr.On("Create", ctx, mock.AnythingOfType("*models.OAuthClient")).Return(nil)

	out, err := svc.RegisterClient(ctx, RegisterClientInput{
		ClientName:   "Claude.ai",
		RedirectURIs: []string{"https://claude.ai/api/mcp/auth_callback"},
	})

	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(out.ClientID, tokens.PrefixClientID))
	assert.Empty(t, out.ClientSecret, "public client (auth_method=none) → без секрета")
	assert.Equal(t, DefaultScope, out.Scope)
	assert.Contains(t, out.GrantTypes, "authorization_code")
}

func TestRegisterClient_MissingClientName(t *testing.T) {
	svc, _, _, _ := newTestService()
	_, err := svc.RegisterClient(context.Background(), RegisterClientInput{
		RedirectURIs: []string{"https://claude.ai/api/mcp/auth_callback"},
	})
	assert.ErrorIs(t, err, ErrInvalidRequest)
}

func TestRegisterClient_HTTPSchemeNonLocalhost(t *testing.T) {
	svc, _, _, _ := newTestService()
	_, err := svc.RegisterClient(context.Background(), RegisterClientInput{
		ClientName:   "Bad",
		RedirectURIs: []string{"http://evil.example.com/cb"},
	})
	assert.ErrorIs(t, err, ErrInvalidRequest, "http:// разрешён только для localhost")
}

func TestRegisterClient_LocalhostHTTPAllowed(t *testing.T) {
	svc, cr, _, _ := newTestService()
	ctx := context.Background()
	cr.On("Create", ctx, mock.AnythingOfType("*models.OAuthClient")).Return(nil)
	_, err := svc.RegisterClient(ctx, RegisterClientInput{
		ClientName:   "Local Dev",
		RedirectURIs: []string{"http://localhost:6274/oauth/callback"},
	})
	assert.NoError(t, err)
}

// -----------------------------------------------------------------------------
// Authorize
// -----------------------------------------------------------------------------

func TestAuthorize_Success(t *testing.T) {
	svc, cr, cdr, _ := newTestService()
	ctx := context.Background()
	client := sampleClient()
	cr.On("GetByClientID", ctx, client.ClientID).Return(client, nil)
	cdr.On("Create", ctx, mock.AnythingOfType("*models.OAuthAuthorizationCode")).Return(nil)

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := pkce.ComputeS256(verifier)

	out, err := svc.Authorize(ctx, AuthorizeInput{
		UserID:              42,
		ClientID:            client.ClientID,
		RedirectURI:         client.RedirectURIs[0],
		Scope:               "mcp:read",
		State:               "random-state",
		CodeChallenge:       challenge,
		CodeChallengeMethod: pkce.MethodS256,
		Resource:            testResource,
	})

	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(out.Code, tokens.PrefixAuthCode))
	assert.Equal(t, "random-state", out.State)
}

func TestAuthorize_MissingPKCE(t *testing.T) {
	svc, _, _, _ := newTestService()
	_, err := svc.Authorize(context.Background(), AuthorizeInput{
		UserID:      42,
		ClientID:    "any",
		RedirectURI: "https://x",
	})
	assert.ErrorIs(t, err, ErrPKCERequired)
}

func TestAuthorize_UnknownClient(t *testing.T) {
	svc, cr, _, _ := newTestService()
	ctx := context.Background()
	cr.On("GetByClientID", ctx, "unknown").Return(nil, repo.ErrNotFound)

	_, err := svc.Authorize(ctx, AuthorizeInput{
		UserID: 1, ClientID: "unknown", RedirectURI: "https://x",
		CodeChallenge: "c", CodeChallengeMethod: pkce.MethodS256,
	})
	assert.ErrorIs(t, err, ErrClientNotFound)
}

func TestAuthorize_RedirectURIMismatch(t *testing.T) {
	svc, cr, _, _ := newTestService()
	ctx := context.Background()
	client := sampleClient()
	cr.On("GetByClientID", ctx, client.ClientID).Return(client, nil)

	_, err := svc.Authorize(ctx, AuthorizeInput{
		UserID: 1, ClientID: client.ClientID,
		RedirectURI:   "https://evil.example.com/cb",
		CodeChallenge: "c", CodeChallengeMethod: pkce.MethodS256,
	})
	assert.ErrorIs(t, err, ErrInvalidRedirectURI)
}

func TestAuthorize_ResourceMismatch(t *testing.T) {
	svc, cr, _, _ := newTestService()
	ctx := context.Background()
	client := sampleClient()
	cr.On("GetByClientID", ctx, client.ClientID).Return(client, nil)

	_, err := svc.Authorize(ctx, AuthorizeInput{
		UserID: 1, ClientID: client.ClientID, RedirectURI: client.RedirectURIs[0],
		CodeChallenge: "c", CodeChallengeMethod: pkce.MethodS256,
		Resource: "https://other.example.com/mcp",
	})
	assert.ErrorIs(t, err, ErrResourceMismatch)
}

// -----------------------------------------------------------------------------
// ExchangeCode
// -----------------------------------------------------------------------------

func TestExchangeCode_Success(t *testing.T) {
	svc, cr, cdr, tr := newTestService()
	ctx := context.Background()
	client := sampleClient()
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := pkce.ComputeS256(verifier)

	code := &models.OAuthAuthorizationCode{
		ClientID:            client.ClientID,
		UserID:              42,
		RedirectURI:         client.RedirectURIs[0],
		CodeChallenge:       challenge,
		CodeChallengeMethod: pkce.MethodS256,
		Scope:               "mcp:read",
		Resource:            testResource,
		Policy:              json.RawMessage(`{}`),
	}
	cdr.On("Consume", ctx, mock.AnythingOfType("string")).Return(code, nil)
	tr.On("Create", ctx, mock.AnythingOfType("*models.OAuthToken")).Return(nil).Times(2)
	cr.On("UpdateLastUsed", ctx, client.ClientID).Return(nil)

	out, err := svc.ExchangeCode(ctx, ExchangeCodeInput{
		ClientID:     client.ClientID,
		Code:         "pvoac_rawcode",
		RedirectURI:  client.RedirectURIs[0],
		CodeVerifier: verifier,
		Resource:     testResource,
	})

	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(out.AccessToken, tokens.PrefixAccessToken))
	assert.True(t, strings.HasPrefix(out.RefreshToken, tokens.PrefixRefreshToken))
	assert.Equal(t, "Bearer", out.TokenType)
	assert.Equal(t, int64(AccessTokenTTL.Seconds()), out.ExpiresIn)
}

func TestExchangeCode_PKCEMismatch(t *testing.T) {
	svc, _, cdr, _ := newTestService()
	ctx := context.Background()
	code := &models.OAuthAuthorizationCode{
		ClientID: "pvoci_test123", UserID: 42,
		RedirectURI:         "https://claude.ai/api/mcp/auth_callback",
		CodeChallenge:       pkce.ComputeS256("other-verifier-must-be-at-least-43-chars"),
		CodeChallengeMethod: pkce.MethodS256,
		Scope:               "mcp:read", Resource: testResource,
		Policy: json.RawMessage(`{}`),
	}
	cdr.On("Consume", ctx, mock.AnythingOfType("string")).Return(code, nil)

	_, err := svc.ExchangeCode(ctx, ExchangeCodeInput{
		ClientID: "pvoci_test123", Code: "pvoac_x",
		RedirectURI:  "https://claude.ai/api/mcp/auth_callback",
		CodeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
	})
	assert.ErrorIs(t, err, ErrInvalidGrant)
}

func TestExchangeCode_Replay(t *testing.T) {
	svc, _, cdr, _ := newTestService()
	ctx := context.Background()
	cdr.On("Consume", ctx, mock.AnythingOfType("string")).Return(nil, repo.ErrNotFound)

	_, err := svc.ExchangeCode(ctx, ExchangeCodeInput{
		ClientID: "c", Code: "pvoac_x", RedirectURI: "https://x", CodeVerifier: "v",
	})
	assert.ErrorIs(t, err, ErrInvalidGrant, "consumed/expired код → invalid_grant")
}

// -----------------------------------------------------------------------------
// RefreshToken
// -----------------------------------------------------------------------------

func TestRefreshToken_Success(t *testing.T) {
	svc, cr, _, tr := newTestService()
	ctx := context.Background()
	refresh := &models.OAuthToken{
		ID: 100, TokenType: "refresh",
		ClientID: "pvoci_test123", UserID: 42,
		Scope: "mcp:read", Resource: testResource,
		Policy:    json.RawMessage(`{}`),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	tr.On("GetByHash", ctx, mock.AnythingOfType("string")).Return(refresh, nil)
	tr.On("Revoke", ctx, mock.AnythingOfType("string")).Return(nil)
	tr.On("Create", ctx, mock.AnythingOfType("*models.OAuthToken")).Return(nil).Times(2)
	cr.On("UpdateLastUsed", ctx, "pvoci_test123").Return(nil)

	out, err := svc.RefreshToken(ctx, RefreshTokenInput{
		ClientID: "pvoci_test123", RefreshToken: "pvort_raw",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out.AccessToken)
	assert.NotEmpty(t, out.RefreshToken)
}

func TestRefreshToken_ReplayDetected(t *testing.T) {
	svc, _, _, tr := newTestService()
	ctx := context.Background()
	revokedAt := time.Now()
	refresh := &models.OAuthToken{
		ID: 100, TokenType: "refresh",
		ClientID: "pvoci_test123", UserID: 42,
		Scope: "mcp:read", Resource: testResource,
		Policy:    json.RawMessage(`{}`),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		RevokedAt: &revokedAt, // уже revoked → replay
	}
	tr.On("GetByHash", ctx, mock.AnythingOfType("string")).Return(refresh, nil)
	tr.On("RevokeChain", ctx, uint(100)).Return(nil)

	_, err := svc.RefreshToken(ctx, RefreshTokenInput{
		ClientID: "pvoci_test123", RefreshToken: "pvort_raw",
	})
	assert.ErrorIs(t, err, ErrInvalidGrant)
	tr.AssertCalled(t, "RevokeChain", ctx, uint(100))
}

func TestRefreshToken_ClientMismatch(t *testing.T) {
	svc, _, _, tr := newTestService()
	ctx := context.Background()
	refresh := &models.OAuthToken{
		ID: 100, TokenType: "refresh",
		ClientID: "different-client", UserID: 42,
		Scope: "mcp:read", Resource: testResource,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	tr.On("GetByHash", ctx, mock.AnythingOfType("string")).Return(refresh, nil)

	_, err := svc.RefreshToken(ctx, RefreshTokenInput{
		ClientID: "pvoci_test123", RefreshToken: "pvort_raw",
	})
	assert.ErrorIs(t, err, ErrInvalidGrant, "refresh принадлежит другому client_id")
}

// -----------------------------------------------------------------------------
// ValidateAccessToken
// -----------------------------------------------------------------------------

func TestValidateAccessToken_Success(t *testing.T) {
	svc, _, _, tr := newTestService()
	ctx := context.Background()
	access := &models.OAuthToken{
		ID: 10, TokenType: "access", ClientID: "c", UserID: 42,
		Scope: "mcp:read mcp:write", Resource: testResource,
		Policy: json.RawMessage(`{"read_only":false}`),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	tr.On("GetByHash", ctx, mock.AnythingOfType("string")).Return(access, nil)

	v, err := svc.ValidateAccessToken(ctx, tokens.PrefixAccessToken+"somevalue")
	require.NoError(t, err)
	assert.Equal(t, uint(42), v.UserID)
	assert.Equal(t, "c", v.ClientID)
}

func TestValidateAccessToken_WrongPrefix(t *testing.T) {
	svc, _, _, _ := newTestService()
	_, err := svc.ValidateAccessToken(context.Background(), "pvlt_staticapikey")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateAccessToken_Expired(t *testing.T) {
	svc, _, _, tr := newTestService()
	ctx := context.Background()
	access := &models.OAuthToken{
		ID: 10, TokenType: "access", Resource: testResource,
		ExpiresAt: time.Now().Add(-1 * time.Second),
	}
	tr.On("GetByHash", ctx, mock.AnythingOfType("string")).Return(access, nil)

	_, err := svc.ValidateAccessToken(ctx, tokens.PrefixAccessToken+"x")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateAccessToken_Revoked(t *testing.T) {
	svc, _, _, tr := newTestService()
	ctx := context.Background()
	revokedAt := time.Now()
	access := &models.OAuthToken{
		ID: 10, TokenType: "access", Resource: testResource,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		RevokedAt: &revokedAt,
	}
	tr.On("GetByHash", ctx, mock.AnythingOfType("string")).Return(access, nil)

	_, err := svc.ValidateAccessToken(ctx, tokens.PrefixAccessToken+"x")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateAccessToken_AudienceMismatch(t *testing.T) {
	svc, _, _, tr := newTestService()
	ctx := context.Background()
	access := &models.OAuthToken{
		ID: 10, TokenType: "access",
		Resource: "https://other.example.com/mcp", // не наш ресурс
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	tr.On("GetByHash", ctx, mock.AnythingOfType("string")).Return(access, nil)

	_, err := svc.ValidateAccessToken(ctx, tokens.PrefixAccessToken+"x")
	assert.ErrorIs(t, err, ErrInvalidToken, "RFC 8707: токен должен быть для нашего resource")
}

// -----------------------------------------------------------------------------
// Revoke (RFC 7009)
// -----------------------------------------------------------------------------

func TestRevoke_Success(t *testing.T) {
	svc, _, _, tr := newTestService()
	ctx := context.Background()
	tr.On("Revoke", ctx, mock.AnythingOfType("string")).Return(nil)

	assert.NoError(t, svc.Revoke(ctx, "pvort_any"))
}

func TestRevoke_UnknownTokenReturns200(t *testing.T) {
	svc, _, _, tr := newTestService()
	ctx := context.Background()
	tr.On("Revoke", ctx, mock.AnythingOfType("string")).Return(repo.ErrNotFound)

	// RFC 7009 §2.2: unknown token → success (no oracle).
	err := svc.Revoke(ctx, "pvort_never-existed")
	assert.NoError(t, err)
}

// -----------------------------------------------------------------------------
// validateScope (internal helper — exercise через ExchangeCode indirectly)
// -----------------------------------------------------------------------------

func TestValidateScope_Helper(t *testing.T) {
	// positive: requested ⊂ allowed
	assert.NoError(t, validateScope("mcp:read", "mcp:read mcp:write"))
	assert.NoError(t, validateScope("", "mcp:read"))
	// negative: extra scope not in allowed
	err := validateScope("mcp:read admin", "mcp:read mcp:write")
	assert.True(t, errors.Is(err, ErrInvalidScope))
}
