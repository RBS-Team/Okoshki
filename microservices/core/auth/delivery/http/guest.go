package http

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

// GuestSession godoc
// @Summary      Создание гостевой сессии
// @Description  Выдаёт временный JWT с ролью "guest". Позволяет идентифицировать гостя. При попытке записи к мастеру возвращает 401 с кодом "registration_required".
// @Tags         auth
// @Produce      json
// @Success      200 {object} dto.GuestSessionResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /guest/session [post]
func (h *AuthHandler) GuestSession(w http.ResponseWriter, r *http.Request) {
	const op = "handler.GuestSession"
	log := middleware.LoggerFromContext(r.Context())

	// Если уже есть валидный гостевой токен — возвращаем тот же guest_id.
	if cookie, err := r.Cookie(sessionTokenCookie); err == nil {
		if claims, err := h.jwtManager.Validate(cookie.Value); err == nil && claims.Role == string(model.RoleGuest) {
			response.JSON(w, http.StatusOK, dto.GuestSessionResponse{GuestID: claims.Subject})
			return
		}
	}

	guestID := uuid.New().String()
	token, err := h.jwtManager.NewToken(guestID, string(model.RoleGuest))
	if err != nil {
		log.Errorf("[%s]: failed to generate guest token: %v", op, err)
		response.InternalErrorJSON(w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionTokenCookie,
		Value:    token,
		Expires:  time.Now().Add(h.jwtManager.GetTTL()),
		HttpOnly: true,
		Path:     "/",
	})

	response.JSON(w, http.StatusOK, dto.GuestSessionResponse{GuestID: guestID})
}
