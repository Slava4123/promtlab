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
	colluc "promptvault/internal/usecases/collection"
	promptuc "promptvault/internal/usecases/prompt"
	searchuc "promptvault/internal/usecases/search"
	shareuc "promptvault/internal/usecases/share"
	taguc "promptvault/internal/usecases/tag"
)

// --- helpers ---

func newToolHandlers() (*toolHandlers, *mockPromptSvc, *mockCollectionSvc, *mockTagSvc, *mockSearchSvc, *mockShareSvc) {
	p := new(mockPromptSvc)
	c := new(mockCollectionSvc)
	tg := new(mockTagSvc)
	s := new(mockSearchSvc)
	sh := new(mockShareSvc)
	return &toolHandlers{prompts: p, collections: c, tags: tg, search: s, shares: sh}, p, c, tg, s, sh
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
	h, _, _, _, s, _ := newToolHandlers()
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
	h, _, _, _, s, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	s.On("Search", ctx, uint(42), (*uint)(nil), "q").Return(nil, assert.AnError)

	_, _, err := h.searchPrompts(ctx, nil, SearchInput{Query: "q"})
	assert.Error(t, err)
	assert.Equal(t, "internal server error", err.Error())
}

// --- list prompts ---

func TestListPrompts(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
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
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("List", ctx, mock.MatchedBy(func(f repo.PromptListFilter) bool {
		return f.PageSize == 20
	})).Return([]models.Prompt{}, int64(0), nil)

	_, _, err := h.listPrompts(ctx, nil, ListPromptsInput{PageSize: 0})
	require.NoError(t, err)
	p.AssertExpectations(t)
}

func TestListPrompts_MaxPageSize(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
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
	h, p, _, _, _, _ := newToolHandlers()
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
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("GetByID", ctx, uint(99), uint(42)).Return(nil, promptuc.ErrNotFound)

	_, _, err := h.getPrompt(ctx, nil, GetPromptInput{ID: 99})
	assert.EqualError(t, err, "prompt not found")
}

// --- list collections ---

func TestListCollections(t *testing.T) {
	h, _, c, _, _, _ := newToolHandlers()
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
	h, _, _, tg, _, _ := newToolHandlers()
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
	h, p, _, _, _, _ := newToolHandlers()
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
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("ListVersions", ctx, uint(1), uint(42), 0, 100).Return([]models.PromptVersion{}, int64(0), nil)

	_, _, err := h.getPromptVersions(ctx, nil, GetVersionsInput{PromptID: 1, PageSize: 500})
	require.NoError(t, err)
	p.AssertExpectations(t)
}

// --- create prompt ---

func TestCreatePrompt(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
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
	h, p, _, _, _, _ := newToolHandlers()
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
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("Delete", ctx, uint(1), uint(42)).Return(nil)

	res, _, err := h.deletePrompt(ctx, nil, DeletePromptInput{ID: 1})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, "deleted", data["status"])
}

func TestDeletePrompt_NotFound(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("Delete", ctx, uint(99), uint(42)).Return(promptuc.ErrNotFound)

	_, _, err := h.deletePrompt(ctx, nil, DeletePromptInput{ID: 99})
	assert.EqualError(t, err, "prompt not found")
}

// --- create tag ---

func TestCreateTag(t *testing.T) {
	h, _, _, tg, _, _ := newToolHandlers()
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
	h, _, c, _, _, _ := newToolHandlers()
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
	h, _, c, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	c.On("Delete", ctx, uint(1), uint(42)).Return(nil)

	res, _, err := h.deleteCollection(ctx, nil, DeleteCollectionInput{ID: 1})
	require.NoError(t, err)

	text := res.Content[0].(*sdkmcp.TextContent).Text
	data := parseJSON(t, text)
	assert.Equal(t, "deleted", data["status"])
}

// ===== NEW TOOL TESTS =====

// --- prompt_favorite ---

func TestPromptFavorite(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	prompt := samplePrompt()
	prompt.Favorite = true
	p.On("ToggleFavorite", ctx, uint(1), uint(42)).Return(prompt, nil)

	res, _, err := h.promptFavorite(ctx, nil, PromptIDInput{ID: 1})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	assert.Equal(t, true, data["favorite"])
}

func TestPromptFavorite_NotFound(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("ToggleFavorite", ctx, uint(99), uint(42)).Return(nil, promptuc.ErrNotFound)

	_, _, err := h.promptFavorite(ctx, nil, PromptIDInput{ID: 99})
	assert.EqualError(t, err, "prompt not found")
}

// --- prompt_pin ---

func TestPromptPin(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("TogglePin", ctx, promptuc.PinInput{PromptID: 1, UserID: 42, TeamWide: true}).
		Return(&promptuc.PinResult{Pinned: true, TeamWide: true}, nil)

	res, _, err := h.promptPin(ctx, nil, PromptPinInput{ID: 1, TeamWide: true})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	assert.Equal(t, true, data["pinned"])
	assert.Equal(t, true, data["team_wide"])
}

