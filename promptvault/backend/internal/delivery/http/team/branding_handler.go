package team

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	teamuc "promptvault/internal/usecases/team"
)

// BrandingHandler — GET/PUT /api/teams/{slug}/branding (Phase 14 D).
// GET доступен всем членам команды; PUT — только owner на Max-тарифе.
type BrandingHandler struct {
	teams    *teamuc.Service
	validate *validator.Validate
}

func NewBrandingHandler(teams *teamuc.Service) *BrandingHandler {
	return &BrandingHandler{teams: teams, validate: validator.New()}
}

// BrandingRequest — тело PUT /branding.
type BrandingRequest struct {
	LogoURL      string `json:"logo_url" validate:"omitempty,max=500,startswith=https://"`
	Tagline      string `json:"tagline" validate:"omitempty,max=200"`
	Website      string `json:"website" validate:"omitempty,max=500,startswith=https://"`
	PrimaryColor string `json:"primary_color" validate:"omitempty,hexcolor"`
}

// BrandingResponse — GET /branding и response после PUT.
type BrandingResponse struct {
	LogoURL      string `json:"logo_url,omitempty"`
	Tagline      string `json:"tagline,omitempty"`
	Website      string `json:"website,omitempty"`
	PrimaryColor string `json:"primary_color,omitempty"`
}

// Get — GET /api/teams/{slug}/branding. Доступен любому члену команды.
func (h *BrandingHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	info, err := h.teams.GetBranding(r.Context(), slug, userID)
	if err != nil {
		respondBrandingError(w, r, err)
		return
	}
	utils.WriteOK(w, BrandingResponse{
		LogoURL:      info.LogoURL,
		Tagline:      info.Tagline,
		Website:      info.Website,
		PrimaryColor: info.PrimaryColor,
	})
}

// Set — PUT /api/teams/{slug}/branding. Только owner на Max.
func (h *BrandingHandler) Set(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	req, err := utils.DecodeAndValidate[BrandingRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	if err := h.teams.SetBranding(r.Context(), slug, userID, teamuc.BrandingInput{
		LogoURL:      req.LogoURL,
		Tagline:      req.Tagline,
		Website:      req.Website,
		PrimaryColor: req.PrimaryColor,
	}); err != nil {
		respondBrandingError(w, r, err)
		return
	}

	utils.WriteOK(w, BrandingResponse(req))
}

func respondBrandingError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, teamuc.ErrNotFound):
		httperr.Respond(w, httperr.NotFound("Команда не найдена"))
	case errors.Is(err, teamuc.ErrForbidden), errors.Is(err, teamuc.ErrNotOwner):
		httperr.Respond(w, httperr.Forbidden("Нет доступа"))
	case errors.Is(err, teamuc.ErrBrandingMaxOnly):
		httperr.RespondQuotaError(w, "branding", 0, 0, "pro",
			"Брендинг публичных ссылок доступен только на тарифе Max. Обновитесь на странице /pricing.")
	case errors.Is(err, teamuc.ErrBrandingInvalidURL),
		errors.Is(err, teamuc.ErrBrandingInvalidColor),
		errors.Is(err, teamuc.ErrBrandingInvalidTagline):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	default:
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
	}
}
