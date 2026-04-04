package user

import (
	"log/slog"
	"net/http"
	"strings"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	useruc "promptvault/internal/usecases/user"
)

type Handler struct {
	svc *useruc.Service
}

func NewHandler(svc *useruc.Service) *Handler {
	return &Handler{svc: svc}
}

type UserSearchResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Email     string `json:"email"`
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	results, err := h.svc.Search(r.Context(), q, 5)
	if err != nil {
		slog.Error("user search failed", "query", q, "error", err)
		httperr.Respond(w, httperr.Internal(err))
		return
	}
	resp := make([]UserSearchResponse, 0, len(results))
	for _, r := range results {
		resp = append(resp, UserSearchResponse{
			ID: r.ID, Name: r.Name, Username: r.Username,
			AvatarURL: r.AvatarURL, Email: r.Email,
		})
	}
	utils.WriteOK(w, resp)
}
