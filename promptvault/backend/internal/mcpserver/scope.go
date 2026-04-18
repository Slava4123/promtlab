package mcpserver

import (
	"context"
	"fmt"

	apikeyuc "promptvault/internal/usecases/apikey"
)

type keyPolicyKeyType struct{}

// keyPolicyKey — ctx-ключ для *apikeyuc.KeyPolicy. Заполняется в APIKeyAuth,
// читается в enforceScope/enforceTeamID. Nil означает полный доступ
// (backward-compat для старых ключей до миграции 000035).
var keyPolicyKey = keyPolicyKeyType{}

// withKeyPolicy кладёт policy в ctx.
func withKeyPolicy(ctx context.Context, policy *apikeyuc.KeyPolicy) context.Context {
	return context.WithValue(ctx, keyPolicyKey, policy)
}

// GetKeyPolicy возвращает policy текущего запроса или nil.
func GetKeyPolicy(ctx context.Context) *apikeyuc.KeyPolicy {
	v, _ := ctx.Value(keyPolicyKey).(*apikeyuc.KeyPolicy)
	return v
}

// enforceScope проверяет право на вызов tool с учётом read-only и allowed_tools.
// Вызывается первой строкой каждого tool-handler. nil policy → разрешено.
func enforceScope(ctx context.Context, toolName string, isWrite bool) error {
	policy := GetKeyPolicy(ctx)
	if policy == nil {
		return nil
	}
	if isWrite && policy.ReadOnly {
		return fmt.Errorf("%w: tool %q requires write access, key is read-only", apikeyuc.ErrScopeDenied, toolName)
	}
	if !policy.IsToolAllowed(toolName) {
		return fmt.Errorf("%w: tool %q is not in key allowed_tools", apikeyuc.ErrScopeDenied, toolName)
	}
	return nil
}

// enforceTeamID проверяет, что team_id запроса совместим с team_id ключа.
// Правила:
//   - policy.TeamID == nil → пропускаем (ключ не привязан к команде);
//   - policy.TeamID != nil, input.TeamID == nil → deny (ключ только для команды);
//   - policy.TeamID != nil, input.TeamID != nil, mismatch → deny.
//
// Вызывается только tool'ами с TeamID в input (search, list, create и т.п.).
// Tool'ы по ID (get_prompt, update_prompt) пропускают эту проверку —
// ownership проверяет usecase по user_id ресурса.
func enforceTeamID(ctx context.Context, requestedTeamID *uint) error {
	policy := GetKeyPolicy(ctx)
	if policy == nil || policy.TeamID == nil {
		return nil
	}
	if requestedTeamID == nil {
		return fmt.Errorf("%w: key bound to team %d, request missing team_id", apikeyuc.ErrTeamMismatch, *policy.TeamID)
	}
	if *requestedTeamID != *policy.TeamID {
		return fmt.Errorf("%w: key bound to team %d, request team %d", apikeyuc.ErrTeamMismatch, *policy.TeamID, *requestedTeamID)
	}
	return nil
}