func TestPromptPin_Forbidden(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("TogglePin", ctx, mock.Anything).Return(nil, promptuc.ErrPinForbidden)

	_, _, err := h.promptPin(ctx, nil, PromptPinInput{ID: 1, TeamWide: true})
	assert.EqualError(t, err, "pin forbidden for viewers")
}

// --- prompt_list_pinned ---

func TestPromptListPinned(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	prompts := []models.Prompt{*samplePrompt(), *samplePrompt()}
	p.On("ListPinned", ctx, uint(42), (*uint)(nil), 10).Return(prompts, nil)

	res, _, err := h.promptListPinned(ctx, nil, ListLimitedInput{})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	list := data["prompts"].([]any)
	assert.Len(t, list, 2)
}

func TestPromptListPinned_Empty(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("ListPinned", ctx, uint(42), (*uint)(nil), 10).Return([]models.Prompt{}, nil)

	res, _, err := h.promptListPinned(ctx, nil, ListLimitedInput{})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	list := data["prompts"].([]any)
	assert.Len(t, list, 0)
}

// --- prompt_list_recent ---

func TestPromptListRecent(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	prompts := []models.Prompt{*samplePrompt()}
	p.On("ListRecent", ctx, uint(42), (*uint)(nil), 10).Return(prompts, nil)

	res, _, err := h.promptListRecent(ctx, nil, ListLimitedInput{})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	list := data["prompts"].([]any)
	assert.Len(t, list, 1)
}

func TestPromptListRecent_Empty(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("ListRecent", ctx, uint(42), (*uint)(nil), 10).Return([]models.Prompt{}, nil)

	res, _, err := h.promptListRecent(ctx, nil, ListLimitedInput{})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	list := data["prompts"].([]any)
	assert.Len(t, list, 0)
}

// --- prompt_revert ---

func TestPromptRevert(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	prompt := samplePrompt()
	p.On("RevertToVersion", ctx, uint(1), uint(42), uint(5)).Return(prompt, nil)

	res, _, err := h.promptRevert(ctx, nil, RevertInput{PromptID: 1, VersionID: 5})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	assert.Equal(t, "Test", data["title"])
}

func TestPromptRevert_VersionNotFound(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("RevertToVersion", ctx, uint(1), uint(42), uint(99)).Return(nil, promptuc.ErrVersionNotFound)

	_, _, err := h.promptRevert(ctx, nil, RevertInput{PromptID: 1, VersionID: 99})
	assert.EqualError(t, err, "version not found")
}

// --- prompt_increment_usage ---

func TestPromptIncrementUsage(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("IncrementUsage", ctx, uint(1), uint(42)).Return(nil)

	res, _, err := h.promptIncrementUsage(ctx, nil, PromptIDInput{ID: 1})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	assert.Equal(t, "recorded", data["status"])
}

func TestPromptIncrementUsage_NotFound(t *testing.T) {
	h, p, _, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	p.On("IncrementUsage", ctx, uint(99), uint(42)).Return(promptuc.ErrNotFound)

	_, _, err := h.promptIncrementUsage(ctx, nil, PromptIDInput{ID: 99})
	assert.EqualError(t, err, "prompt not found")
}

// --- share_create ---

func TestShareCreate(t *testing.T) {
	h, _, _, _, _, sh := newToolHandlers()
	ctx := ctxWithUser(42)

	link := &shareuc.ShareLinkInfo{
		ID: 1, Token: "abc123", URL: "https://promtlabs.ru/s/abc123",
		IsActive: true, ViewCount: 0, CreatedAt: time.Now(),
	}
	sh.On("CreateOrGet", ctx, uint(1), uint(42)).Return(link, true, nil)

	res, _, err := h.shareCreate(ctx, nil, ShareCreateInput{PromptID: 1})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	assert.Equal(t, "abc123", data["token"])
	assert.Equal(t, true, data["is_active"])
}

func TestShareCreate_ViewerReadOnly(t *testing.T) {
	h, _, _, _, _, sh := newToolHandlers()
	ctx := ctxWithUser(42)

	sh.On("CreateOrGet", ctx, uint(1), uint(42)).Return(nil, false, shareuc.ErrViewerReadOnly)

	_, _, err := h.shareCreate(ctx, nil, ShareCreateInput{PromptID: 1})
	assert.EqualError(t, err, "read-only access")
}

// --- share_deactivate ---

func TestShareDeactivate(t *testing.T) {
	h, _, _, _, _, sh := newToolHandlers()
	ctx := ctxWithUser(42)

	sh.On("Deactivate", ctx, uint(1), uint(42)).Return(nil)

	res, _, err := h.shareDeactivate(ctx, nil, ShareDeactivateInput{PromptID: 1})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	assert.Equal(t, "deactivated", data["status"])
}

