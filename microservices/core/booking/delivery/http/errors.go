package http

import (
	"errors"
	"net/http"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.NotFoundJSON(w)
	case errors.Is(err, domain.ErrConflict),
		errors.Is(err, domain.ErrSlotNotAvailable),
		errors.Is(err, domain.ErrTimeConflict):
		response.ConflictJSON(w)
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrLeadTimeViolation),
		errors.Is(err, domain.ErrInvalidTimezone):
		response.BadRequestJSON(w)
	case errors.Is(err, domain.ErrForbidden):
		response.ForbiddenJSON(w)
	case errors.Is(err, domain.ErrUnauthorized):
		response.UnauthorizedJSON(w)
	default:
		response.InternalErrorJSON(w)
	}
}
