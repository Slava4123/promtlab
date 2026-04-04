package search

import (
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
)

func respondError(w http.ResponseWriter, err error) {
	httperr.Respond(w, httperr.Internal(err))
}