func TestShareDeactivate_NotFound(t *testing.T) {
	h, _, _, _, _, sh := newToolHandlers()
	ctx := ctxWithUser(42)

	sh.On("Deactivate", ctx, uint(1), uint(42)).Return(shareuc.ErrNotFound)

	_, _, err := h.shareDeactivate(ctx, nil, ShareDeactivateInput{PromptID: 1})
	assert.EqualError(t, err, "share link not found")
}

// --- collection_get ---

func TestCollectionGet(t *testing.T) {
	h, _, c, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	coll := &models.Collection{ID: 1, Name: "Dev", Color: "#8b5cf6", Description: "Dev stuff"}
	c.On("GetByID", ctx, uint(1), uint(42)).Return(coll, nil)

	res, _, err := h.collectionGet(ctx, nil, CollectionGetInput{ID: 1})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	assert.Equal(t, "Dev", data["name"])
}

func TestCollectionGet_NotFound(t *testing.T) {
	h, _, c, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	c.On("GetByID", ctx, uint(99), uint(42)).Return(nil, colluc.ErrNotFound)

	_, _, err := h.collectionGet(ctx, nil, CollectionGetInput{ID: 99})
	assert.EqualError(t, err, "collection not found")
}

// --- collection_update ---

func TestCollectionUpdate(t *testing.T) {
	h, _, c, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	current := &models.Collection{ID: 1, Name: "Old", Description: "desc", Color: "#000000", Icon: "📁"}
	c.On("GetByID", ctx, uint(1), uint(42)).Return(current, nil)

	newName := "New"
	updated := &models.Collection{ID: 1, Name: "New", Description: "desc", Color: "#000000", Icon: "📁"}
	c.On("Update", ctx, uint(1), uint(42), "New", "desc", "#000000", "📁").Return(updated, nil)

	res, _, err := h.collectionUpdate(ctx, nil, CollectionUpdateInput{ID: 1, Name: &newName})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	assert.Equal(t, "New", data["name"])
}

func TestCollectionUpdate_InvalidColor(t *testing.T) {
	h, _, c, _, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	current := &models.Collection{ID: 1, Name: "Old", Description: "", Color: "#000000", Icon: ""}
	c.On("GetByID", ctx, uint(1), uint(42)).Return(current, nil)

	badColor := "red"
	c.On("Update", ctx, uint(1), uint(42), "Old", "", "red", "").Return(nil, colluc.ErrInvalidColor)

	_, _, err := h.collectionUpdate(ctx, nil, CollectionUpdateInput{ID: 1, Color: &badColor})
	assert.EqualError(t, err, "invalid color: use #RRGGBB format")
}

// --- tag_delete ---

func TestTagDelete(t *testing.T) {
	h, _, _, tg, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	tg.On("Delete", ctx, uint(1), uint(42)).Return(nil)

	res, _, err := h.tagDelete(ctx, nil, TagDeleteInput{ID: 1})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	assert.Equal(t, "deleted", data["status"])
}

func TestTagDelete_NotFound(t *testing.T) {
	h, _, _, tg, _, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	tg.On("Delete", ctx, uint(99), uint(42)).Return(taguc.ErrNotFound)

	_, _, err := h.tagDelete(ctx, nil, TagDeleteInput{ID: 99})
	assert.EqualError(t, err, "tag not found")
}

// --- search_suggest ---

func TestSearchSuggest(t *testing.T) {
	h, _, _, _, s, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	output := &searchuc.SuggestOutput{
		Suggestions: []searchuc.Suggestion{
			{Text: "hello world", Type: "prompt"},
			{Text: "http client", Type: "prompt"},
		},
	}
	s.On("Suggest", ctx, uint(42), (*uint)(nil), "hel").Return(output, nil)

	res, _, err := h.searchSuggest(ctx, nil, SuggestInput{Prefix: "hel"})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	suggestions := data["suggestions"].([]any)
	assert.Len(t, suggestions, 2)
}

func TestSearchSuggest_Empty(t *testing.T) {
	h, _, _, _, s, _ := newToolHandlers()
	ctx := ctxWithUser(42)

	output := &searchuc.SuggestOutput{Suggestions: []searchuc.Suggestion{}}
	s.On("Suggest", ctx, uint(42), (*uint)(nil), "zzz").Return(output, nil)

	res, _, err := h.searchSuggest(ctx, nil, SuggestInput{Prefix: "zzz"})
	require.NoError(t, err)

	data := parseJSON(t, res.Content[0].(*sdkmcp.TextContent).Text)
	suggestions := data["suggestions"].([]any)
	assert.Len(t, suggestions, 0)
}
