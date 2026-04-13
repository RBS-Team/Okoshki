package http

import (
	"errors"
	"net/http"

	"github.com/RBS-Team/Okoshki/microservices/core/auth/service"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.NotFoundJSON(w)
	case errors.Is(err, service.ErrValidation):
		response.UnauthorizedJSON(w)
	case errors.Is(err, service.ErrConflict):
		response.ConflictJSON(w)
	default:
		response.InternalErrorJSON(w)
	}
}
