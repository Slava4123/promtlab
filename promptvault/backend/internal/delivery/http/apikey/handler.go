package apikey

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	apikeyuc "promptvault/internal/usecases/apikey"
)

type Handler struct {
	svc      *apikeyuc.Service
	maxKeys  int
	validate *validator.Validate
}

func NewHandler(svc *apikeyuc.Service, maxKeys int) *Handler {
	return &Handler{svc: svc, maxKeys: maxKeys, validate: validator.New()}
}

// GET /api/api-keys
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	keys, err := h.svc.List(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	resp := make([]APIKeyResponse, len(keys))
	for i, k := range keys {
		resp[i] = toAPIKeyResponse(k)
	}

	utils.WriteOK(w, ListResponse{Keys: resp, MaxKeys: h.maxKeys})
}

// POST /api/api-keys
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[CreateRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	plaintext, info, err := h.svc.Create(r.Context(), userID, req.Name)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteCreated(w, CreatedAPIKeyResponse{
		ID:        info.ID,
		Name:      info.Name,
		Key:       plaintext,
		KeyPrefix: info.KeyPrefix,
		CreatedAt: info.CreatedAt,
	})
}

// DELETE /api/api-keys/{id}
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	if err := h.svc.Revoke(r.Context(), uint(id), userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}
