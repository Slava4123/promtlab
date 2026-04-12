package mcpserver

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	promptuc "promptvault/internal/usecases/prompt"
	searchuc "promptvault/internal/usecases/search"
)

// --- helpers ---

func newToolHandlers() (*toolHandlers, *mockPromptSvc, *mockCollectionSvc, *mockTagSvc, *mockSearchSvc) {
	p := new(mockPromptSvc)
	c := new(mockCollectionSvc)
	tg := new(mockTagSvc)
	s := new(mockSearchSvc)
	return &toolHandlers{prompts: p, collections: c, tags: tg, search: s}, p, c, tg, s
}

func samplePrompt() *models.Prompt {
	return &models.Prompt{
		ID: 1, Title: "Test", Content: "Hello", Model: "claude",
		Favorite: true, UsageCount: 5,
		Tags:        []models.Tag{{ID: 1, Name: "go", Color: "#00ff00"}},
		Collections: []models.Collection{{ID: 1, Name: "Dev", Color: "#8b5cf6"}},
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
}

func parseJSON(t *testing.T, result string) map[string]any {
	t.Helper()
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(result), &data))
	return data
}

// --- search ---

func TestSearchPrompts(t *testing.T) {
	h, _, _, _, s := newToolHandlers()
	ctx := ctxWithUser(42)

	output := &searchuc.SearchOutput{
		Prompts: []searchuc.SearchResult{{ID: 1, Title: "Test"}},
	}
	s.On("Search", ctx, uint(42), (*uint)(nil), "test").Return(output, nil)

	res, _, err := h.searchPrompts(ctx, nil, SearchInput{Query: "test"})
	require.NoError(t, err)
	assert.Len(t, res.Content, 1)
}

func TestSearchPrompts_Error(t *testing.T) {
	h, _, _, _, s := newToolHandlers()
	ctx := ctxWithUser(42)

	s.On("Search", ctx, uint(42), (*uint)(nil), "q").Return(nil, assert.AnError)

	_, _, err := h.searchPrompts(ctx, nil, SearchInput{Query: "q"})
	assert.Error(t, err)
	assert.Equal(t, "internal server error", err.Error())
}

// --- list prompts ---

func TestListPrompts(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	prompts := []models.Prompt{*samplePrompt()}
	p.On("List", ctx, mock.Anything).Return(prompts, int64(1), nil)

	res, _, err := h.listPrompts(ctx, nil, ListPromptsInput{})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, float64(1), data["total"])
}

func TestListPrompts_DefaultPageSize(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("List", ctx, mock.MatchedBy(func(f repo.PromptListFilter) bool {
		return f.PageSize == 20
	})).Return([]models.Prompt{}, int64(0), nil)

	_, _, err := h.listPrompts(ctx, nil, ListPromptsInput{PageSize: 0})
	require.NoError(t, err)
	p.AssertExpectations(t)
}

func TestListPrompts_MaxPageSize(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("List", ctx, mock.MatchedBy(func(f repo.PromptListFilter) bool {
		return f.PageSize == 100
	})).Return([]models.Prompt{}, int64(0), nil)

	_, _, err := h.listPrompts(ctx, nil, ListPromptsInput{PageSize: 999})
	require.NoError(t, err)
	p.AssertExpectations(t)
}

// --- get prompt ---

func TestGetPrompt(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	prompt := samplePrompt()
	p.On("GetByID", ctx, uint(1), uint(42)).Return(prompt, nil)

	res, _, err := h.getPrompt(ctx, nil, GetPromptInput{ID: 1})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "Test", data["title"])
}

func TestGetPrompt_NotFound(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("GetByID", ctx, uint(99), uint(42)).Return(nil, promptuc.ErrNotFound)

	_, _, err := h.getPrompt(ctx, nil, GetPromptInput{ID: 99})
	assert.EqualError(t, err, "prompt not found")
}

// --- list collections ---

