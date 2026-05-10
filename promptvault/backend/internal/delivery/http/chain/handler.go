package chain

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	chainuc "promptvault/internal/usecases/chain"
)

type Handler struct {
	svc      *chainuc.Service
	validate *validator.Validate
}

func NewHandler(svc *chainuc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// POST /api/chains
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	req, err := utils.DecodeAndValidate[CreateChainRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	c, err := h.svc.Create(r.Context(), userID, utils.SanitizeString(req.Name), utils.SanitizeString(req.Description), req.TeamID)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteCreated(w, NewChainResponse(*c))
}

// GET /api/chains?team_id=N&limit=20&offset=0
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	var teamIDs []uint
	if tid := r.URL.Query().Get("team_id"); tid != "" {
		id, err := strconv.ParseUint(tid, 10, 32)
		if err != nil {
			httperr.Respond(w, httperr.BadRequest("Неверный team_id"))
			return
		}
		teamIDs = []uint{uint(id)}
	}
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	offset := parseInt(r.URL.Query().Get("offset"), 0)

	rows, total, err := h.svc.ListWithStats(r.Context(), userID, teamIDs, limit, offset)
	if err != nil {
		respondError(w, r, err)
		return
	}
	items := make([]ChainListItem, len(rows))
	for i, row := range rows {
		items[i] = NewChainListItem(row)
	}
	utils.WriteOK(w, ChainListResponse{Items: items, Total: total, Limit: limit, Offset: offset})
}

// GET /api/chains/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	c, err := h.svc.GetByIDWithSteps(r.Context(), id, userID)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, NewChainDetailResponse(*c))
}

// PUT /api/chains/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	req, err := utils.DecodeAndValidate[UpdateChainRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	c, err := h.svc.Update(r.Context(), id, userID, utils.SanitizeString(req.Name), utils.SanitizeString(req.Description))
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, NewChainResponse(*c))
}

// DELETE /api/chains/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	id, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteNoContent(w)
}

// POST /api/chains/{id}/steps
func (h *Handler) AddStep(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	chainID, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	req, err := utils.DecodeAndValidate[AddStepRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	step, err := h.svc.AddStep(r.Context(), chainID, userID, chainuc.AddStepInput{
		PromptID:         req.PromptID,
		Name:             utils.SanitizeString(req.Name),
		VariableMapping:  req.VariableMapping,
		ManualCheckpoint: req.ManualCheckpoint,
		StepType:         req.StepType,
		Conditions:       req.Conditions,
		AfterStepID:      req.AfterStepID,
		ParentForkID:     req.ParentForkID,
		BranchIndex:      req.BranchIndex,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteCreated(w, NewStepResponse(*step))
}

// PUT /api/chains/{id}/steps/{step_id}
func (h *Handler) UpdateStep(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	chainID, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	stepID, err := parseURLID(r, "step_id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный step_id"))
		return
	}
	req, err := utils.DecodeAndValidate[UpdateStepRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	step, err := h.svc.UpdateStep(r.Context(), chainID, stepID, userID, chainuc.UpdateStepInput{
		Name:             utils.SanitizeString(req.Name),
		VariableMapping:  req.VariableMapping,
		ManualCheckpoint: req.ManualCheckpoint,
		StepType:         req.StepType,
		Conditions:       req.Conditions,
	})
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, NewStepResponse(*step))
}

// DELETE /api/chains/{id}/steps/{step_id}
func (h *Handler) RemoveStep(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	chainID, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	stepID, err := parseURLID(r, "step_id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный step_id"))
		return
	}
	if err := h.svc.RemoveStep(r.Context(), chainID, stepID, userID); err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteNoContent(w)
}

// POST /api/chains/{id}/steps/{step_id}/move-up
func (h *Handler) MoveStepUp(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	chainID, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	stepID, err := parseURLID(r, "step_id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный step_id"))
		return
	}
	if err := h.svc.MoveStepUp(r.Context(), chainID, stepID, userID); err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteNoContent(w)
}

// POST /api/chains/{id}/steps/{step_id}/move-down
func (h *Handler) MoveStepDown(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	chainID, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	stepID, err := parseURLID(r, "step_id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный step_id"))
		return
	}
	if err := h.svc.MoveStepDown(r.Context(), chainID, stepID, userID); err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteNoContent(w)
}

// POST /api/chains/{id}/reorder
func (h *Handler) ReorderSteps(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	chainID, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	req, err := utils.DecodeAndValidate[ReorderStepsRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	if err := h.svc.ReorderSteps(r.Context(), chainID, userID, req.StepIDs); err != nil {
		respondError(w, r, err)
		return
	}
	c, err := h.svc.GetByIDWithSteps(r.Context(), chainID, userID)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, NewChainDetailResponse(*c))
}

// POST /api/chains/{id}/executions
func (h *Handler) StartExecution(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	chainID, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	req, err := utils.DecodeAndValidate[StartExecutionRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	exec, err := h.svc.StartExecution(r.Context(), chainID, userID, req.InitialVars)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteCreated(w, NewExecutionResponse(*exec))
}

// GET /api/chains/{id}/executions?limit=50
// Возвращает список последних N запусков цепочки. RBAC через chain.checkReadAccess.
func (h *Handler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	chainID, err := parseURLID(r, "id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный ID"))
		return
	}
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	executions, err := h.svc.ListExecutions(r.Context(), chainID, userID, limit)
	if err != nil {
		respondError(w, r, err)
		return
	}
	items := make([]ExecutionSummary, len(executions))
	for i, e := range executions {
		items[i] = NewExecutionSummary(e)
	}
	utils.WriteOK(w, ExecutionListResponse{Items: items, Limit: limit})
}

// GET /api/executions/{exec_id}
func (h *Handler) GetExecution(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	execID, err := parseURLID(r, "exec_id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный exec_id"))
		return
	}
	exec, err := h.svc.GetExecution(r.Context(), execID, userID)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, NewExecutionResponse(*exec))
}

// POST /api/executions/{exec_id}/advance
func (h *Handler) AdvanceStep(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	execID, err := parseURLID(r, "exec_id")
	if err != nil {
		httperr.Respond(w, httperr.BadRequest("Неверный exec_id"))
		return
	}
	req, err := utils.DecodeAndValidate[AdvanceStepRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	exec, err := h.svc.AdvanceStep(r.Context(), execID, userID, req.StepOutput, req.ChosenBranchIndex)
	if err != nil {
		respondError(w, r, err)
		return
	}
	utils.WriteOK(w, NewExecutionResponse(*exec))
}

func parseURLID(r *http.Request, name string) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, name), 10, 32)
	return uint(id), err
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return def
	}
	return v
}
