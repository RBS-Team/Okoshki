package http

//go:generate easyjson $GOFILE

import (
	"net/http"
	"regexp"
	"time"

	easyjson "github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

const (
	sessionTokenCookie = "session_token"
)

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Создаёт нового пользователя. Роль может быть только "client" или "master".
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Данные для регистрации"
// @Success      201 {object} dto.RegisterResponse "Пользователь успешно создан"
// @Failure      400 {object} response.ErrorResponse "Неверный формат запроса, невалидные данные или недопустимая роль"
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

	if req.Role != string(model.RoleClient) {
		log.Warnf("[%s]: Invalid role attempted: %s", op, req.Role)
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

// RegisterMaster godoc
// @Summary      Регистрация мастера
// @Description  Создаёт пользователя с ролью "master" и профиль мастера за один запрос. Устанавливает httpOnly cookie с JWT токеном.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterMasterRequest true "Данные мастера"
// @Success      201 {object} dto.RegisterMasterResponse "Мастер успешно создан"
// @Failure      400 {object} response.ErrorResponse "Неверный формат запроса"
// @Failure      409 {object} response.ErrorResponse "Пользователь с таким email уже существует"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /master/register [post]
func (h *AuthHandler) RegisterMaster(w http.ResponseWriter, r *http.Request) {
	const op = "handler.RegisterMaster"
	defer r.Body.Close()

	log := middleware.LoggerFromContext(r.Context())

	var req dto.RegisterMasterRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Errorf("[%s]: Invalid request body: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if !h.validateCredentials(req.Email, req.Password) {
		response.BadRequestJSON(w)
		return
	}

	if req.Name == "" {
		log.Warnf("[%s]: missing required field: name", op)
		response.BadRequestJSON(w)
		return
	}

	user, err := h.service.RegisterNewUser(r.Context(), dto.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Role:     string(model.RoleMaster),
	})
	if err != nil {
		log.Errorf("[%s]: failed to create user: %v", op, err)
		h.handleAuthError(w, err)
		return
	}

	masterID, err := h.masterCreator.CreateMasterProfile(
		r.Context(), user.ID, req.Name, req.Bio, req.Timezone, req.Lat, req.Lon,
	)
	if err != nil {
		log.Errorf("[%s]: failed to create master profile for user %s: %v", op, user.ID, err)
		if delErr := h.service.DeleteUser(r.Context(), user.ID); delErr != nil {
			log.Errorf("[%s]: rollback failed, orphaned user %s: %v", op, user.ID, delErr)
		}
		response.InternalErrorJSON(w)
		return
	}

	token, err := h.jwtManager.NewToken(user.ID, string(model.RoleMaster))
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

	log.Infof("[%s]: Master registered: userID=%s masterID=%s", op, user.ID, masterID)
	response.JSON(w, http.StatusCreated, dto.RegisterMasterResponse{
		UserID:   user.ID,
		MasterID: masterID,
		Email:    user.Email,
		Role:     string(model.RoleMaster),
	})
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
