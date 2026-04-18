package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	"promptvault/internal/template"
	apikeyuc "promptvault/internal/usecases/apikey"
	colluc "promptvault/internal/usecases/collection"
	promptuc "promptvault/internal/usecases/prompt"
	quotauc "promptvault/internal/usecases/quota"
	shareuc "promptvault/internal/usecases/share"
	taguc "promptvault/internal/usecases/tag"
)

type toolHandlers struct {
	prompts     PromptService
	collections CollectionService
	tags        TagService
	search      SearchService
	shares      ShareService
	quotas      *quotauc.Service
	cache       *listCache // P-11: TTL cache для list_collections/list_tags
	notifier    *notifier  // C-1: рассылка resources/updated подписчикам
}

// checkAndIncrementMCP проверяет MCP-квоту перед write-операцией и инкрементит после.
func (h *toolHandlers) checkMCPQuota(ctx context.Context) error {
	if h.quotas == nil {
		return nil
	}
	return h.quotas.CheckMCPQuota(ctx, authmw.GetUserID(ctx))
}

func (h *toolHandlers) incrementMCPUsage(ctx context.Context) {
	if h.quotas == nil {
		return
	}
	userID := authmw.GetUserID(ctx)
	if err := h.quotas.IncrementMCPUsage(ctx, userID); err != nil {
		// Quota drift = revenue leak (write прошёл, счётчик не увеличился).
		// Sentry ловит — ops должен реагировать.
		slog.Error("mcp.quota.increment_failed", "user_id", userID, "error", err)
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.CaptureException(err)
		}
	}
}

// invalidateCollectionsCache / invalidateTagsCache — вызываются после успешных CUD
// через MCP, чтобы юзер не видел stale при собственных изменениях (P-11).
// При CUD через HTTP API у MCP-клиента всё равно будет stale до истечения TTL —
// это приемлемо для 30с окна.
func (h *toolHandlers) invalidateCollectionsCache(ctx context.Context) {
	if h.cache == nil {
		return
	}
	h.cache.InvalidateUser(authmw.GetUserID(ctx), "collections")
}

func (h *toolHandlers) invalidateTagsCache(ctx context.Context) {
	if h.cache == nil {
		return
	}
	h.cache.InvalidateUser(authmw.GetUserID(ctx), "tags")
}

// --- tool input types ---

type SearchInput struct {
	Query  string `json:"query" jsonschema:"Search query"`
	TeamID *uint  `json:"team_id,omitempty" jsonschema:"Team ID (omit for personal workspace)"`
}

type ListPromptsInput struct {
	TeamID       *uint  `json:"team_id,omitempty" jsonschema:"Team ID (omit for personal workspace)"`
	CollectionID *uint  `json:"collection_id,omitempty" jsonschema:"Filter by collection"`
	TagIDs       []uint `json:"tag_ids,omitempty" jsonschema:"Filter by tag IDs"`
	FavoriteOnly bool   `json:"favorite_only,omitempty" jsonschema:"Show only favorites"`
	Query        string `json:"query,omitempty" jsonschema:"Search within prompts"`
	Page         int    `json:"page,omitempty" jsonschema:"Page number (0-based). Ignored if cursor is provided."`
	PageSize     int    `json:"page_size,omitempty" jsonschema:"Items per page (default 20). Legacy; use 'limit' instead."`
	Cursor       string `json:"cursor,omitempty" jsonschema:"Opaque pagination cursor from previous response's next_cursor. Omit for first page."`
	Limit        int    `json:"limit,omitempty" jsonschema:"Max items per page when using cursor (default 50, max 200)."`
}

type GetPromptInput struct {
	ID uint `json:"id" jsonschema:"required,Prompt ID"`
}

type CreatePromptInput struct {
	Title         string `json:"title" jsonschema:"required,Prompt title"`
	Content       string `json:"content" jsonschema:"required,Prompt content"`
	TeamID        *uint  `json:"team_id,omitempty" jsonschema:"Team ID (omit for personal workspace)"`
	Model         string `json:"model,omitempty" jsonschema:"AI model name"`
	CollectionIDs []uint `json:"collection_ids,omitempty" jsonschema:"Collection IDs to assign"`
	TagIDs        []uint `json:"tag_ids,omitempty" jsonschema:"Tag IDs to assign"`
}

