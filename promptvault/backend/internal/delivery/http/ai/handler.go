package ai

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	"promptvault/internal/infrastructure/openrouter"
	authmw "promptvault/internal/middleware/auth"
	aiuc "promptvault/internal/usecases/ai"
)

type Handler struct {
	svc      *aiuc.Service
	validate *validator.Validate
}

func NewHandler(svc *aiuc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// GET /api/ai/models
func (h *Handler) Models(w http.ResponseWriter, r *http.Request) {
	models := h.svc.Models()
	utils.WriteOK(w, toModelResponses(models))
}

// POST /api/ai/enhance
func (h *Handler) Enhance(w http.ResponseWriter, r *http.Request) {
	var req EnhanceRequest
	h.streamEndpoint(w, r, &req, func(userID uint, cb openrouter.StreamCallback) error {
		return h.svc.Enhance(r.Context(), aiuc.EnhanceInput{
			UserID:  userID,
			Content: req.Content,
			Model:   req.Model,
		}, cb)
	})
}

// POST /api/ai/rewrite
func (h *Handler) Rewrite(w http.ResponseWriter, r *http.Request) {
	var req RewriteRequest
	h.streamEndpoint(w, r, &req, func(userID uint, cb openrouter.StreamCallback) error {
		return h.svc.Rewrite(r.Context(), aiuc.RewriteInput{
			UserID:  userID,
			Content: req.Content,
			Model:   req.Model,
			Style:   aiuc.RewriteStyle(req.Style),
		}, cb)
	})
}

// POST /api/ai/analyze
func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	h.streamEndpoint(w, r, &req, func(userID uint, cb openrouter.StreamCallback) error {
		return h.svc.Analyze(r.Context(), aiuc.AnalyzeInput{
			UserID:  userID,
			Content: req.Content,
			Model:   req.Model,
		}, cb)
	})
}

// POST /api/ai/variations
func (h *Handler) Variations(w http.ResponseWriter, r *http.Request) {
	var req VariationsRequest
	h.streamEndpoint(w, r, &req, func(userID uint, cb openrouter.StreamCallback) error {
		return h.svc.Variations(r.Context(), aiuc.VariationsInput{
			UserID:  userID,
			Content: req.Content,
			Model:   req.Model,
			Count:   req.Count,
		}, cb)
	})
}

// streamEndpoint handles the common SSE streaming pattern:
// decode → validate → rate limit → init SSE → call service → finish.
func (h *Handler) streamEndpoint(w http.ResponseWriter, r *http.Request, req any, call func(userID uint, cb openrouter.StreamCallback) error) {
	if err := utils.DecodeJSON(r, req); err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	userID := authmw.GetUserID(r.Context())
	if err := h.svc.CheckRateLimit(userID); err != nil {
		respondError(w, err)
		return
	}

	flusher, ok := acquireFlusher(w)
	if !ok {
		return
	}
	initSSE(w)
	flusher.Flush()

	err := call(userID, streamCallback(w, flusher))
	finishSSE(w, flusher, err)
}

// SSE helpers

func acquireFlusher(w http.ResponseWriter) (http.Flusher, bool) {
	f, ok := w.(http.Flusher)
	if !ok {
		httperr.Respond(w, httperr.Internal(errors.New("streaming not supported")))
	}
	return f, ok
}

func initSSE(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
}

func streamCallback(w http.ResponseWriter, f http.Flusher) openrouter.StreamCallback {
	return func(chunk string) error {
		for _, line := range strings.Split(chunk, "\n") {
			if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "\n"); err != nil {
			return err
		}
		f.Flush()
		return nil
	}
}

func finishSSE(w http.ResponseWriter, f http.Flusher, err error) {
	if err != nil {
		slog.Error("AI stream error", "error", err)
		msg := userFriendlyError(err)
		if _, werr := fmt.Fprintf(w, "event: error\ndata: %s\n\n", msg); werr != nil {
			slog.Error("failed to write SSE error event", "error", werr)
		}
		f.Flush()
		return
	}
	if _, werr := fmt.Fprintf(w, "data: [DONE]\n\n"); werr != nil {
		slog.Error("failed to write SSE done event", "error", werr)
	}
	f.Flush()
}

func userFriendlyError(err error) string {
	switch {
	case errors.Is(err, openrouter.ErrUnauthorized):
		return "AI-сервис временно недоступен"
	case errors.Is(err, openrouter.ErrRateLimited):
		return "Превышен лимит запросов к AI. Попробуйте позже"
	case errors.Is(err, openrouter.ErrInsufficientCredits):
		return "AI-сервис временно недоступен"
	case errors.Is(err, openrouter.ErrEmptyResponse):
		return "Модель вернула пустой ответ. Попробуйте другую модель"
	default:
		return "Ошибка AI-сервиса. Попробуйте позже"
	}
}
