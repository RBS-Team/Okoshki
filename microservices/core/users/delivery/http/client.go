package http

import (
	"net/http"
	"regexp"
	"time"

	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/users/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

const sessionTokenCookie = "session_token"

// RegisterClient godoc
// @Summary      Регистрация клиента
// @Description  Создаёт учётную запись и профиль клиента атомарно. Устанавливает httpOnly cookie с JWT.
// @Tags         registration
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterClientRequest true "Данные клиента"
// @Success      201 {object} dto.RegisterClientResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /client/register [post]
func (h *Handler) RegisterClient(w http.ResponseWriter, r *http.Request) {
	const op = "users.handler.RegisterClient"
	defer r.Body.Close()

	log := middleware.LoggerFromContext(r.Context())

	var req dto.RegisterClientRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Errorf("[%s]: invalid request body: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if !isValidCredentials(req.Email, req.Password) || req.FirstName == "" || req.LastName == "" {
		response.BadRequestJSON(w)
		return
	}

	result, err := h.service.RegisterClient(r.Context(), req)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleUsersError(w, err)
		return
	}

	token, err := h.jwtManager.NewToken(result.UserID, result.Role)
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

	log.Infof("[%s]: client registered: userID=%s", op, result.UserID)
	response.JSON(w, http.StatusCreated, result)
}

func isValidCredentials(email, pass string) bool {
	return email != "" && pass != "" && len(pass) >= 6 && isValidEmail(email)
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return len(email) <= 254 && emailRegex.MatchString(email)
}
