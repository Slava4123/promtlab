package mcpserver

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"promptvault/internal/models"
	promptuc "promptvault/internal/usecases/prompt"
)

func newResourceHandlers() (*resourceHandlers, *mockPromptSvc, *mockCollectionSvc, *mockTagSvc) {
	p := new(mockPromptSvc)
	c := new(mockCollectionSvc)
	tg := new(mockTagSvc)
	return &resourceHandlers{prompts: p, collections: c, tags: tg}, p, c, tg
}

// --- readCollections ---

func TestReadCollections(t *testing.T) {
	h, _, c, _ := newResourceHandlers()
	ctx := ctxWithUser(42)

	colls := []models.CollectionWithCount{
		{Collection: models.Collection{ID: 1, Name: "Dev", Color: "#8b5cf6"}, PromptCount: 3},
		{Collection: models.Collection{ID: 2, Name: "Work", Color: "#ef4444"}, PromptCount: 7},
	}
	c.On("List", ctx, uint(42), ([]uint)(nil)).Return(colls, nil)

	req := &sdkmcp.ReadResourceRequest{Params: &sdkmcp.ReadResourceParams{}}
	res, err := h.readCollections(ctx, req)
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	assert.Equal(t, "promptvault://collections", res.Contents[0].URI)

	var result []CollectionWithCountResponse
	require.NoError(t, json.Unmarshal([]byte(res.Contents[0].Text), &result))
	assert.Len(t, result, 2)
	assert.Equal(t, "Dev", result[0].Name)
	assert.Equal(t, int64(3), result[0].PromptCount)
}

// --- readTags ---

func TestReadTags(t *testing.T) {
	h, _, _, tg := newResourceHandlers()
	ctx := ctxWithUser(42)

	tags := []models.Tag{
		{ID: 1, Name: "go", Color: "#00ff00"},
		{ID: 2, Name: "react", Color: "#61dafb"},
	}
	tg.On("List", ctx, uint(42), (*uint)(nil)).Return(tags, nil)

	req := &sdkmcp.ReadResourceRequest{Params: &sdkmcp.ReadResourceParams{}}
	res, err := h.readTags(ctx, req)
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	assert.Equal(t, "promptvault://tags", res.Contents[0].URI)

	var result []TagResponse
	require.NoError(t, json.Unmarshal([]byte(res.Contents[0].Text), &result))
	assert.Len(t, result, 2)
	assert.Equal(t, "go", result[0].Name)
}

// --- readPrompt ---

func TestReadPrompt(t *testing.T) {
	h, p, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)

	prompt := samplePrompt()
	p.On("GetByID", ctx, uint(1), uint(42)).Return(prompt, nil)

	req := &sdkmcp.ReadResourceRequest{Params: &sdkmcp.ReadResourceParams{URI: "promptvault://prompts/1"}}

	res, err := h.readPrompt(ctx, req)
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	assert.Equal(t, "promptvault://prompts/1", res.Contents[0].URI)

	var result PromptResponse
	require.NoError(t, json.Unmarshal([]byte(res.Contents[0].Text), &result))
	assert.Equal(t, "Test", result.Title)
	assert.Equal(t, uint(1), result.ID)
}

func TestReadPrompt_NotFound(t *testing.T) {
	h, p, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)

	p.On("GetByID", ctx, uint(99), uint(42)).Return(nil, promptuc.ErrNotFound)

	req := &sdkmcp.ReadResourceRequest{Params: &sdkmcp.ReadResourceParams{URI: "promptvault://prompts/99"}}

	_, err := h.readPrompt(ctx, req)
	assert.Error(t, err)
}

func TestReadPrompt_InvalidURI(t *testing.T) {
	h, _, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)

	req := &sdkmcp.ReadResourceRequest{Params: &sdkmcp.ReadResourceParams{URI: "promptvault://prompts/abc"}}

	_, err := h.readPrompt(ctx, req)
	assert.Error(t, err)
}

// --- usePrompt ---

func TestUsePrompt(t *testing.T) {
	h, p, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)

	prompt := samplePrompt()
	p.On("GetByID", ctx, uint(1), uint(42)).Return(prompt, nil)

	req := &sdkmcp.GetPromptRequest{Params: &sdkmcp.GetPromptParams{Arguments: map[string]string{"id": "1"}}}

	res, err := h.usePrompt(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, res.Description, "Test")
	require.Len(t, res.Messages, 1)

	text := res.Messages[0].Content.(*sdkmcp.TextContent).Text
	assert.Contains(t, text, "# Test")
	assert.Contains(t, text, "Hello")
}

