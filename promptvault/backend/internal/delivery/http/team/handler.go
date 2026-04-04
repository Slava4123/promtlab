package team

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	"promptvault/internal/models"
	teamuc "promptvault/internal/usecases/team"
)

type Handler struct {
	svc      *teamuc.Service
	validate *validator.Validate
}

func NewHandler(svc *teamuc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// GET /api/teams
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	resp := make([]TeamResponse, len(items))
	for i, item := range items {
		resp[i] = NewTeamResponse(&item.Team, item.Role, item.MemberCount)
	}

	utils.WriteOK(w, resp)
}

// POST /api/teams
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[CreateTeamRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	team, err := h.svc.Create(r.Context(), userID, teamuc.CreateInput{
		Name:        utils.SanitizeString(req.Name),
		Description: utils.SanitizeString(req.Description),
	})
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteCreated(w, NewTeamResponse(team, models.RoleOwner, 1))
}

// GET /api/teams/{slug}
func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	team, members, err := h.svc.GetBySlug(r.Context(), slug, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	// Найти роль текущего пользователя из members
	role := models.RoleViewer // fallback
	for _, m := range members {
		if m.UserID == userID {
			role = m.Role
			break
		}
	}

	utils.WriteOK(w, NewTeamDetailResponse(team, role, members))
}

// PUT /api/teams/{slug}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	req, err := utils.DecodeAndValidate[UpdateTeamRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	_, err = h.svc.Update(r.Context(), slug, userID, teamuc.UpdateInput{
		Name:        utils.SanitizeStringPtr(req.Name),
		Description: utils.SanitizeStringPtr(req.Description),
	})
	if err != nil {
		respondError(w, err)
		return
	}

	// Перезагрузить для актуальных members
	team, members, err := h.svc.GetBySlug(r.Context(), slug, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	// Найти роль текущего пользователя из members
	role := models.RoleViewer // fallback
	for _, m := range members {
		if m.UserID == userID {
			role = m.Role
			break
		}
	}

	utils.WriteOK(w, NewTeamResponse(team, role, len(members)))
}

// DELETE /api/teams/{slug}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	if err := h.svc.Delete(r.Context(), slug, userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}

// POST /api/teams/{slug}/invitations
func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	req, err := utils.DecodeAndValidate[AddMemberRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	inv, err := h.svc.InviteMember(r.Context(), slug, userID, teamuc.AddMemberInput{
		Query: req.Query,
		Role:  models.TeamRole(req.Role),
	})
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteCreated(w, NewInvitationResponse(*inv))
}

// GET /api/teams/{slug}/invitations
func (h *Handler) ListTeamInvitations(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	invitations, err := h.svc.ListTeamInvitations(r.Context(), slug, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	resp := make([]PendingInvitationResponse, len(invitations))
	for i, inv := range invitations {
		resp[i] = NewPendingInvitationResponse(inv)
	}

	utils.WriteOK(w, resp)
}

// DELETE /api/teams/{slug}/invitations/{invitationId}
func (h *Handler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")
	invID, err := parseInvitationID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID приглашения"))
		return
	}

	if err := h.svc.CancelInvitation(r.Context(), slug, userID, invID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}

// GET /api/invitations
func (h *Handler) ListMyInvitations(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	invitations, err := h.svc.ListMyInvitations(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	resp := make([]InvitationResponse, len(invitations))
	for i, inv := range invitations {
		resp[i] = NewInvitationResponse(inv)
	}

	utils.WriteOK(w, resp)
}

// POST /api/invitations/{invitationId}/accept
func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	invID, err := parseInvitationID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID приглашения"))
		return
	}

	if err := h.svc.AcceptInvitation(r.Context(), invID, userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}

// POST /api/invitations/{invitationId}/decline
func (h *Handler) DeclineInvitation(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	invID, err := parseInvitationID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID приглашения"))
		return
	}

	if err := h.svc.DeclineInvitation(r.Context(), invID, userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}

// PUT /api/teams/{slug}/members/{userId}
func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")
	targetUserID, err := parseUserID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID пользователя"))
		return
	}

	req, err := utils.DecodeAndValidate[UpdateMemberRoleRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.svc.UpdateMemberRole(r.Context(), slug, userID, targetUserID, models.TeamRole(req.Role)); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}

// DELETE /api/teams/{slug}/members/{userId}
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")
	targetUserID, err := parseUserID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID пользователя"))
		return
	}

	if err := h.svc.RemoveMember(r.Context(), slug, userID, targetUserID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}

func parseUserID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "userId"), 10, 32)
	return uint(id), err
}

func parseInvitationID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "invitationId"), 10, 32)
	return uint(id), err
}
