package starter

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	httperr "promptvault/internal/delivery/http/errors"
	"promptvault/internal/delivery/http/utils"
	authmw "promptvault/internal/middleware/auth"
	starteruc "promptvault/internal/usecases/starter"
)

type Handler struct {
	svc      *starteruc.Service
	validate *validator.Validate
}

func NewHandler(svc *starteruc.Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// GET /api/starter/catalog — возвращает встроенный каталог. Доступен любому
// залогиненному, в том числе уже прошедшему wizard. Никакой БД-нагрузки.
func (h *Handler) Catalog(w http.ResponseWriter, _ *http.Request) {
	c := h.svc.ListCatalog()
	utils.WriteOK(w, NewCatalogResponse(c))
}

// POST /api/starter/complete — установка выбранных промптов + маркировка
// юзера как прошедшего онбординг. Атомарно. Идемпотентно через 409.
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	userID := authmw.GetUserID(r.Context())

	req, err := utils.DecodeAndValidate[CompleteRequest](r, h.validate)
	if err != nil {
		httperr.Respond(w, httperr.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.Install(r.Context(), userID, req.Install)
	if err != nil {
		respondError(w, err)
		return
	}

	utils.WriteOK(w, NewCompleteResponse(result))
}
