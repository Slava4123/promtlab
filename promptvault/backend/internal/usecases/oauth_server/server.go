// Package oauth_server — OAuth 2.1 Authorization Server для внешних MCP-клиентов.
//
// Реализует (subset of OAuth 2.1 per MCP spec 2025-06-18 §Authorization):
//   - RFC 7591 Dynamic Client Registration (endpoint Register)
//   - Authorization Code Grant + PKCE S256 (endpoints Authorize + ExchangeCode)
//   - Refresh Token Grant с rotation + reuse detection (endpoint RefreshToken)
//   - RFC 7009 Token Revocation (endpoint Revoke)
//   - Access token validation для MCP middleware (endpoint ValidateAccessToken)
//
// Токены opaque (pvoat_/pvort_/pvoac_), SHA256 хэш в БД. Scope-policy храним
// в JSONB-поле (models.Policy), на момент issue копируем из defaults user'а.
package oauth_server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	"promptvault/internal/pkg/pkce"
	"promptvault/internal/pkg/tokens"
)

// Service инкапсулирует всю OAuth-логику. Репозитории инжектируются снаружи
// для тестируемости; конкретные реализации — в infrastructure/postgres/repository.
type Service struct {
	clients repo.OAuthClientRepository
	codes   repo.OAuthAuthorizationCodeRepository
	tokens  repo.OAuthTokenRepository

	// canonicalResource — RFC 8707 audience, к которому должны быть привязаны токены.
	// Обычно "https://promtlabs.ru/mcp". Устанавливается из cfg при старте.
	canonicalResource string
}

// NewService создаёт OAuth AS.
// canonicalResource — URL MCP-сервера, к которому issue'атся токены.
func NewService(
	clients repo.OAuthClientRepository,
	codes repo.OAuthAuthorizationCodeRepository,
	tokenRepo repo.OAuthTokenRepository,
	canonicalResource string,
) *Service {
	return &Service{
		clients:           clients,
		codes:             codes,
		tokens:            tokenRepo,
		canonicalResource: canonicalResource,
	}
}

// -----------------------------------------------------------------------------
// RFC 7591 — Dynamic Client Registration
// -----------------------------------------------------------------------------

// RegisterClient регистрирует нового OAuth-клиента.
// Для public clients (PKCE-only): TokenEndpointAuthMethod="none", client_secret не выдаётся.
func (s *Service) RegisterClient(ctx context.Context, in RegisterClientInput) (*RegisterClientOutput, error) {
	if strings.TrimSpace(in.ClientName) == "" {
		return nil, fmt.Errorf("%w: client_name required", ErrInvalidRequest)
	}
	if len(in.RedirectURIs) == 0 {
		return nil, fmt.Errorf("%w: redirect_uris required", ErrInvalidRequest)
	}
	for _, uri := range in.RedirectURIs {
		if err := validateRedirectURI(uri); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
		}
	}

	grantTypes := in.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{GrantTypeAuthCode, GrantTypeRefresh}
	}
	responseTypes := in.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []string{ResponseTypeCode}
	}
	authMethod := in.TokenEndpointAuthMethod
	if authMethod == "" {
		authMethod = TokenAuthMethodNone
	}
	scope := strings.TrimSpace(in.Scope)
	if scope == "" {
		scope = DefaultScope
	}

	clientIDRaw, _, err := tokens.New(tokens.PrefixClientID)
	if err != nil {
		return nil, fmt.Errorf("generate client_id: %w", err)
	}

	var clientSecretRaw, clientSecretHash string
	if authMethod != TokenAuthMethodNone {
		clientSecretRaw, clientSecretHash, err = tokens.New(tokens.PrefixClientSecret)
		if err != nil {
			return nil, fmt.Errorf("generate client_secret: %w", err)
		}
	}

	now := time.Now()
	client := &models.OAuthClient{
		ClientID:                clientIDRaw,
		ClientSecretHash:        clientSecretHash,
		ClientName:              in.ClientName,
		RedirectURIs:            in.RedirectURIs,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		TokenEndpointAuthMethod: authMethod,
		Scope:                   scope,
		IsDynamic:               true,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if err := s.clients.Create(ctx, client); err != nil {
		return nil, fmt.Errorf("store oauth client: %w", err)
	}

	slog.Info("oauth.client.registered",
		"client_id", clientIDRaw,
		"client_name", in.ClientName,
		"redirect_uris", len(in.RedirectURIs),
		"auth_method", authMethod,
	)

	return &RegisterClientOutput{
		ClientID:                clientIDRaw,
		ClientSecret:            clientSecretRaw,
		ClientIDIssuedAt:        now.Unix(),
		ClientName:              in.ClientName,
		RedirectURIs:            in.RedirectURIs,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		TokenEndpointAuthMethod: authMethod,
		Scope:                   scope,
		CreatedAt:               now,
	}, nil
}

