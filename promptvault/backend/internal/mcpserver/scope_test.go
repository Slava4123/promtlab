package mcpserver

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apikeyuc "promptvault/internal/usecases/apikey"
)

func ctxWith(policy *apikeyuc.KeyPolicy) context.Context {
	return withKeyPolicy(context.Background(), policy)
}

func uintPtr(v uint) *uint { return &v }

func TestEnforceScope_NilPolicyAllowsAll(t *testing.T) {
	ctx := context.Background() // без policy
	assert.NoError(t, enforceScope(ctx, "create_prompt", true))
	assert.NoError(t, enforceScope(ctx, "list_prompts", false))
}

func TestEnforceScope_ReadOnlyBlocksWrite(t *testing.T) {
	ctx := ctxWith(&apikeyuc.KeyPolicy{ReadOnly: true})
	err := enforceScope(ctx, "create_prompt", true)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apikeyuc.ErrScopeDenied))
}

func TestEnforceScope_ReadOnlyAllowsRead(t *testing.T) {
	ctx := ctxWith(&apikeyuc.KeyPolicy{ReadOnly: true})
	assert.NoError(t, enforceScope(ctx, "list_prompts", false))
}

func TestEnforceScope_AllowedToolsBlocks(t *testing.T) {
	ctx := ctxWith(&apikeyuc.KeyPolicy{AllowedTools: []string{"get_prompt"}})
	err := enforceScope(ctx, "list_prompts", false)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apikeyuc.ErrScopeDenied))
}

func TestEnforceScope_AllowedToolsAllows(t *testing.T) {
	ctx := ctxWith(&apikeyuc.KeyPolicy{AllowedTools: []string{"get_prompt", "list_prompts"}})
	assert.NoError(t, enforceScope(ctx, "list_prompts", false))
	assert.NoError(t, enforceScope(ctx, "get_prompt", false))
}

func TestEnforceScope_AllowedToolsEmptyMeansAll(t *testing.T) {
	ctx := ctxWith(&apikeyuc.KeyPolicy{AllowedTools: nil})
	assert.NoError(t, enforceScope(ctx, "list_prompts", false))
	assert.NoError(t, enforceScope(ctx, "create_prompt", true))
}

func TestEnforceTeamID_NilPolicyAllows(t *testing.T) {
	ctx := context.Background()
	assert.NoError(t, enforceTeamID(ctx, uintPtr(42)))
	assert.NoError(t, enforceTeamID(ctx, nil))
}

func TestEnforceTeamID_PolicyTeamNilAllows(t *testing.T) {
	ctx := ctxWith(&apikeyuc.KeyPolicy{TeamID: nil})
	assert.NoError(t, enforceTeamID(ctx, uintPtr(42)))
	assert.NoError(t, enforceTeamID(ctx, nil))
}

func TestEnforceTeamID_Match(t *testing.T) {
	ctx := ctxWith(&apikeyuc.KeyPolicy{TeamID: uintPtr(42)})
	assert.NoError(t, enforceTeamID(ctx, uintPtr(42)))
}

func TestEnforceTeamID_Mismatch(t *testing.T) {
	ctx := ctxWith(&apikeyuc.KeyPolicy{TeamID: uintPtr(42)})
	err := enforceTeamID(ctx, uintPtr(99))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apikeyuc.ErrTeamMismatch))
}

func TestEnforceTeamID_PersonalWhenTeamRequired(t *testing.T) {
	ctx := ctxWith(&apikeyuc.KeyPolicy{TeamID: uintPtr(42)})
	err := enforceTeamID(ctx, nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apikeyuc.ErrTeamMismatch))
}