type UpdatePromptInput struct {
	ID            uint    `json:"id" jsonschema:"required,Prompt ID"`
	Title         *string `json:"title,omitempty" jsonschema:"New title"`
	Content       *string `json:"content,omitempty" jsonschema:"New content"`
	Model         *string `json:"model,omitempty" jsonschema:"New model"`
	ChangeNote    string  `json:"change_note,omitempty" jsonschema:"Description of changes"`
	CollectionIDs []uint  `json:"collection_ids,omitempty" jsonschema:"New collection IDs"`
	TagIDs        []uint  `json:"tag_ids,omitempty" jsonschema:"New tag IDs"`
}

type DeletePromptInput struct {
	ID uint `json:"id" jsonschema:"required,Prompt ID"`
}

type ListCollectionsInput struct {
	TeamID *uint `json:"team_id,omitempty" jsonschema:"Team ID (omit for personal workspace)"`
}

type ListTagsInput struct {
	TeamID *uint `json:"team_id,omitempty" jsonschema:"Team ID (omit for personal workspace)"`
}

type CreateTagInput struct {
	Name   string `json:"name" jsonschema:"required,Tag name"`
	Color  string `json:"color,omitempty" jsonschema:"Tag color (#RRGGBB)"`
	TeamID *uint  `json:"team_id,omitempty" jsonschema:"Team ID (omit for personal workspace)"`
}

type GetVersionsInput struct {
	PromptID uint `json:"prompt_id" jsonschema:"required,Prompt ID"`
	Page     int  `json:"page,omitempty" jsonschema:"Page number (0-based)"`
	PageSize int  `json:"page_size,omitempty" jsonschema:"Items per page (default 20)"`
}

type CreateCollectionInput struct {
	Name        string `json:"name" jsonschema:"required,Collection name"`
	Description string `json:"description,omitempty" jsonschema:"Collection description"`
	Color       string `json:"color,omitempty" jsonschema:"Color in #RRGGBB format (default #8b5cf6)"`
	Icon        string `json:"icon,omitempty" jsonschema:"Icon emoji"`
	TeamID      *uint  `json:"team_id,omitempty" jsonschema:"Team ID (omit for personal workspace)"`
}

type DeleteCollectionInput struct {
	ID uint `json:"id" jsonschema:"required,Collection ID"`
}

// --- new tool input types ---

type PromptIDInput struct {
	ID uint `json:"id" jsonschema:"required,Prompt ID"`
}

type PromptPinInput struct {
	ID       uint `json:"id" jsonschema:"required,Prompt ID"`
	TeamWide bool `json:"team_wide,omitempty" jsonschema:"Pin for entire team (owner/editor only)"`
}

type ListLimitedInput struct {
	TeamID *uint `json:"team_id,omitempty" jsonschema:"Team ID (omit for personal workspace)"`
	Limit  int   `json:"limit,omitempty" jsonschema:"Max items to return (default 10, max 50)"`
}

type RevertInput struct {
	PromptID  uint `json:"prompt_id" jsonschema:"required,Prompt ID"`
	VersionID uint `json:"version_id" jsonschema:"required,Version ID to revert to"`
}

type ShareCreateInput struct {
	PromptID uint `json:"prompt_id" jsonschema:"required,Prompt ID"`
}

type ShareDeactivateInput struct {
	PromptID uint `json:"prompt_id" jsonschema:"required,Prompt ID"`
}

type CollectionGetInput struct {
	ID uint `json:"id" jsonschema:"required,Collection ID"`
}

type CollectionUpdateInput struct {
	ID          uint    `json:"id" jsonschema:"required,Collection ID"`
	Name        *string `json:"name,omitempty" jsonschema:"New name"`
	Description *string `json:"description,omitempty" jsonschema:"New description"`
	Color       *string `json:"color,omitempty" jsonschema:"New color (#RRGGBB)"`
	Icon        *string `json:"icon,omitempty" jsonschema:"New icon emoji"`
}

type TagDeleteInput struct {
	ID uint `json:"id" jsonschema:"required,Tag ID"`
}

type SuggestInput struct {
	Prefix string `json:"prefix" jsonschema:"required,Search prefix for autocomplete"`
	TeamID *uint  `json:"team_id,omitempty" jsonschema:"Team ID (omit for personal workspace)"`
}

