package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	authmw "promptvault/internal/middleware/auth"
	promptuc "promptvault/internal/usecases/prompt"
)

type resourceHandlers struct {
	prompts     PromptService
	collections CollectionService
	tags        TagService
}

func (r *resourceHandlers) register(server *sdkmcp.Server) {
	// Static resources — lists that LLM can read for context
	server.AddResource(&sdkmcp.Resource{
		URI:         "promptvault://collections",
		Name:        "collections",
		Description: "List of all collections in the current workspace",
		MIMEType:    "application/json",
	}, r.readCollections)

	server.AddResource(&sdkmcp.Resource{
		URI:         "promptvault://tags",
		Name:        "tags",
		Description: "List of all tags in the current workspace",
		MIMEType:    "application/json",
	}, r.readTags)

	// Resource template — read a specific prompt by ID
	server.AddResourceTemplate(&sdkmcp.ResourceTemplate{
		URITemplate: "promptvault://prompts/{id}",
		Name:        "prompt",
		Description: "Read a specific prompt by ID",
		MIMEType:    "application/json",
	}, r.readPrompt)

	// MCP Prompt — use_prompt: fetch a prompt and format it for LLM use
	server.AddPrompt(&sdkmcp.Prompt{
		Name:        "use_prompt",
		Description: "Fetch a prompt from your library and format it for use",
		Arguments: []*sdkmcp.PromptArgument{
			{Name: "id", Description: "Prompt ID", Required: true},
		},
	}, r.usePrompt)
}

// --- resource handlers ---

func (r *resourceHandlers) readCollections(ctx context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
	userID := authmw.GetUserID(ctx)
	colls, err := r.collections.List(ctx, userID, nil)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
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

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}
	return &sdkmcp.ReadResourceResult{
		Contents: []*sdkmcp.ResourceContents{{
			URI:  "promptvault://collections",
			Text: string(data),
		}},
	}, nil
}

func (r *resourceHandlers) readTags(ctx context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
	userID := authmw.GetUserID(ctx)
	tags, err := r.tags.List(ctx, userID, nil)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	result := make([]TagResponse, len(tags))
	for i, t := range tags {
		result[i] = TagResponse{ID: t.ID, Name: t.Name, Color: t.Color}
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}
	return &sdkmcp.ReadResourceResult{
		Contents: []*sdkmcp.ResourceContents{{
			URI:  "promptvault://tags",
			Text: string(data),
		}},
	}, nil
}

func (r *resourceHandlers) readPrompt(ctx context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
	userID := authmw.GetUserID(ctx)

	// parse ID from URI: promptvault://prompts/{id}
	uri := req.Params.URI
	// extract last path segment
	id := uint(0)
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '/' {
			parsed, err := strconv.ParseUint(uri[i+1:], 10, 32)
			if err != nil {
				return nil, sdkmcp.ResourceNotFoundError(uri)
			}
			id = uint(parsed)
			break
		}
	}
	if id == 0 {
		return nil, sdkmcp.ResourceNotFoundError(uri)
	}

	prompt, err := r.prompts.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, promptuc.ErrNotFound) {
			return nil, sdkmcp.ResourceNotFoundError(uri)
		}
		return nil, fmt.Errorf("read prompt %d: %w", id, err)
	}

	resp := toPromptResponse(prompt)
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal prompt: %w", err)
	}
	return &sdkmcp.ReadResourceResult{
		Contents: []*sdkmcp.ResourceContents{{
			URI:  uri,
			Text: string(data),
		}},
	}, nil
}

// --- MCP Prompt handler ---

func (r *resourceHandlers) usePrompt(ctx context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
	userID := authmw.GetUserID(ctx)

	idStr := req.Params.Arguments["id"]
	parsed, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || parsed == 0 {
		return nil, fmt.Errorf("invalid prompt id: %s", idStr)
	}

	prompt, err := r.prompts.GetByID(ctx, uint(parsed), userID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	// Format the prompt content for LLM use
	content := fmt.Sprintf("# %s\n\n%s", prompt.Title, prompt.Content)

	return &sdkmcp.GetPromptResult{
		Description: fmt.Sprintf("Prompt: %s", prompt.Title),
		Messages: []*sdkmcp.PromptMessage{
			{
				Role:    "user",
				Content: &sdkmcp.TextContent{Text: content},
			},
		},
	}, nil
}