// -----------------------------------------------------------------------------
// Authorization Code Grant — Authorize
// -----------------------------------------------------------------------------

// Authorize выдаёт authorization code. Вызывается из HTTP-хэндлера после того,
// как пользователь залогинен и одобрил consent-screen.
// Возвращает код + оригинальный state (прозрачно передаётся клиенту через redirect).
func (s *Service) Authorize(ctx context.Context, in AuthorizeInput) (*AuthorizeOutput, error) {
	if in.UserID == 0 {
		return nil, fmt.Errorf("%w: user not authenticated", ErrInvalidRequest)
	}
	if in.CodeChallenge == "" {
		return nil, ErrPKCERequired
	}
	if in.CodeChallengeMethod == "" {
		in.CodeChallengeMethod = CodeChallengeS256
	}
	if in.CodeChallengeMethod != CodeChallengeS256 {
		return nil, fmt.Errorf("%w: only S256 is supported", ErrInvalidRequest)
	}

	client, err := s.clients.GetByClientID(ctx, in.ClientID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrClientNotFound
		}
		return nil, fmt.Errorf("load client: %w", err)
	}

	if !slices.Contains(client.RedirectURIs, in.RedirectURI) {
		return nil, ErrInvalidRedirectURI
	}
	if !slices.Contains(client.ResponseTypes, ResponseTypeCode) {
		return nil, ErrUnsupportedResponseType
	}

	resource := in.Resource
	if resource == "" {
		resource = s.canonicalResource
	}
	if resource != s.canonicalResource {
		return nil, ErrResourceMismatch
	}

	scope := strings.TrimSpace(in.Scope)
	if scope == "" {
		scope = client.Scope
	}
	if err := validateScope(scope, client.Scope); err != nil {
		return nil, err
	}

	raw, hash, err := tokens.New(tokens.PrefixAuthCode)
	if err != nil {
		return nil, fmt.Errorf("generate code: %w", err)
	}

	// Policy — пока пустой JSON {}. В будущем сюда копируем scope-restrictions
	// из consent-screen (если пользователь выбрал «только чтение»).
	policy, _ := json.Marshal(models.Policy{})

	code := &models.OAuthAuthorizationCode{
		CodeHash:            hash,
		ClientID:            in.ClientID,
		UserID:              in.UserID,
		RedirectURI:         in.RedirectURI,
		CodeChallenge:       in.CodeChallenge,
		CodeChallengeMethod: in.CodeChallengeMethod,
		Scope:               scope,
		Resource:            resource,
		Policy:              policy,
		ExpiresAt:           time.Now().Add(AuthorizationCodeTTL),
		CreatedAt:           time.Now(),
	}
	if err := s.codes.Create(ctx, code); err != nil {
		return nil, fmt.Errorf("store code: %w", err)
	}

	slog.Info("oauth.code.issued",
		"client_id", in.ClientID,
		"user_id", in.UserID,
		"scope", scope,
	)

	return &AuthorizeOutput{Code: raw, State: in.State}, nil
}

// -----------------------------------------------------------------------------
// Authorization Code Grant — ExchangeCode
// -----------------------------------------------------------------------------

