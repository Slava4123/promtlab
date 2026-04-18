package mcpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/middleware/ratelimit"
	apikeyuc "promptvault/internal/usecases/apikey"
	quotauc "promptvault/internal/usecases/quota"
)

type MCPServer struct {
	server  *mcp.Server
	handler http.Handler
}

const serverInstructions = `PromptLab MCP Server — управление AI-промптами.

Все tools поддерживают параметр team_id для работы в командном пространстве.
Без team_id — личное пространство пользователя.

Рабочий процесс:
1. search_prompts / list_prompts / search_suggest — поиск и просмотр
2. get_prompt — получение полного содержимого
3. create_prompt / update_prompt — создание и редактирование
4. prompt_favorite / prompt_pin — организация библиотеки
5. prompt_list_pinned / prompt_list_recent — быстрый доступ
6. share_create / share_deactivate — управление ссылками
7. get_prompt_versions / prompt_revert — история и откат
8. prompt_increment_usage — трекинг использования

Коллекции и теги:
- collection_get / collection_update / create_collection / delete_collection
- create_tag / tag_delete / list_tags

Командные пространства:
- list_teams — список команд пользователя с ролью и количеством участников.
  Возвращённый team.id передавайте в team_id других инструментов,
  чтобы работать с промптами/коллекциями/тегами внутри команды.

Аккаунт:
- whoami — информация о текущем пользователе (id, email, plan, default_model).

Корзина:
- list_trash — посмотреть удалённые промпты (soft-delete, 30 дней)
- restore_prompt — восстановить промпт из корзины
- purge_prompt — удалить навсегда (необратимо)

Ролевые ограничения в командах:
- owner/editor: полный доступ (чтение + запись)
- viewer: только чтение (search, list, get)

Удаление: delete_prompt перемещает в корзину (восстановимо 30 дней).`

func NewMCPServer(
	apiKeySvc *apikeyuc.Service,
	promptSvc PromptService,
	collSvc CollectionService,
	tagSvc TagService,
	searchSvc SearchService,
	shareSvc ShareService,
	teamSvc TeamService,
	trashSvc TrashService,
	userSvc UserService,
	userRPM int,
	quotas *quotauc.Service,
) *MCPServer {
	logger := slog.Default().With("component", "mcp")

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "promptvault",
		Version: "v1.2.0",
	}, &mcp.ServerOptions{
		Instructions:      serverInstructions,
		Logger:            logger,
		KeepAlive:         5 * time.Minute,
		CompletionHandler: makeCompletionHandler(promptSvc),
		// SubscribeHandler / UnsubscribeHandler: no-op acceptors.
		// SDK внутри трекует подписки; серверу остаётся только вызывать
		// server.ResourceUpdated после CUD — это делает notifier.
		SubscribeHandler: func(ctx context.Context, req *mcp.SubscribeRequest) error {
			slog.Debug("mcp.subscription.added",
				"user_id", authmw.GetUserID(ctx),
				"uri", req.Params.URI,
			)
			return nil
		},
		UnsubscribeHandler: func(ctx context.Context, req *mcp.UnsubscribeRequest) error {
			slog.Debug("mcp.subscription.removed",
				"user_id", authmw.GetUserID(ctx),
				"uri", req.Params.URI,
			)
			return nil
		},
	})

	notif := newNotifier(server)

	tools := &toolHandlers{
		prompts:     promptSvc,
		collections: collSvc,
		tags:        tagSvc,
		search:      searchSvc,
		shares:      shareSvc,
		teams:       teamSvc,
		trash:       trashSvc,
		users:       userSvc,
		quotas:      quotas,
		cache:       newListCache(30 * time.Second),
		notifier:    notif,
	}
	tools.register(server)

	resources := &resourceHandlers{
		prompts:     promptSvc,
		collections: collSvc,
		tags:        tagSvc,
	}
	resources.register(server)

	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		Logger:         logger,
		SessionTimeout: 30 * time.Minute,
	})

	// Per-user rate limit (applied after auth sets userID in context).
	var handler http.Handler = mcpHandler
	if userRPM > 0 {
		handler = ratelimit.ByUserID(userRPM, func(r *http.Request) uint {
			return authmw.GetUserID(r.Context())
		})(handler)
	}

	// Auth middleware wraps everything.
	authed := APIKeyAuth(apiKeySvc)(handler)

	return &MCPServer{
		server:  server,
		handler: authed,
	}
}

func (s *MCPServer) Handler() http.Handler {
	return s.handler
}
