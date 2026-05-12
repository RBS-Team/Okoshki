package http

import (
	"errors"
	"net/http"

	"github.com/RBS-Team/Okoshki/microservices/core/users/service"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

func (h *Handler) handleUsersError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.NotFoundJSON(w)
	case errors.Is(err, service.ErrConflict):
		response.ConflictJSON(w)
	case errors.Is(err, service.ErrForbidden):
		response.ForbiddenJSON(w)
	case errors.Is(err, service.ErrInvalidInput),
		errors.Is(err, service.ErrInvalidTimezone):
		response.BadRequestJSON(w)
	default:
		response.InternalErrorJSON(w)
	}
}

func (h *Handler) handleMasterError(w http.ResponseWriter, err error) {
	h.handleUsersError(w, err)
}
