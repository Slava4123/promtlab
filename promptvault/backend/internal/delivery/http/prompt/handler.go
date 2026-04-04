package prompt

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	promptuc "promptvault/internal/usecases/prompt"
)

type Handler struct {
	svc      *promptuc.Service
	validate *validator.Validate
}

func NewHandler(svc *promptuc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// GET /api/prompts
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := repo.PromptListFilter{
		UserID:       userID,
		Query:        q.Get("q"),
		FavoriteOnly: q.Get("favorite") == "true",
		Page:         page,
		PageSize:     pageSize,
	}

	// Team context
	if tid := q.Get("team_id"); tid != "" {
		id, err := strconv.ParseUint(tid, 10, 32)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest("Неверный team_id"))
			return
		}
		filter.TeamIDs = []uint{uint(id)}
	}

	if cid := q.Get("collection_id"); cid != "" {
		id, err := strconv.ParseUint(cid, 10, 32)
		if err == nil {
			uid := uint(id)
			filter.CollectionID = &uid
		}
	}

	if tagParam := q.Get("tag_ids"); tagParam != "" {
		for _, p := range strings.Split(tagParam, ",") {
			id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 32)
			if err == nil {
				filter.TagIDs = append(filter.TagIDs, uint(id))
			}
		}
	}

	prompts, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WritePaginated(w, NewPromptListResponse(prompts), total, page, pageSize)
}

// POST /api/prompts
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[CreatePromptRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	p, err := h.svc.Create(r.Context(), promptuc.CreateInput{
		UserID:        userID,
		TeamID:        req.TeamID,
		Title:         strings.TrimSpace(req.Title),
		Content:       strings.TrimSpace(req.Content),
		Model:         req.Model,
		CollectionIDs: req.CollectionIDs,
		TagIDs:        req.TagIDs,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteCreated(w, NewPromptResponse(*p))
}

// GET /api/prompts/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	p, err := h.svc.GetByID(r.Context(), id, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, NewPromptResponse(*p))
}

// PUT /api/prompts/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	req, err := utils.DecodeAndValidate[UpdatePromptRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	p, err := h.svc.Update(r.Context(), id, userID, promptuc.UpdateInput{
		Title:         trimSpacePtr(req.Title),
		Content:       trimSpacePtr(req.Content),
		Model:         req.Model,
		ChangeNote:    strings.TrimSpace(req.ChangeNote),
		CollectionIDs: req.CollectionIDs,
		TagIDs:        req.TagIDs,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, NewPromptResponse(*p))
}

// DELETE /api/prompts/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteNoContent(w)
}

// POST /api/prompts/{id}/favorite
func (h *Handler) ToggleFavorite(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	p, err := h.svc.ToggleFavorite(r.Context(), id, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, NewPromptResponse(*p))
}

// POST /api/prompts/{id}/use
func (h *Handler) IncrementUsage(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	if err := h.svc.IncrementUsage(r.Context(), id, userID); err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, map[string]string{"message": "ok"})
}

// GET /api/prompts/{id}/versions
func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	versions, total, err := h.svc.ListVersions(r.Context(), id, userID, page, pageSize)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WritePaginated(w, NewVersionListResponse(versions), total, page, pageSize)
}

// POST /api/prompts/{id}/revert/{versionId}
func (h *Handler) RevertToVersion(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}

	versionID, err := parseVersionID(r)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID версии"))
		return
	}

	p, err := h.svc.RevertToVersion(r.Context(), id, userID, versionID)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, NewPromptResponse(*p))
}

func trimSpacePtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	return &v
}

func parseID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	return uint(id), err
}

func parseVersionID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "versionId"), 10, 32)
	return uint(id), err
}
