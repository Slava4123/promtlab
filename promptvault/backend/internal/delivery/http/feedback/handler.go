package feedback

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	feedbackuc "promptvault/internal/usecases/feedback"
)

type Handler struct {
	svc      *feedbackuc.Service
	validate *validator.Validate
}

func NewHandler(svc *feedbackuc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// POST /api/feedback
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[SubmitRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.Submit(r.Context(), feedbackuc.SubmitInput{
		UserID:  userID,
		Type:    req.Type,
		Message: req.Message,
		PageURL: req.PageURL,
	})
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteCreated(w, SubmitResponse{
		ID:      result.ID,
		Message: "Спасибо за обратную связь!",
	})
}
