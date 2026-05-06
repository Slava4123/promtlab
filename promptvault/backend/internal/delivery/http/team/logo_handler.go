package team

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	"promptvault/internal/infrastructure/metrics"
	repo "promptvault/internal/interface/repository"
	authmw "promptvault/internal/middleware/auth"
	teamuc "promptvault/internal/usecases/team"
)

// LogoHandler — Phase 16-X. POST/DELETE/GET /api/teams/{slug}/branding/logo.
// POST/DELETE — protected (owner Max-only); GET — public с ETag/Cache-Control.
type LogoHandler struct {
	teams *teamuc.Service
}

func NewLogoHandler(teams *teamuc.Service) *LogoHandler {
	return &LogoHandler{teams: teams}
}

// maxLogoUploadBytes — handler-level guard через MaxBytesReader. Чуть больше
// payload-лимита usecase'а: +4 KiB на multipart boundary/headers, чтобы 1 МБ
// файла «честно» прошёл, а 1 МБ + 1 байт payload отвергся 413.
const maxLogoUploadBytes = teamuc.MaxLogoFileSize + 4096

// LogoUploadResponse — JSON-ответ при успешном POST /logo.
type LogoUploadResponse struct {
	LogoSource       string `json:"logo_source"`
	EffectiveLogoURL string `json:"effective_logo_url"`
	SizeBytes        int64  `json:"size_bytes"`
	ContentType      string `json:"content_type"`
}

// Upload — POST /api/teams/{slug}/branding/logo.
// multipart/form-data, поле "file". Owner+Max gate в usecase.
func (h *LogoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	r.Body = http.MaxBytesReader(w, r.Body, maxLogoUploadBytes)
	file, _, err := r.FormFile("file")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrMissingFile):
			metrics.TeamBrandingLogoUploads.WithLabelValues("bad_format").Inc()
			httperr.Respond(w, httperr.BadRequest("Файл не передан в поле 'file'"))
		case isMaxBytesError(err):
			metrics.TeamBrandingLogoUploads.WithLabelValues("too_large").Inc()
			httperr.Respond(w, httperr.PayloadTooLarge("Файл больше 1 МБ"))
		default:
			metrics.TeamBrandingLogoUploads.WithLabelValues("other").Inc()
			httperr.Respond(w, httperr.BadRequest("Не удалось прочитать multipart: "+err.Error()))
		}
		return
	}
	defer func() { _ = file.Close() }()

	saved, err := h.teams.UploadLogo(r.Context(), slug, userID, file)
	if err != nil {
		metrics.TeamBrandingLogoUploads.WithLabelValues(uploadResultLabel(err)).Inc()
		respondLogoError(w, r, err)
		return
	}

	metrics.TeamBrandingLogoUploads.WithLabelValues("success").Inc()
	metrics.TeamBrandingLogoSizeBytes.Observe(float64(saved.SizeBytes))
	slog.Info("team.branding.logo.uploaded",
		"slug", slug,
		"user_id", userID,
		"team_id", saved.TeamID,
		"content_type", saved.ContentType,
		"size_bytes", saved.SizeBytes,
		"sha256", saved.SHA256,
	)

	utils.WriteOK(w, LogoUploadResponse{
		LogoSource:       "file",
		EffectiveLogoURL: "/api/teams/" + slug + "/branding/logo",
		SizeBytes:        saved.SizeBytes,
		ContentType:      saved.ContentType,
	})
}

// uploadResultLabel мапит usecase-ошибку в Prometheus label.
// Контракт: success | too_large | bad_format | forbidden | other (см. metrics.go).
func uploadResultLabel(err error) string {
	switch {
	case errors.Is(err, teamuc.ErrLogoFileTooLarge):
		return "too_large"
	case errors.Is(err, teamuc.ErrLogoFileBadFormat),
		errors.Is(err, teamuc.ErrLogoFileMissing),
		errors.Is(err, teamuc.ErrLogoImageTooLarge):
		return "bad_format"
	case errors.Is(err, teamuc.ErrBrandingMaxOnly),
		errors.Is(err, teamuc.ErrForbidden),
		errors.Is(err, teamuc.ErrNotOwner):
		return "forbidden"
	default:
		return "other"
	}
}

// Delete — DELETE /api/teams/{slug}/branding/logo. Owner+Max.
// Идемпотентно: если файла не было, всё равно ставим source='none'.
func (h *LogoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	if err := h.teams.DeleteLogo(r.Context(), slug, userID); err != nil {
		respondLogoError(w, r, err)
		return
	}
	utils.WriteOK(w, map[string]string{"logo_source": "none"})
}

// Serve — GET /api/teams/{slug}/branding/logo. Без auth (public).
//   - ETag = "<sha256>"; If-None-Match → 304 Not Modified.
//   - Cache-Control: public, max-age=86400 (24ч). НЕ immutable: при замене
//     файла sha256 меняется → ETag меняется → revalidate возвращает новый payload.
//   - 404 одинаков для «нет команды» и «нет файла» — защита от enumeration.
func (h *LogoHandler) Serve(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	file, err := h.teams.GetLogo(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) || errors.Is(err, teamuc.ErrNotFound) {
			httperr.Respond(w, httperr.NotFound("Логотип не найден"))
			return
		}
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
		return
	}

	etag := `"` + file.SHA256 + `"`
	if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
		metrics.TeamBrandingLogoServe.WithLabelValues("hit").Inc()
		w.Header().Set("ETag", etag)
		w.WriteHeader(http.StatusNotModified)
		return
	}
	metrics.TeamBrandingLogoServe.WithLabelValues("miss").Inc()
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.SizeBytes, 10))
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("ETag", etag)
	if _, werr := w.Write(file.Bytes); werr != nil {
		slog.Error("logo write failed", "error", werr, "slug", slug)
	}
}

// isMaxBytesError детектит ошибку от MaxBytesReader — std lib не экспортирует
// типизированную ошибку, поэтому совпадение по тексту. См. net/http source.
func isMaxBytesError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "request body too large")
}

func respondLogoError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, teamuc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound("Команда не найдена"))
	case errors.Is(err, teamuc.ErrForbidden), errors.Is(err, teamuc.ErrNotOwner):
		httperr.Respond(w, httperr.Forbidden("Нет доступа"))
	case errors.Is(err, teamuc.ErrBrandingMaxOnly):
		httperr.RespondQuotaError(w, "branding", 0, 0, "pro",
			"Загрузка логотипа доступна только на тарифе Max. Обновитесь на странице /pricing.")
	case errors.Is(err, teamuc.ErrLogoFileMissing):
		httperr.Respond(w, httperr.BadRequest("Файл пустой"))
	case errors.Is(err, teamuc.ErrLogoFileTooLarge):
		httperr.Respond(w, httperr.PayloadTooLarge("Файл больше 1 МБ"))
	case errors.Is(err, teamuc.ErrLogoFileBadFormat):
		httperr.Respond(w, httperr.UnsupportedMediaType(err.Error()))
	case errors.Is(err, teamuc.ErrLogoImageTooLarge):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, teamuc.ErrLogoStorageDisabled):
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
	default:
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
	}
}
