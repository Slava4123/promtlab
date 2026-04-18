package mcpserver

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	promptuc "promptvault/internal/usecases/prompt"
)

func newCompletionHandler(t *testing.T) (func(context.Context, *sdkmcp.CompleteRequest) (*sdkmcp.CompleteResult, error), *mockPromptSvc) {
	t.Helper()
	p := new(mockPromptSvc)
	return makeCompletionHandler(p), p
}

func makeCompleteReq(refType, refName, argName, argValue string, ctxArgs map[string]string) *sdkmcp.CompleteRequest {
	var cctx *sdkmcp.CompleteContext
	if ctxArgs != nil {
		cctx = &sdkmcp.CompleteContext{Arguments: ctxArgs}
	}
	return &sdkmcp.CompleteRequest{Params: &sdkmcp.CompleteParams{
		Ref:      &sdkmcp.CompleteReference{Type: refType, Name: refName},
		Argument: sdkmcp.CompleteParamsArgument{Name: argName, Value: argValue},
		Context:  cctx,
	}}
}

// --- ref/wrong type ---

func TestCompletion_WrongRefType(t *testing.T) {
	h, _ := newCompletionHandler(t)
	res, err := h(ctxWithUser(42), makeCompleteReq("ref/resource", "x", "id", "", nil))
	require.NoError(t, err)
	assert.Empty(t, res.Completion.Values)
}

func TestCompletion_WrongPromptName(t *testing.T) {
	h, _ := newCompletionHandler(t)
	res, err := h(ctxWithUser(42), makeCompleteReq("ref/prompt", "other", "id", "", nil))
	require.NoError(t, err)
	assert.Empty(t, res.Completion.Values)
}

// --- role completion ---

func TestCompletion_Role_Empty(t *testing.T) {
	h, _ := newCompletionHandler(t)
	res, err := h(ctxWithUser(42), makeCompleteReq("ref/prompt", "use_prompt", "role", "", nil))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"user", "assistant"}, res.Completion.Values)
}

func TestCompletion_Role_Prefix(t *testing.T) {
	h, _ := newCompletionHandler(t)
	res, err := h(ctxWithUser(42), makeCompleteReq("ref/prompt", "use_prompt", "role", "a", nil))
	require.NoError(t, err)
	assert.Equal(t, []string{"assistant"}, res.Completion.Values)
}

// --- id completion ---

func TestCompletion_ID_Success(t *testing.T) {
	h, p := newCompletionHandler(t)
	ctx := ctxWithUser(42)
	p.On("List", ctx, mock.MatchedBy(func(f repo.PromptListFilter) bool {
		return f.UserID == 42 && f.Query == "rev"
	})).Return([]models.Prompt{{ID: 7}, {ID: 42}}, int64(2), nil)

	res, err := h(ctx, makeCompleteReq("ref/prompt", "use_prompt", "id", "rev", nil))
	require.NoError(t, err)
	assert.Equal(t, []string{"7", "42"}, res.Completion.Values)
	assert.False(t, res.Completion.HasMore)
}

func TestCompletion_ID_ListError_ReturnsEmpty(t *testing.T) {
	h, p := newCompletionHandler(t)
	ctx := ctxWithUser(42)
	p.On("List", ctx, mock.Anything).Return([]models.Prompt{}, int64(0), errors.New("db down"))

	res, err := h(ctx, makeCompleteReq("ref/prompt", "use_prompt", "id", "x", nil))
	require.NoError(t, err)
	assert.Empty(t, res.Completion.Values)
}

// --- vars completion ---

func TestCompletion_Vars_NoContext(t *testing.T) {
	h, _ := newCompletionHandler(t)
	res, err := h(ctxWithUser(42), makeCompleteReq("ref/prompt", "use_prompt", "vars", "", nil))
	require.NoError(t, err)
	assert.Empty(t, res.Completion.Values)
}

func TestCompletion_Vars_NoIDInContext(t *testing.T) {
	h, _ := newCompletionHandler(t)
	res, err := h(ctxWithUser(42), makeCompleteReq("ref/prompt", "use_prompt", "vars", "", map[string]string{}))
	require.NoError(t, err)
	assert.Empty(t, res.Completion.Values)
}

func TestCompletion_Vars_InvalidID(t *testing.T) {
	h, _ := newCompletionHandler(t)
	res, err := h(ctxWithUser(42), makeCompleteReq("ref/prompt", "use_prompt", "vars", "",
		map[string]string{"id": "abc"}))
	require.NoError(t, err)
	assert.Empty(t, res.Completion.Values)
}

func TestCompletion_Vars_NotFound_NoLeak(t *testing.T) {
	h, p := newCompletionHandler(t)
	ctx := ctxWithUser(42)
	// Эмулируем чужой промпт: usecase возвращает ErrNotFound.
	p.On("GetByID", ctx, uint(99), uint(42)).Return(nil, promptuc.ErrNotFound)

	res, err := h(ctx, makeCompleteReq("ref/prompt", "use_prompt", "vars", "",
		map[string]string{"id": "99"}))
	require.NoError(t, err)
	// Клиент НЕ должен различить 404 от "no vars" — оба дают пустой ответ.
	assert.Empty(t, res.Completion.Values)
}

func TestCompletion_Vars_Happy(t *testing.T) {
	h, p := newCompletionHandler(t)
	ctx := ctxWithUser(42)
	prompt := samplePrompt()
	prompt.Content = "Write {{lang}} code for {{task}}"
	p.On("GetByID", ctx, uint(1), uint(42)).Return(prompt, nil)

	res, err := h(ctx, makeCompleteReq("ref/prompt", "use_prompt", "vars", "",
		map[string]string{"id": "1"}))
	require.NoError(t, err)
	require.Len(t, res.Completion.Values, 1)
	skeleton := res.Completion.Values[0]
	assert.Contains(t, skeleton, `"lang"`)
	assert.Contains(t, skeleton, `"task"`)
}

func TestCompletion_Vars_NoVarsInPrompt(t *testing.T) {
	h, p := newCompletionHandler(t)
	ctx := ctxWithUser(42)
	prompt := samplePrompt()
	prompt.Content = "no variables here"
	p.On("GetByID", ctx, uint(1), uint(42)).Return(prompt, nil)

	res, err := h(ctx, makeCompleteReq("ref/prompt", "use_prompt", "vars", "",
		map[string]string{"id": "1"}))
	require.NoError(t, err)
	assert.Empty(t, res.Completion.Values)
}