// ExchangeCode меняет authorization code на (access, refresh).
// Проверяет: client_id, PKCE verifier, redirect_uri match, resource.
func (s *Service) ExchangeCode(ctx context.Context, in ExchangeCodeInput) (*TokenResponse, error) {
	codeHash := tokens.Hash(in.Code)
	code, err := s.codes.Consume(ctx, codeHash)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrInvalidGrant
		}
		return nil, fmt.Errorf("consume code: %w", err)
	}

	if code.ClientID != in.ClientID {
		slog.Warn("oauth.code.client_mismatch", "client_id", in.ClientID, "code_client_id", code.ClientID)
		return nil, ErrInvalidGrant
	}
	if code.RedirectURI != in.RedirectURI {
		return nil, ErrInvalidRedirectURI
	}
	if in.Resource != "" && in.Resource != code.Resource {
		return nil, ErrResourceMismatch
	}

	if err := pkce.Verify(code.CodeChallengeMethod, code.CodeChallenge, in.CodeVerifier); err != nil {
		slog.Warn("oauth.pkce.mismatch", "client_id", in.ClientID, "user_id", code.UserID)
		return nil, ErrInvalidGrant
	}

	return s.issueTokenPair(ctx, code.ClientID, code.UserID, code.Scope, code.Resource, code.Policy, nil)
}

// -----------------------------------------------------------------------------
// Refresh Token Grant
// -----------------------------------------------------------------------------

// RefreshToken меняет refresh-токен на новую пару.
// При обнаружении replay'а (использование уже revoked refresh) — revoke всю
// цепочку потомков (breach detection per OAuth 2.1 §4.3.1).
func (s *Service) RefreshToken(ctx context.Context, in RefreshTokenInput) (*TokenResponse, error) {
	oldHash := tokens.Hash(in.RefreshToken)
	oldToken, err := s.tokens.GetByHash(ctx, oldHash)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrInvalidGrant
		}
		return nil, fmt.Errorf("load refresh: %w", err)
	}

	if oldToken.TokenType != "refresh" {
		return nil, ErrInvalidGrant
	}
	if oldToken.ClientID != in.ClientID {
		return nil, ErrInvalidGrant
	}
	if oldToken.RevokedAt != nil {
		// Replay detected — revoke whole chain.
		slog.Error("oauth.refresh.replay_detected",
			"token_id", oldToken.ID, "client_id", oldToken.ClientID, "user_id", oldToken.UserID)
		_ = s.tokens.RevokeChain(ctx, oldToken.ID)
		return nil, ErrInvalidGrant
	}
	if time.Now().After(oldToken.ExpiresAt) {
		return nil, ErrInvalidGrant
	}

	scope := in.Scope
	if scope == "" {
		scope = oldToken.Scope
	}
	if err := validateScope(scope, oldToken.Scope); err != nil {
		return nil, err
	}

	if in.Resource != "" && in.Resource != oldToken.Resource {
		return nil, ErrResourceMismatch
	}

	// Revoke старый refresh перед выдачей новой пары (rotation).
	if err := s.tokens.Revoke(ctx, oldHash); err != nil {
		return nil, fmt.Errorf("revoke old refresh: %w", err)
	}

	return s.issueTokenPair(ctx, oldToken.ClientID, oldToken.UserID, scope, oldToken.Resource, oldToken.Policy, &oldToken.ID)
}

