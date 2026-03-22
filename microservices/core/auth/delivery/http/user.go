package http

//go:generate easyjson $GOFILE

import (
	"net/http"
	"regexp"
	"time"

	"github.com/RBS-Team/Okoshki/internal/middleware"

	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"

	easyjson "github.com/mailru/easyjson"
)

const (
	sessionTokenCookie = "session_token"
)

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Создаёт нового пользователя с указанными email, паролем и ролью
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Данные для регистрации"
// @Success      201 {object} dto.RegisterResponse "Пользователь успешно создан"
// @Failure      400 {object} response.ErrorResponse "Неверный формат запроса или email/пароль не проходят валидацию"
// @Failure      409 {object} response.ErrorResponse "Пользователь с таким email уже существует"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /client/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "handler.Register"
	defer r.Body.Close()

	log := middleware.LoggerFromContext(r.Context())

	var req dto.RegisterRequest

	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Errorf("[%s]: Invalid request body: %v", op, err)
		response.BadRequestJSON(w)
		return
	}
	if !h.validateCredentials(req.Email, req.Password) {
		response.BadRequestJSON(w)
		return
	}

	user, err := h.service.RegisterNewUser(r.Context(), dto.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		log.Errorf("[%s]: Service error: %v", op, err)
		h.handleAuthError(w, err)
		return
	}

	token, err := h.jwtManager.NewToken(user.ID, user.Role)
	if err != nil {
		log.Errorf("[%s]: Failed to generate token: %v", op, err)
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

	log.Infof("[%s]: User registered successfully: %s", op, user.ID)
	response.JSON(w, http.StatusCreated, dto.RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	})
}

// Login godoc
// @Summary      Аутентификация пользователя
// @Description  Вход в систему по email и паролю. При успешном входе устанавливается httpOnly cookie с JWT токеном
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Учётные данные пользователя"
// @Success      200 {object} dto.LoginResponse "Успешный вход"
// @Failure      400 {object} response.ErrorResponse "Неверный формат запроса или email/пароль не проходят валидацию"
// @Failure      401 {object} response.ErrorResponse "Неверный email или пароль"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /client/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "handler.Login"
	defer r.Body.Close()

	log := middleware.LoggerFromContext(r.Context())

	var req dto.LoginRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Errorf("[%s]: Invalid request body: %v", op, err)
		response.BadRequestJSON(w)
		return
	}
	// Валидация
	if !h.validateCredentials(req.Email, req.Password) {
		response.BadRequestJSON(w)
		return
	}

	user, err := h.service.Login(r.Context(), dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		log.Errorf("[%s]: Service error: %v", op, err)
		h.handleAuthError(w, err)
		return
	}

	token, err := h.jwtManager.NewToken(user.ID, user.Role)
	if err != nil {
		log.Errorf("[%s]: Failed to generate token: %v", op, err)
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

	log.Infof("[%s]: User login successfully: %s", op, user.ID)
	response.JSON(w, http.StatusOK, dto.LoginResponse{
		ID:   user.ID,
		Role: user.Role,
	})
}

// Login godoc
// @Summary      Выход из аккаунта
// @Description  Выход из системы. При успешном выходе юзеру устанавливается кука с пустым jwt токеном
// @Tags         auth
// @Accept       json
// @Produce      json
// @Router       /logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	const op = "handler.Logout"
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
	log.Infof("[%s]: User logged out successfully: %s", op, userID)

	response.JSON(w, http.StatusOK, "Ok")
}

// validateCredentials - простая валидация формата
// func (h *AuthHandler) validateCredentials(email, password string) error {
// 	if email == "" || password == "" {
// 		return errors.New("email and password are required")
// 	}

// 	if !strings.Contains(email, "@") {
// 		return errors.New("invalid email format")
// 	}

// 	if len(password) < 6 {
// 		return errors.New("password must be at least 6 characters")
// 	}

//		return nil
//	}
func (h *AuthHandler) validateCredentials(email, pass string) bool {

	if email == "" || pass == "" {
		return false
	}
	if !isValidEmail(email) {
		return false
	}
	if len(pass) < 6 {
		return false
	}
	return true
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {

	if len(email) > 254 { // Максимальная длина email
		return false
	}
	return emailRegex.MatchString(email)
}
