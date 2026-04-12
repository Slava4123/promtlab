package badge

import (
	"context"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	badgeuc "promptvault/internal/usecases/badge"
)

// Service — локальный интерфейс для того, что Handler использует от badge
// usecase. Объявлен здесь (на consumer side), идиоматично для Go. Позволяет
// подставить fake в handler_test.go без импорта внутренних типов usecase-слоя.
// *badgeuc.Service удовлетворяет этому интерфейсу.
type Service interface {
	List(ctx context.Context, userID uint) ([]badgeuc.BadgeWithState, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/badges
// Auth: JWT required (монтируется в protected group в app.MountRoutes).
// Returns: BadgeListResponse со всеми бейджами каталога + состоянием юзера.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httperr.Respond(w, httperr.Internal(err))
		return
	}
	utils.WriteOK(w, NewBadgeListResponse(items))
}