func TestListCollections(t *testing.T) {
	h, _, c, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	colls := []models.CollectionWithCount{
		{Collection: models.Collection{ID: 1, Name: "Dev", Color: "#8b5cf6"}, PromptCount: 3},
	}
	c.On("List", ctx, uint(42), ([]uint)(nil)).Return(colls, nil)

	res, _, err := h.listCollections(ctx, nil, ListCollectionsInput{})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	collections := data["collections"].([]any)
	assert.Len(t, collections, 1)
}

// --- list tags ---

func TestListTags(t *testing.T) {
	h, _, _, tg, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	tags := []models.Tag{{ID: 1, Name: "go", Color: "#00ff00"}}
	tg.On("List", ctx, uint(42), (*uint)(nil)).Return(tags, nil)

	res, _, err := h.listTags(ctx, nil, ListTagsInput{})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	tagList := data["tags"].([]any)
	assert.Len(t, tagList, 1)
}

// --- get prompt versions ---

func TestGetPromptVersions(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	versions := []models.PromptVersion{
		{ID: 1, VersionNumber: 1, Title: "v1", Content: "hello", CreatedAt: time.Now()},
	}
	p.On("ListVersions", ctx, uint(1), uint(42), 0, 20).Return(versions, int64(1), nil)

	res, _, err := h.getPromptVersions(ctx, nil, GetVersionsInput{PromptID: 1})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, float64(1), data["total"])
}

func TestGetPromptVersions_MaxPageSize(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("ListVersions", ctx, uint(1), uint(42), 0, 100).Return([]models.PromptVersion{}, int64(0), nil)

	_, _, err := h.getPromptVersions(ctx, nil, GetVersionsInput{PromptID: 1, PageSize: 500})
	require.NoError(t, err)
	p.AssertExpectations(t)
}

// --- create prompt ---

func TestCreatePrompt(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	prompt := samplePrompt()
	p.On("Create", ctx, mock.Anything).Return(prompt, nil)

	res, _, err := h.createPrompt(ctx, nil, CreatePromptInput{Title: "Test", Content: "Hello"})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, "Test", data["title"])
}

// --- update prompt ---

func TestUpdatePrompt(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	title := "Updated"
	prompt := samplePrompt()
	prompt.Title = title
	p.On("Update", ctx, uint(1), uint(42), mock.Anything).Return(prompt, nil)

	res, _, err := h.updatePrompt(ctx, nil, UpdatePromptInput{ID: 1, Title: &title})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, "Updated", data["title"])
}

// --- delete prompt ---

func TestDeletePrompt(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("Delete", ctx, uint(1), uint(42)).Return(nil)

	res, _, err := h.deletePrompt(ctx, nil, DeletePromptInput{ID: 1})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, "deleted", data["status"])
}

func TestDeletePrompt_NotFound(t *testing.T) {
	h, p, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("Delete", ctx, uint(99), uint(42)).Return(promptuc.ErrNotFound)

	_, _, err := h.deletePrompt(ctx, nil, DeletePromptInput{ID: 99})
	assert.EqualError(t, err, "prompt not found")
}

// --- create tag ---

func TestCreateTag(t *testing.T) {
	h, _, _, tg, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	tag := &models.Tag{ID: 1, Name: "go", Color: "#00ff00"}
	tg.On("Create", ctx, "go", "#00ff00", uint(42), (*uint)(nil)).Return(tag, nil)

	res, _, err := h.createTag(ctx, nil, CreateTagInput{Name: "go", Color: "#00ff00"})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, "go", data["name"])
}

// --- create collection ---

func TestCreateCollection(t *testing.T) {
	h, _, c, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	coll := &models.Collection{ID: 1, Name: "Dev", Color: "#8b5cf6"}
	c.On("Create", ctx, uint(42), "Dev", "", "#8b5cf6", "", (*uint)(nil)).Return(coll, nil)

	res, _, err := h.createCollection(ctx, nil, CreateCollectionInput{Name: "Dev", Color: "#8b5cf6"})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, "Dev", data["name"])
}

// --- delete collection ---

func TestDeleteCollection(t *testing.T) {
	h, _, c, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	c.On("Delete", ctx, uint(1), uint(42)).Return(nil)

	res, _, err := h.deleteCollection(ctx, nil, DeleteCollectionInput{ID: 1})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, "deleted", data["status"])
}