// issueTokenPair создаёт access + refresh записи и возвращает TokenResponse.
// parentRefreshID != nil для refresh-rotation (строит цепочку).
func (s *Service) issueTokenPair(
	ctx context.Context,
	clientID string,
	userID uint,
	scope, resource string,
	policy json.RawMessage,
	parentRefreshID *uint,
) (*TokenResponse, error) {
	accessRaw, accessHash, err := tokens.New(tokens.PrefixAccessToken)
	if err != nil {
		return nil, fmt.Errorf("generate access: %w", err)
	}
	refreshRaw, refreshHash, err := tokens.New(tokens.PrefixRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("generate refresh: %w", err)
	}

	now := time.Now()
	access := &models.OAuthToken{
		TokenType: "access",
		TokenHash: accessHash,
		ClientID:  clientID,
		UserID:    userID,
		Scope:     scope,
		Resource:  resource,
		Policy:    policy,
		ExpiresAt: now.Add(AccessTokenTTL),
		CreatedAt: now,
	}
	if err := s.tokens.Create(ctx, access); err != nil {
		return nil, fmt.Errorf("store access: %w", err)
	}

	refresh := &models.OAuthToken{
		TokenType:     "refresh",
		TokenHash:     refreshHash,
		ClientID:      clientID,
		UserID:        userID,
		Scope:         scope,
		Resource:      resource,
		Policy:        policy,
		ExpiresAt:     now.Add(RefreshTokenTTL),
		ParentTokenID: parentRefreshID,
		CreatedAt:     now,
	}
	if err := s.tokens.Create(ctx, refresh); err != nil {
		return nil, fmt.Errorf("store refresh: %w", err)
	}

	_ = s.clients.UpdateLastUsed(ctx, clientID)

	slog.Info("oauth.token.exchanged",
		"client_id", clientID,
		"user_id", userID,
		"scope", scope,
		"rotated", parentRefreshID != nil,
	)

	return &TokenResponse{
		AccessToken:  accessRaw,
		TokenType:    "Bearer",
		ExpiresIn:    int64(AccessTokenTTL.Seconds()),
		RefreshToken: refreshRaw,
		Scope:        scope,
	}, nil
}

// -----------------------------------------------------------------------------
// RFC 7009 — Token Revocation
// -----------------------------------------------------------------------------

// Revoke — RFC 7009. Клиент присылает свой токен (access или refresh) → БД
// помечает revoked_at=NOW. Не-существующие токены → нет ошибки (security by obscurity
// всё равно не работает, но spec требует 200).
func (s *Service) Revoke(ctx context.Context, token string) error {
	hash := tokens.Hash(token)
	if err := s.tokens.Revoke(ctx, hash); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil // RFC 7009: неизвестные токены → 200.
		}
		return fmt.Errorf("revoke: %w", err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Access Token Validation (для MCP middleware)
// -----------------------------------------------------------------------------

// ValidateAccessToken проверяет Bearer-токен, пришедший на /mcp.
// Возвращает данные для rehydrate KeyPolicy в auth middleware.
func (s *Service) ValidateAccessToken(ctx context.Context, raw string) (*ValidatedAccessToken, error) {
	if !strings.HasPrefix(raw, tokens.PrefixAccessToken) {
		return nil, ErrInvalidToken
	}
	hash := tokens.Hash(raw)
	t, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("load token: %w", err)
	}
	if t.TokenType != "access" {
		return nil, ErrInvalidToken
	}
	if t.RevokedAt != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().After(t.ExpiresAt) {
		return nil, ErrInvalidToken
	}
	// RFC 8707 audience check — token должен быть для нас.
	if t.Resource != s.canonicalResource {
		slog.Warn("oauth.token.audience_mismatch",
			"token_id", t.ID, "expected", s.canonicalResource, "got", t.Resource)
		return nil, ErrInvalidToken
	}

	return &ValidatedAccessToken{
		UserID:   t.UserID,
		ClientID: t.ClientID,
		Scope:    t.Scope,
		Resource: t.Resource,
		Policy:   t.Policy,
	}, nil
}

// -----------------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------------

func validateRedirectURI(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URI: %w", err)
	}
	if u.Fragment != "" {
		return errors.New("redirect_uri must not contain fragment")
	}
	// Allow https:// всегда; http:// только для localhost (OAuth 2.1 §2.3.3).
	if u.Scheme != "https" && (u.Scheme != "http" || !isLoopback(u.Hostname())) {
		return errors.New("redirect_uri must use https:// (http:// only for localhost)")
	}
	return nil
}

func isLoopback(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	return false
}

// validateScope проверяет, что requested — подмножество allowed.
func validateScope(requested, allowed string) error {
	if requested == "" {
		return nil
	}
	allowedSet := make(map[string]struct{})
	for _, s := range strings.Fields(allowed) {
		allowedSet[s] = struct{}{}
	}
	for _, s := range strings.Fields(requested) {
		if _, ok := allowedSet[s]; !ok {
			return fmt.Errorf("%w: %q not in %q", ErrInvalidScope, s, allowed)
		}
	}
	return nil
}
