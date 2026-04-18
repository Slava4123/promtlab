package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/template"
	promptuc "promptvault/internal/usecases/prompt"
)

// maxVarsJSONSize — лимит размера JSON-строки vars в use_prompt (10 KB).
// Защита от аномальных payload'ов. На всех практических промптах хватает с запасом.
const maxVarsJSONSize = 10 * 1024

// validRoles — допустимые значения аргумента role в use_prompt.
// Соответствуют MCP spec PromptMessage.Role.
var validRoles = map[string]sdkmcp.Role{
	"user":      "user",
	"assistant": "assistant",
}

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

	// MCP Prompt — use_prompt: fetch a prompt, substitute {{vars}}, return as a message
	server.AddPrompt(&sdkmcp.Prompt{
		Name:        "use_prompt",
		Description: "Fetch a prompt from your library, substitute {{variables}}, and return as a message. Use list_prompt_vars to discover which variables a prompt needs.",
		Arguments: []*sdkmcp.PromptArgument{
			{Name: "id", Description: "Prompt ID", Required: true},
			{Name: "vars", Description: "JSON object with values for {{variables}}: {\"name\":\"Alice\",\"lang\":\"Go\"}. Required if the prompt contains variables."},
			{Name: "role", Description: "Message role: user (default) or assistant"},
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

	uri := req.Params.URI
	idx := strings.LastIndex(uri, "/")
	if idx < 0 || idx == len(uri)-1 {
		return nil, sdkmcp.ResourceNotFoundError(uri)
	}
	parsed, err := strconv.ParseUint(uri[idx+1:], 10, 32)
	if err != nil || parsed == 0 {
		return nil, sdkmcp.ResourceNotFoundError(uri)
	}

	prompt, err := r.prompts.GetByID(ctx, uint(parsed), userID)
	if err != nil {
		if errors.Is(err, promptuc.ErrNotFound) {
			return nil, sdkmcp.ResourceNotFoundError(uri)
		}
		return nil, fmt.Errorf("read prompt %d: %w", parsed, err)
	}

	data, err := json.Marshal(toPromptResponse(prompt))
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

	// 1. Parse id (required)
	idStr := req.Params.Arguments["id"]
	parsed, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || parsed == 0 {
		return nil, fmt.Errorf("use_prompt: invalid prompt id %q", idStr)
	}

	// 2. Parse role (optional, default "user")
	role := sdkmcp.Role("user")
	if raw := strings.TrimSpace(req.Params.Arguments["role"]); raw != "" {
		resolved, ok := validRoles[strings.ToLower(raw)]
		if !ok {
			return nil, fmt.Errorf("use_prompt: invalid role %q: must be user or assistant", raw)
		}
		role = resolved
	}

	// 3. Parse vars (optional, JSON object)
	var vars map[string]string
	if raw := strings.TrimSpace(req.Params.Arguments["vars"]); raw != "" {
		if len(raw) > maxVarsJSONSize {
			return nil, fmt.Errorf("use_prompt: vars exceed %d bytes limit", maxVarsJSONSize)
		}
		if err := json.Unmarshal([]byte(raw), &vars); err != nil {
			return nil, fmt.Errorf("use_prompt: vars must be a JSON object of string→string: %w", err)
		}
	}

	// 4. Load the prompt
	prompt, err := r.prompts.GetByID(ctx, uint(parsed), userID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	// 5. Render template
	rendered, missing := template.Render(prompt.Content, vars)
	if len(missing) > 0 {
		slog.Warn("mcp.use_prompt.missing_vars",
			"user_id", userID,
			"prompt_id", prompt.ID,
			"missing", missing,
		)
		return nil, fmt.Errorf(
			"use_prompt: prompt requires variables: %s. Pass them as vars=%s",
			strings.Join(missing, ", "),
			jsonSkeletonForVars(missing),
		)
	}

	slog.Debug("mcp.use_prompt.rendered",
		"user_id", userID,
		"prompt_id", prompt.ID,
		"role", string(role),
		"vars_count", len(vars),
	)

	body := fmt.Sprintf("# %s\n\n%s", prompt.Title, rendered)
	return &sdkmcp.GetPromptResult{
		Description: fmt.Sprintf("Prompt: %s", prompt.Title),
		Messages: []*sdkmcp.PromptMessage{{
			Role:    role,
			Content: &sdkmcp.TextContent{Text: body},
		}},
	}, nil
}

// jsonSkeletonForVars строит подсказку-скелет JSON для ошибки missing vars.
// Результат: `{"name":"","lang":""}` для [name, lang].
func jsonSkeletonForVars(names []string) string {
	skeleton := make(map[string]string, len(names))
	for _, n := range names {
		skeleton[n] = ""
	}
	data, err := json.Marshal(skeleton)
	if err != nil {
		return "{}"
	}
	return string(data)
}
