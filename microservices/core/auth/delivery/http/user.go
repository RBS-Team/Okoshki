package http

import (
	"net/http"
	"regexp"
	"time"

	easyjson "github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

const sessionTokenCookie = "session_token"

// Login godoc
// @Summary      Аутентификация пользователя
// @Description  Вход по email и паролю. При успехе устанавливается httpOnly cookie с JWT
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Учётные данные"
// @Success      200 {object} dto.LoginResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "auth.handler.Login"
	defer r.Body.Close()

	log := middleware.LoggerFromContext(r.Context())

	var req dto.LoginRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Errorf("[%s]: invalid request body: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if !isValidCredentials(req.Email, req.Password) {
		response.BadRequestJSON(w)
		return
	}

	user, err := h.service.Login(r.Context(), req)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleAuthError(w, err)
		return
	}

	token, err := h.jwtManager.NewToken(user.ID, user.Role)
	if err != nil {
		log.Errorf("[%s]: failed to generate token: %v", op, err)
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

	log.Infof("[%s]: user logged in: %s", op, user.ID)
	response.JSON(w, http.StatusOK, user)
}

// Logout godoc
// @Summary      Выход из аккаунта
// @Tags         auth
// @Produce      json
// @Router       /logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	const op = "auth.handler.Logout"
	defer r.Body.Close()

	userID, _ := middleware.GetUserID(r.Context())
	log := middleware.LoggerFromContext(r.Context())

	http.SetCookie(w, &http.Cookie{
		Name:     sessionTokenCookie,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	log.Infof("[%s]: user logged out: %s", op, userID)
	response.JSON(w, http.StatusOK, "Ok")
}

func isValidCredentials(email, pass string) bool {
	return email != "" && pass != "" && len(pass) >= 6 && isValidEmail(email)
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return len(email) <= 254 && emailRegex.MatchString(email)
}