var (
	readOnlyAnnotations = &sdkmcp.ToolAnnotations{
		ReadOnlyHint:  true,
		OpenWorldHint: boolPtr(false),
	}
	writeAnnotations = &sdkmcp.ToolAnnotations{
		DestructiveHint: boolPtr(false),
		OpenWorldHint:   boolPtr(false),
	}
	deleteAnnotations = &sdkmcp.ToolAnnotations{
		// DestructiveHint defaults to true — correct for delete operations.
		OpenWorldHint: boolPtr(false),
	}
	// writeIdempotentAnnotations — for toggle/upsert operations safe to repeat.
	writeIdempotentAnnotations = &sdkmcp.ToolAnnotations{
		DestructiveHint: boolPtr(false),
		IdempotentHint:  true,
		OpenWorldHint:   boolPtr(false),
	}
	// deleteIdempotentAnnotations — for destructive but idempotent operations (deactivate, delete).
	deleteIdempotentAnnotations = &sdkmcp.ToolAnnotations{
		IdempotentHint: true,
		OpenWorldHint:  boolPtr(false),
	}
)

func boolPtr(b bool) *bool { return &b }

func (t *toolHandlers) register(server *sdkmcp.Server) {
	// --- read tools ---
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "search_prompts",
		Title:       "Search prompts",
		Description: "Search prompts, collections, and tags by query",
		Annotations: readOnlyAnnotations,
	}, t.searchPrompts)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "list_prompts",
		Title:       "List prompts",
		Description: "List prompts with optional filters (collection, tags, favorites, search)",
		Annotations: readOnlyAnnotations,
	}, t.listPrompts)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "get_prompt",
		Title:       "Get prompt",
		Description: "Get a single prompt by ID with full content",
		Annotations: readOnlyAnnotations,
	}, t.getPrompt)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "list_collections",
		Title:       "List collections",
		Description: "List all collections",
		Annotations: readOnlyAnnotations,
	}, t.listCollections)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "list_tags",
		Title:       "List tags",
		Description: "List all tags",
		Annotations: readOnlyAnnotations,
	}, t.listTags)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "get_prompt_versions",
		Title:       "Get prompt versions",
		Description: "Get version history for a prompt",
		Annotations: readOnlyAnnotations,
	}, t.getPromptVersions)

	// --- write tools ---
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "create_prompt",
		Title:       "Create prompt",
		Description: "Create a new prompt",
		Annotations: writeAnnotations,
	}, t.createPrompt)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "update_prompt",
		Title:       "Update prompt",
		Description: "Update an existing prompt (creates a new version)",
		Annotations: writeAnnotations,
	}, t.updatePrompt)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "delete_prompt",
		Title:       "Delete prompt",
		Description: "Move a prompt to trash (recoverable for 30 days)",
		Annotations: deleteAnnotations,
	}, t.deletePrompt)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "create_tag",
		Title:       "Create tag",
		Description: "Create a new tag",
		Annotations: writeAnnotations,
	}, t.createTag)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "create_collection",
		Title:       "Create collection",
		Description: "Create a new collection for organizing prompts",
		Annotations: writeAnnotations,
	}, t.createCollection)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "delete_collection",
		Title:       "Delete collection",
		Description: "Delete a collection (prompts inside are not deleted)",
		Annotations: deleteAnnotations,
	}, t.deleteCollection)

	// --- new read tools ---
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "prompt_list_pinned",
		Title:       "List pinned prompts",
		Description: "List pinned prompts. Use to quickly access frequently used prompts.",
		Annotations: readOnlyAnnotations,
	}, t.promptListPinned)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "prompt_list_recent",
		Title:       "List recent prompts",
		Description: "List recently used prompts, ordered by last access time.",
		Annotations: readOnlyAnnotations,
	}, t.promptListRecent)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "collection_get",
		Title:       "Get collection",
		Description: "Get a single collection by ID.",
		Annotations: readOnlyAnnotations,
	}, t.collectionGet)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "search_suggest",
		Title:       "Search suggestions",
		Description: "Get autocomplete suggestions by prefix. Use for interactive search.",
		Annotations: readOnlyAnnotations,
	}, t.searchSuggest)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "list_prompt_vars",
		Title:       "List prompt variables",
		Description: "Extract {{variables}} from a prompt's content. Use before use_prompt to know what values to pass in the vars argument.",
		Annotations: readOnlyAnnotations,
	}, t.listPromptVars)

	// --- new write tools ---
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "prompt_favorite",
		Title:       "Toggle favorite",
		Description: "Toggle favorite status on a prompt. Safe to call repeatedly.",
		Annotations: writeIdempotentAnnotations,
	}, t.promptFavorite)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "prompt_pin",
		Title:       "Toggle pin",
		Description: "Pin or unpin a prompt. Use team_wide=true to pin for entire team (owner/editor only).",
		Annotations: writeIdempotentAnnotations,
	}, t.promptPin)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "prompt_increment_usage",
		Title:       "Record prompt usage",
		Description: "Record that a prompt was used. Increments usage counter for analytics.",
		Annotations: writeAnnotations,
	}, t.promptIncrementUsage)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "share_create",
		Title:       "Create share link",
		Description: "Create a public share link for a prompt. Returns existing link if one is already active.",
		Annotations: writeIdempotentAnnotations,
	}, t.shareCreate)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "collection_update",
		Title:       "Update collection",
		Description: "Update collection name, description, color, or icon.",
		Annotations: writeIdempotentAnnotations,
	}, t.collectionUpdate)

	// --- new destructive/revert tools ---
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "prompt_revert",
		Title:       "Revert prompt to version",
		Description: "Revert a prompt to a previous version. Creates a new version with the old content. Use get_prompt_versions first to find the version ID.",
		Annotations: writeAnnotations,
	}, t.promptRevert)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "share_deactivate",
		Title:       "Deactivate share link",
		Description: "Deactivate the public share link for a prompt. The link will no longer be accessible.",
		Annotations: deleteIdempotentAnnotations,
	}, t.shareDeactivate)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "tag_delete",
		Title:       "Delete tag",
		Description: "Delete a tag. Prompts using this tag are not affected.",
		Annotations: deleteIdempotentAnnotations,
	}, t.tagDelete)
}

