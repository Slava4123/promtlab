package mcpserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/getsentry/sentry-go"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	authmw "promptvault/internal/middleware/auth"
)

// notifier — обёртка над *mcp.Server.ResourceUpdated для ресурсных URI.
// SDK сам трекает подписки через SubscribeHandler/UnsubscribeHandler и
// рассылает уведомления подписчикам; наша задача — своевременно вызывать
// Notify после CUD операций.
//
// Все методы non-blocking: SDK реализует ResourceUpdated как put в очередь сессии.
type notifier struct {
	server *sdkmcp.Server
}

func newNotifier(server *sdkmcp.Server) *notifier {
	return &notifier{server: server}
}

// NotifyCollections вызывается при create/delete/update коллекции через MCP.
func (n *notifier) NotifyCollections(ctx context.Context) {
	n.notify(ctx, "promptvault://collections")
}

// NotifyTags вызывается при create/delete тега через MCP.
func (n *notifier) NotifyTags(ctx context.Context) {
	n.notify(ctx, "promptvault://tags")
}

// NotifyPrompt вызывается при update/delete/favorite/pin/revert/share промпта через MCP.
func (n *notifier) NotifyPrompt(ctx context.Context, promptID uint) {
	n.notify(ctx, fmt.Sprintf("promptvault://prompts/%d", promptID))
}

func (n *notifier) notify(ctx context.Context, uri string) {
	if n == nil || n.server == nil {
		return
	}
	userID := authmw.GetUserID(ctx)
	// Detach from request ctx: notification заказан после успешного ответа,
	// отмена request-ctx после return не должна терять уведомление.
	ncCtx := context.WithoutCancel(ctx)
	err := n.server.ResourceUpdated(ncCtx, &sdkmcp.ResourceUpdatedNotificationParams{URI: uri})
	if err != nil {
		slog.Error("mcp.notification.failed", "uri", uri, "user_id", userID, "error", err)
		// Инфраструктурная ошибка — подписанные сессии навсегда stale, ops должен знать.
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.CaptureException(fmt.Errorf("mcp notification failed for %s: %w", uri, err))
		}
		return
	}
	slog.Debug("mcp.notification.sent", "uri", uri, "user_id", userID)
}
