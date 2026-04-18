package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"

	"github.com/getsentry/sentry-go"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/template"
	promptuc "promptvault/internal/usecases/prompt"
)

const (
	idCompletionLimit = 20
)

// makeCompletionHandler возвращает обработчик MCP `completion/complete` для
// аргументов prompt `use_prompt`.
//
// Supported refs/arguments:
//   - ref/prompt + name=use_prompt + argument.name=id
//     → префиксный поиск по промптам юзера, возвращает до 20 ID как строки.
//   - ref/prompt + name=use_prompt + argument.name=vars + context.arguments[id]=<N>
//     → извлекает {{vars}} из промпта, возвращает JSON-скелет как одно значение.
//   - ref/prompt + name=use_prompt + argument.name=role
//     → возвращает ["user","assistant"] с фильтром по value prefix.
//
// Для всех остальных запросов — пустой CompletionResultDetails (валидно по MCP).
//
// Безопасность: handler работает в ctx авторизованного юзера (API-key middleware
// выставил UserIDKey и KeyPolicy); чужие промпты недоступны через List()
// благодаря фильтру UserID.
func makeCompletionHandler(prompts PromptService) func(ctx context.Context, req *sdkmcp.CompleteRequest) (*sdkmcp.CompleteResult, error) {
	return func(ctx context.Context, req *sdkmcp.CompleteRequest) (*sdkmcp.CompleteResult, error) {
		empty := &sdkmcp.CompleteResult{
			Completion: sdkmcp.CompletionResultDetails{Values: []string{}},
		}

		if req == nil || req.Params == nil || req.Params.Ref == nil {
			return empty, nil
		}

		ref := req.Params.Ref
		if ref.Type != "ref/prompt" || ref.Name != "use_prompt" {
			return empty, nil
		}

		userID := authmw.GetUserID(ctx)
		arg := req.Params.Argument

		switch arg.Name {
		case "id":
			return completeID(ctx, prompts, userID, arg.Value)
		case "role":
			return completeRole(arg.Value), nil
		case "vars":
			return completeVars(ctx, prompts, userID, req.Params.Context)
		default:
			return empty, nil
		}
	}
}

func completeID(ctx context.Context, prompts PromptService, userID uint, prefix string) (*sdkmcp.CompleteResult, error) {
	empty := &sdkmcp.CompleteResult{
		Completion: sdkmcp.CompletionResultDetails{Values: []string{}},
	}
	list, _, err := prompts.List(ctx, repo.PromptListFilter{
		UserID:   userID,
		Query:    prefix,
		Page:     0,
		PageSize: idCompletionLimit,
	})
	if err != nil {
		// По MCP spec completions — best-effort, клиенту возвращаем пустой ответ.
		// Но инфра-ошибку ловим в Sentry, иначе ops не узнает о падении БД.
		slog.Error("mcp.completion.id.list_failed", "user_id", userID, "error", err)
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.CaptureException(err)
		}
		return empty, nil
	}

	values := make([]string, 0, len(list))
	for i := range list {
		values = append(values, strconv.FormatUint(uint64(list[i].ID), 10))
	}
	slog.Debug("mcp.completion.id", "user_id", userID, "prefix_len", len(prefix), "results", len(values))

	return &sdkmcp.CompleteResult{
		Completion: sdkmcp.CompletionResultDetails{
			Values:  values,
			Total:   len(values),
			HasMore: len(values) >= idCompletionLimit,
		},
	}, nil
}

func completeRole(prefix string) *sdkmcp.CompleteResult {
	candidates := []string{"user", "assistant"}
	values := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if prefix == "" || startsWith(c, prefix) {
			values = append(values, c)
		}
	}
	return &sdkmcp.CompleteResult{
		Completion: sdkmcp.CompletionResultDetails{
			Values: values,
			Total:  len(values),
		},
	}
}

func completeVars(ctx context.Context, prompts PromptService, userID uint, cctx *sdkmcp.CompleteContext) (*sdkmcp.CompleteResult, error) {
	empty := &sdkmcp.CompleteResult{
		Completion: sdkmcp.CompletionResultDetails{Values: []string{}},
	}
	if cctx == nil || cctx.Arguments == nil {
		return empty, nil
	}
	idStr := cctx.Arguments["id"]
	if idStr == "" {
		return empty, nil
	}
	parsed, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || parsed == 0 {
		return empty, nil
	}

	prompt, err := prompts.GetByID(ctx, uint(parsed), userID)
	if err != nil {
		// 404/403 — клиенту не утечём причину (нет oracle для чужих промптов).
		// Прочее — инфра-ошибка, ловим в Sentry.
		if !errors.Is(err, promptuc.ErrNotFound) && !errors.Is(err, promptuc.ErrForbidden) {
			slog.Error("mcp.completion.vars.fetch_failed", "user_id", userID, "prompt_id", parsed, "error", err)
			if hub := sentry.GetHubFromContext(ctx); hub != nil {
				hub.CaptureException(err)
			}
		}
		return empty, nil
	}

	vars := template.Extract(prompt.Content)
	if len(vars) == 0 {
		return empty, nil
	}
	skeleton := make(map[string]string, len(vars))
	for _, n := range vars {
		skeleton[n] = ""
	}
	data, err := json.Marshal(skeleton)
	if err != nil {
		return empty, nil
	}
	slog.Debug("mcp.completion.vars", "user_id", userID, "prompt_id", prompt.ID, "vars_count", len(vars))
	return &sdkmcp.CompleteResult{
		Completion: sdkmcp.CompletionResultDetails{
			Values: []string{string(data)},
			Total:  1,
		},
	}, nil
}

func startsWith(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	return s[:len(prefix)] == prefix
}