// --- logging wrapper ---

func logTool(ctx context.Context, name string, start time.Time, err error) {
	dur := time.Since(start).Milliseconds()
	userID := authmw.GetUserID(ctx)
	if err != nil {
		if isDomainError(err) {
			slog.Warn("mcp.tool.error", "tool", name, "user_id", userID, "duration_ms", dur, "error", err)
		} else {
			slog.Error("mcp.tool.error", "tool", name, "user_id", userID, "duration_ms", dur, "error", err)
			if hub := sentry.GetHubFromContext(ctx); hub != nil {
				hub.CaptureException(err)
			}
		}
	} else {
		slog.Info("mcp.tool.called", "tool", name, "user_id", userID, "duration_ms", dur)
	}
}

// isDomainError returns true for expected user/business errors (4xx-equivalent).
func isDomainError(err error) bool {
	domainErrors := []error{
		promptuc.ErrNotFound, promptuc.ErrForbidden, promptuc.ErrViewerReadOnly,
		promptuc.ErrVersionNotFound, promptuc.ErrWorkspaceMismatch, promptuc.ErrPinForbidden,
		colluc.ErrNotFound, colluc.ErrForbidden, colluc.ErrViewerReadOnly,
		taguc.ErrNotFound, taguc.ErrForbidden, taguc.ErrViewerReadOnly, taguc.ErrNameEmpty,
		shareuc.ErrNotFound, shareuc.ErrPromptNotFound, shareuc.ErrForbidden, shareuc.ErrViewerReadOnly,
		apikeyuc.ErrScopeDenied, apikeyuc.ErrTeamMismatch,
	}
	for _, de := range domainErrors {
		if errors.Is(err, de) {
			return true
		}
	}
	return false
}

// --- converters ---

func toPromptResponse(p *models.Prompt) PromptResponse {
	tags := make([]TagResponse, len(p.Tags))
	for i, t := range p.Tags {
		tags[i] = TagResponse{ID: t.ID, Name: t.Name, Color: t.Color}
	}
	colls := make([]CollectionResponse, len(p.Collections))
	for i, c := range p.Collections {
		colls[i] = CollectionResponse{ID: c.ID, Name: c.Name, Description: c.Description, Color: c.Color, Icon: c.Icon}
	}
	return PromptResponse{
		ID: p.ID, Title: p.Title, Content: p.Content, Model: p.Model,
		Favorite: p.Favorite, UsageCount: p.UsageCount,
		Tags: tags, Collections: colls,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toPromptList(prompts []models.Prompt) []PromptResponse {
	result := make([]PromptResponse, len(prompts))
	for i := range prompts {
		result[i] = toPromptResponse(&prompts[i])
	}
	return result
}

func jsonResult(data any) (*sdkmcp.CallToolResult, error) {
	b, err := json.Marshal(data)
	if err != nil {
		slog.Error("mcp.json_marshal_failed", "error", err)
		return nil, fmt.Errorf("failed to serialize response: %w", err)
	}
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: string(b)}},
	}, nil
}

