package http

import (
	"errors"
	"net/http"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

func (h *handler) handleUsersError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.NotFoundJSON(w)
	case errors.Is(err, domain.ErrConflict):
		response.ConflictJSON(w)
	case errors.Is(err, domain.ErrForbidden):
		response.ForbiddenJSON(w)
	case errors.Is(err, domain.ErrUnauthorized):
		response.UnauthorizedJSON(w)
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrInvalidTimezone):
		response.BadRequestJSON(w)
	default:
		response.InternalErrorJSON(w)
	}
}

func (h *handler) handleMasterError(w http.ResponseWriter, err error) {
	h.handleUsersError(w, err)
}