func TestUsePrompt_InvalidID(t *testing.T) {
	h, _, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)

	req := &sdkmcp.GetPromptRequest{Params: &sdkmcp.GetPromptParams{Arguments: map[string]string{"id": "abc"}}}

	_, err := h.usePrompt(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid prompt id")
}

// --- usePrompt: vars, role, size, missing (HIGH-5) ---

func promptWithVars(content string) *models.Prompt {
	p := samplePrompt()
	p.Content = content
	return p
}

func TestUsePrompt_ZeroIDRejected(t *testing.T) {
	h, _, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)

	req := &sdkmcp.GetPromptRequest{Params: &sdkmcp.GetPromptParams{Arguments: map[string]string{"id": "0"}}}
	_, err := h.usePrompt(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid prompt id")
}

func TestUsePrompt_InvalidRole(t *testing.T) {
	h, p, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)
	p.On("GetByID", ctx, uint(1), uint(42)).Return(samplePrompt(), nil).Maybe()

	req := &sdkmcp.GetPromptRequest{Params: &sdkmcp.GetPromptParams{Arguments: map[string]string{
		"id": "1", "role": "system",
	}}}
	_, err := h.usePrompt(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestUsePrompt_RoleCaseInsensitive(t *testing.T) {
	h, p, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)
	p.On("GetByID", ctx, uint(1), uint(42)).Return(samplePrompt(), nil)

	req := &sdkmcp.GetPromptRequest{Params: &sdkmcp.GetPromptParams{Arguments: map[string]string{
		"id": "1", "role": "ASSISTANT",
	}}}
	res, err := h.usePrompt(ctx, req)
	require.NoError(t, err)
	require.Len(t, res.Messages, 1)
	assert.Equal(t, sdkmcp.Role("assistant"), res.Messages[0].Role)
}

func TestUsePrompt_VarsNotJSON(t *testing.T) {
	h, p, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)
	p.On("GetByID", ctx, uint(1), uint(42)).Return(samplePrompt(), nil).Maybe()

	req := &sdkmcp.GetPromptRequest{Params: &sdkmcp.GetPromptParams{Arguments: map[string]string{
		"id": "1", "vars": "[1,2,3]",
	}}}
	_, err := h.usePrompt(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JSON object")
}

func TestUsePrompt_VarsTooLarge(t *testing.T) {
	h, p, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)
	p.On("GetByID", ctx, uint(1), uint(42)).Return(samplePrompt(), nil).Maybe()

	big := strings.Repeat("x", maxVarsJSONSize+10)
	req := &sdkmcp.GetPromptRequest{Params: &sdkmcp.GetPromptParams{Arguments: map[string]string{
		"id": "1", "vars": big,
	}}}
	_, err := h.usePrompt(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceed")
}

func TestUsePrompt_MissingVarsReturnsHint(t *testing.T) {
	h, p, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)
	p.On("GetByID", ctx, uint(1), uint(42)).Return(
		promptWithVars("Hello {{name}} and {{lang}}"), nil,
	)

	req := &sdkmcp.GetPromptRequest{Params: &sdkmcp.GetPromptParams{Arguments: map[string]string{"id": "1"}}}
	_, err := h.usePrompt(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires variables")
	assert.Contains(t, err.Error(), "name")
	assert.Contains(t, err.Error(), "lang")
}

func TestUsePrompt_VarsRendered(t *testing.T) {
	h, p, _, _ := newResourceHandlers()
	ctx := ctxWithUser(42)
	p.On("GetByID", ctx, uint(1), uint(42)).Return(
		promptWithVars("Write {{lang}} code"), nil,
	)

	req := &sdkmcp.GetPromptRequest{Params: &sdkmcp.GetPromptParams{Arguments: map[string]string{
		"id": "1", "vars": `{"lang":"Go"}`,
	}}}
	res, err := h.usePrompt(ctx, req)
	require.NoError(t, err)
	require.Len(t, res.Messages, 1)
	text := res.Messages[0].Content.(*sdkmcp.TextContent).Text
	assert.Contains(t, text, "Write Go code")
	assert.NotContains(t, text, "{{")
}
