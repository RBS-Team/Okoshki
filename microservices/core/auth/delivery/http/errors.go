package http

import (
	"errors"
	"net/http"

	"github.com/RBS-Team/Okoshki/microservices/core/auth/service"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

func (h *AuthHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.NotFoundJSON(w)
	default:
		response.InternalErrorJSON(w)
	}
}
