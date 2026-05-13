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
	default:
		response.InternalErrorJSON(w)
	}
}

func (h *Handler) handleServiceItemError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.NotFoundJSON(w)
	case errors.Is(err, domain.ErrConflict):
		response.ConflictJSON(w)
	default:
		response.InternalErrorJSON(w)
	}
}

func (h *Handler) handleScheduleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.NotFoundJSON(w)
	case errors.Is(err, domain.ErrConflict):
		response.ConflictJSON(w)
	default:
		response.BadRequestJSON(w)
	}
}