// --- read tool handlers ---

func (t *toolHandlers) searchPrompts(ctx context.Context, _ *sdkmcp.CallToolRequest, input SearchInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "search_prompts", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	result, err := t.search.Search(ctx, userID, input.TeamID, input.Query)
	logTool(ctx, "search_prompts", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	res, err := jsonResult(result)
	return res, nil, err
}

func (t *toolHandlers) listPrompts(ctx context.Context, _ *sdkmcp.CallToolRequest, input ListPromptsInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "list_prompts", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	// pageSize: новый "limit" (cursor mode) имеет приоритет — до 200,
	// иначе legacy page_size — до 100. Default 20 (исторический).
	var pageSize int
	switch {
	case input.Limit > 0:
		pageSize = input.Limit
		if pageSize > 200 {
			pageSize = 200
		}
	case input.PageSize > 0:
		pageSize = input.PageSize
		if pageSize > 100 {
			pageSize = 100
		}
	default:
		pageSize = 20
	}

	filter := repo.PromptListFilter{
		UserID:       userID,
		CollectionID: input.CollectionID,
		TagIDs:       input.TagIDs,
		FavoriteOnly: input.FavoriteOnly,
		Query:        input.Query,
		Page:         input.Page,
		PageSize:     pageSize,
	}
	if input.TeamID != nil {
		filter.TeamIDs = []uint{*input.TeamID}
	}

	// Keyset cursor (C-3): декодируем, сверяем filter_hash.
	usingCursor := input.Cursor != ""
	if usingCursor {
		c, err := decodeCursor(input.Cursor)
		if err != nil {
			logTool(ctx, "list_prompts", start, err)
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}
		if c.Fh != filterHash(filter) {
			logTool(ctx, "list_prompts", start, ErrCursorFilterMismatch)
			return nil, nil, ErrCursorFilterMismatch
		}
		lid := c.Lid
		lts := c.Lts
		filter.AfterID = &lid
		filter.AfterUpdatedAt = &lts
	}

	prompts, total, err := t.prompts.List(ctx, filter)
	logTool(ctx, "list_prompts", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}

	slog.Info("mcp.pagination.cursor",
		"tool", "list_prompts",
		"user_id", userID,
		"limit", pageSize,
		"page_size_returned", len(prompts),
		"used_cursor", usingCursor,
	)

	payload := map[string]any{
		"prompts": toPromptList(prompts),
		"total":   total,
	}
	if len(prompts) == pageSize {
		last := prompts[len(prompts)-1]
		next, err := encodeCursor(cursorData{
			Lid: last.ID,
			Lts: last.UpdatedAt,
			Fh:  filterHash(filter),
		})
		if err == nil {
			payload["next_cursor"] = next
		}
	}
	res, err := jsonResult(payload)
	return res, nil, err
}

func (t *toolHandlers) getPrompt(ctx context.Context, _ *sdkmcp.CallToolRequest, input GetPromptInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "get_prompt", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	prompt, err := t.prompts.GetByID(ctx, input.ID, userID)
	logTool(ctx, "get_prompt", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	res, err := jsonResult(toPromptResponse(prompt))
	return res, nil, err
}

func (t *toolHandlers) listCollections(ctx context.Context, _ *sdkmcp.CallToolRequest, input ListCollectionsInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "list_collections", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	teamKey := "nil"
	if input.TeamID != nil {
		teamKey = uintToStr(*input.TeamID)
	}
	cacheKey := "collections:" + uintToStr(userID) + ":" + teamKey

	if t.cache != nil {
		if cached, ok := t.cache.Get(cacheKey); ok {
			logTool(ctx, "list_collections.cache_hit", start, nil)
			res, err := jsonResult(cached)
			return res, nil, err
		}
	}

	var teamIDs []uint
	if input.TeamID != nil {
		teamIDs = []uint{*input.TeamID}
	}

	colls, err := t.collections.List(ctx, userID, teamIDs)
	logTool(ctx, "list_collections", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}

	result := make([]CollectionWithCountResponse, len(colls))
	for i, c := range colls {
		result[i] = CollectionWithCountResponse{
			CollectionResponse: CollectionResponse{
				ID: c.ID, Name: c.Name, Description: c.Description, Color: c.Color, Icon: c.Icon,
			},
			PromptCount: c.PromptCount,
		}
	}
	payload := map[string]any{"collections": result}
	if t.cache != nil {
		t.cache.Set(cacheKey, payload)
	}
	res, err := jsonResult(payload)
	return res, nil, err
}

func (t *toolHandlers) listTags(ctx context.Context, _ *sdkmcp.CallToolRequest, input ListTagsInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "list_tags", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	teamKey := "nil"
	if input.TeamID != nil {
		teamKey = uintToStr(*input.TeamID)
	}
	cacheKey := "tags:" + uintToStr(userID) + ":" + teamKey

	if t.cache != nil {
		if cached, ok := t.cache.Get(cacheKey); ok {
			logTool(ctx, "list_tags.cache_hit", start, nil)
			res, err := jsonResult(cached)
			return res, nil, err
		}
	}

	tags, err := t.tags.List(ctx, userID, input.TeamID)
	logTool(ctx, "list_tags", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}

	result := make([]TagResponse, len(tags))
	for i, tag := range tags {
		result[i] = TagResponse{ID: tag.ID, Name: tag.Name, Color: tag.Color}
	}
	payload := map[string]any{"tags": result}
	if t.cache != nil {
		t.cache.Set(cacheKey, payload)
	}
	res, err := jsonResult(payload)
	return res, nil, err
}

func (t *toolHandlers) getPromptVersions(ctx context.Context, _ *sdkmcp.CallToolRequest, input GetVersionsInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "get_prompt_versions", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	versions, total, err := t.prompts.ListVersions(ctx, input.PromptID, userID, input.Page, pageSize)
	logTool(ctx, "get_prompt_versions", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}

	result := make([]VersionResponse, len(versions))
	for i, v := range versions {
		result[i] = VersionResponse{
			ID: v.ID, VersionNumber: v.VersionNumber, Title: v.Title,
			Content: v.Content, Model: v.Model, ChangeNote: v.ChangeNote, CreatedAt: v.CreatedAt,
		}
	}
	res, err := jsonResult(map[string]any{"versions": result, "total": total})
	return res, nil, err
}

// --- write tool handlers ---

func (t *toolHandlers) createPrompt(ctx context.Context, _ *sdkmcp.CallToolRequest, input CreatePromptInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "create_prompt", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := t.checkMCPQuota(ctx); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	prompt, err := t.prompts.Create(ctx, promptuc.CreateInput{
		UserID:        userID,
		TeamID:        input.TeamID,
		Title:         input.Title,
		Content:       input.Content,
		Model:         input.Model,
		CollectionIDs: input.CollectionIDs,
		TagIDs:        input.TagIDs,
	})
	logTool(ctx, "create_prompt", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.incrementMCPUsage(ctx)
	res, err := jsonResult(toPromptResponse(prompt))
	return res, nil, err
}

func (t *toolHandlers) updatePrompt(ctx context.Context, _ *sdkmcp.CallToolRequest, input UpdatePromptInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "update_prompt", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := t.checkMCPQuota(ctx); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	prompt, err := t.prompts.Update(ctx, input.ID, userID, promptuc.UpdateInput{
		Title:         input.Title,
		Content:       input.Content,
		Model:         input.Model,
		ChangeNote:    input.ChangeNote,
		CollectionIDs: input.CollectionIDs,
		TagIDs:        input.TagIDs,
	})
	logTool(ctx, "update_prompt", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.incrementMCPUsage(ctx)
	t.notifier.NotifyPrompt(ctx, input.ID)
	res, err := jsonResult(toPromptResponse(prompt))
	return res, nil, err
}

func (t *toolHandlers) deletePrompt(ctx context.Context, _ *sdkmcp.CallToolRequest, input DeletePromptInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "delete_prompt", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := t.checkMCPQuota(ctx); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	err := t.prompts.Delete(ctx, input.ID, userID)
	logTool(ctx, "delete_prompt", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.notifier.NotifyPrompt(ctx, input.ID)
	res, err := jsonResult(map[string]string{
		"status":  "deleted",
		"message": "Prompt moved to trash. Recoverable for 30 days.",
	})
	return res, nil, err
}

func (t *toolHandlers) createTag(ctx context.Context, _ *sdkmcp.CallToolRequest, input CreateTagInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "create_tag", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := t.checkMCPQuota(ctx); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	tag, err := t.tags.Create(ctx, input.Name, input.Color, userID, input.TeamID)
	logTool(ctx, "create_tag", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.invalidateTagsCache(ctx)
	t.notifier.NotifyTags(ctx)
	res, err := jsonResult(TagResponse{ID: tag.ID, Name: tag.Name, Color: tag.Color})
	return res, nil, err
}

func (t *toolHandlers) createCollection(ctx context.Context, _ *sdkmcp.CallToolRequest, input CreateCollectionInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "create_collection", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := t.checkMCPQuota(ctx); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	coll, err := t.collections.Create(ctx, userID, input.Name, input.Description, input.Color, input.Icon, input.TeamID)
	logTool(ctx, "create_collection", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.invalidateCollectionsCache(ctx)
	t.notifier.NotifyCollections(ctx)
	res, err := jsonResult(CollectionResponse{
		ID: coll.ID, Name: coll.Name, Description: coll.Description, Color: coll.Color, Icon: coll.Icon,
	})
	return res, nil, err
}

func (t *toolHandlers) deleteCollection(ctx context.Context, _ *sdkmcp.CallToolRequest, input DeleteCollectionInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "delete_collection", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := t.checkMCPQuota(ctx); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	err := t.collections.Delete(ctx, input.ID, userID)
	logTool(ctx, "delete_collection", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.invalidateCollectionsCache(ctx)
	t.notifier.NotifyCollections(ctx)
	res, err := jsonResult(map[string]string{
		"status":  "deleted",
		"message": "Collection deleted. Prompts inside are not affected.",
	})
	return res, nil, err
}

// --- new read tool handlers ---

func (t *toolHandlers) promptListPinned(ctx context.Context, _ *sdkmcp.CallToolRequest, input ListLimitedInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "prompt_list_pinned", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	prompts, err := t.prompts.ListPinned(ctx, userID, input.TeamID, limit)
	logTool(ctx, "prompt_list_pinned", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	res, err := jsonResult(map[string]any{"prompts": toPromptList(prompts)})
	return res, nil, err
}

func (t *toolHandlers) promptListRecent(ctx context.Context, _ *sdkmcp.CallToolRequest, input ListLimitedInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "prompt_list_recent", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	prompts, err := t.prompts.ListRecent(ctx, userID, input.TeamID, limit)
	logTool(ctx, "prompt_list_recent", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	res, err := jsonResult(map[string]any{"prompts": toPromptList(prompts)})
	return res, nil, err
}

func (t *toolHandlers) collectionGet(ctx context.Context, _ *sdkmcp.CallToolRequest, input CollectionGetInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "collection_get", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	coll, err := t.collections.GetByID(ctx, input.ID, userID)
	logTool(ctx, "collection_get", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	res, err := jsonResult(CollectionResponse{
		ID: coll.ID, Name: coll.Name, Description: coll.Description, Color: coll.Color, Icon: coll.Icon,
	})
	return res, nil, err
}

func (t *toolHandlers) listPromptVars(ctx context.Context, _ *sdkmcp.CallToolRequest, input PromptIDInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "list_prompt_vars", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	prompt, err := t.prompts.GetByID(ctx, input.ID, userID)
	logTool(ctx, "list_prompt_vars", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	vars := template.Extract(prompt.Content)
	if vars == nil {
		vars = []string{}
	}
	res, err := jsonResult(map[string]any{"variables": vars})
	return res, nil, err
}

func (t *toolHandlers) searchSuggest(ctx context.Context, _ *sdkmcp.CallToolRequest, input SuggestInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "search_suggest", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	result, err := t.search.Suggest(ctx, userID, input.TeamID, input.Prefix)
	logTool(ctx, "search_suggest", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	res, err := jsonResult(result)
	return res, nil, err
}

// --- new write tool handlers ---

func (t *toolHandlers) promptFavorite(ctx context.Context, _ *sdkmcp.CallToolRequest, input PromptIDInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "prompt_favorite", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	prompt, err := t.prompts.ToggleFavorite(ctx, input.ID, userID)
	logTool(ctx, "prompt_favorite", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.notifier.NotifyPrompt(ctx, input.ID)
	res, err := jsonResult(toPromptResponse(prompt))
	return res, nil, err
}

func (t *toolHandlers) promptPin(ctx context.Context, _ *sdkmcp.CallToolRequest, input PromptPinInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "prompt_pin", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	result, err := t.prompts.TogglePin(ctx, promptuc.PinInput{
		PromptID: input.ID,
		UserID:   userID,
		TeamWide: input.TeamWide,
	})
	logTool(ctx, "prompt_pin", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.notifier.NotifyPrompt(ctx, input.ID)
	res, err := jsonResult(PinResultResponse{Pinned: result.Pinned, TeamWide: result.TeamWide})
	return res, nil, err
}

func (t *toolHandlers) promptIncrementUsage(ctx context.Context, _ *sdkmcp.CallToolRequest, input PromptIDInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "prompt_increment_usage", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	err := t.prompts.IncrementUsage(ctx, input.ID, userID)
	logTool(ctx, "prompt_increment_usage", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	res, err := jsonResult(map[string]string{"status": "recorded"})
	return res, nil, err
}

func (t *toolHandlers) shareCreate(ctx context.Context, _ *sdkmcp.CallToolRequest, input ShareCreateInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "share_create", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	link, _, err := t.shares.CreateOrGet(ctx, input.PromptID, userID)
	logTool(ctx, "share_create", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.notifier.NotifyPrompt(ctx, input.PromptID)
	res, err := jsonResult(ShareLinkResponse{
		ID: link.ID, Token: link.Token, URL: link.URL,
		IsActive: link.IsActive, ViewCount: link.ViewCount,
		LastViewedAt: link.LastViewedAt, CreatedAt: link.CreatedAt,
	})
	return res, nil, err
}

func (t *toolHandlers) collectionUpdate(ctx context.Context, _ *sdkmcp.CallToolRequest, input CollectionUpdateInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "collection_update", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	// Resolve optional fields: fetch current values for fields not provided.
	current, err := t.collections.GetByID(ctx, input.ID, userID)
	if err != nil {
		logTool(ctx, "collection_update", start, err)
		return nil, nil, mapDomainError(err)
	}

	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}
	description := current.Description
	if input.Description != nil {
		description = *input.Description
	}
	color := current.Color
	if input.Color != nil {
		color = *input.Color
	}
	icon := current.Icon
	if input.Icon != nil {
		icon = *input.Icon
	}

	coll, err := t.collections.Update(ctx, input.ID, userID, name, description, color, icon)
	logTool(ctx, "collection_update", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.invalidateCollectionsCache(ctx)
	t.notifier.NotifyCollections(ctx)
	res, err := jsonResult(CollectionResponse{
		ID: coll.ID, Name: coll.Name, Description: coll.Description, Color: coll.Color, Icon: coll.Icon,
	})
	return res, nil, err
}

// --- new destructive/revert tool handlers ---

func (t *toolHandlers) promptRevert(ctx context.Context, _ *sdkmcp.CallToolRequest, input RevertInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "prompt_revert", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	prompt, err := t.prompts.RevertToVersion(ctx, input.PromptID, userID, input.VersionID)
	logTool(ctx, "prompt_revert", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.notifier.NotifyPrompt(ctx, input.PromptID)
	res, err := jsonResult(toPromptResponse(prompt))
	return res, nil, err
}

func (t *toolHandlers) shareDeactivate(ctx context.Context, _ *sdkmcp.CallToolRequest, input ShareDeactivateInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "share_deactivate", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	err := t.shares.Deactivate(ctx, input.PromptID, userID)
	logTool(ctx, "share_deactivate", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.notifier.NotifyPrompt(ctx, input.PromptID)
	res, err := jsonResult(map[string]string{"status": "deactivated"})
	return res, nil, err
}

func (t *toolHandlers) tagDelete(ctx context.Context, _ *sdkmcp.CallToolRequest, input TagDeleteInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "tag_delete", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	start := time.Now()
	userID := authmw.GetUserID(ctx)

	err := t.tags.Delete(ctx, input.ID, userID)
	logTool(ctx, "tag_delete", start, err)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.invalidateTagsCache(ctx)
	t.notifier.NotifyTags(ctx)
	res, err := jsonResult(map[string]string{
		"status":  "deleted",
		"message": "Tag deleted. Prompts are not affected.",
	})
	return res, nil, err
}
