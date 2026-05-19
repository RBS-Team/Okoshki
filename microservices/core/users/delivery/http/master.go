package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/users/dto"
	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

// RegisterMaster godoc
// @Summary      Регистрация мастера
// @Description  Создаёт учётную запись и профиль мастера атомарно. Устанавливает httpOnly cookie с JWT.
// @Tags         registration
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterMasterRequest true "Данные мастера"
// @Success      201 {object} dto.RegisterMasterResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /master/register [post]
func (h *handler) RegisterMaster(w http.ResponseWriter, r *http.Request) {
	const op = "users.handler.RegisterMaster"
	defer r.Body.Close()

	log := middleware.LoggerFromContext(r.Context())

	var req dto.RegisterMasterRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Errorf("[%s]: invalid request body: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if !isValidCredentials(req.Email, req.Password) ||
		req.FirstName == "" || req.LastName == "" || req.Phone == "" || req.CategoryID == "" {
		response.BadRequestJSON(w)
		return
	}

	result, err := h.service.RegisterMaster(r.Context(), req)
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

	log.Infof("[%s]: master registered: userID=%s masterID=%s", op, result.UserID, result.MasterID)
	response.JSON(w, http.StatusCreated, result)
}

// GetMasterByID godoc
// @Summary      Получение мастера по ID
// @Description  Возвращает информацию о мастере по его уникальному идентификатору
// @Tags         masters
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID мастера" format(uuid)
// @Success      200 {object} dto.Master "Мастер найден"
// @Failure      400 {object} response.ErrorResponse "Неверный формат ID"
// @Failure      404 {object} response.ErrorResponse "Мастер не найден"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /masters/{id} [get]
func (h *handler) GetMasterByID(w http.ResponseWriter, r *http.Request) {
	const op = "users.handler.GetMasterByID"
	log := middleware.LoggerFromContext(r.Context())

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Errorf("[%s]: id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("[%s]: Failed to parse master ID from URL: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	master, err := h.service.GetMasterByID(r.Context(), id)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			log.Errorf("[%s]: Service error: %v", op, err)
		}
		h.handleMasterError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, master)
}

// GetMasterByUserID godoc
// @Summary      Получение мастера по userID
// @Description  Возвращает профиль мастера по UUID пользователя
// @Tags         masters
// @Accept       json
// @Produce      json
// @Param        userID path string true "UUID пользователя" format(uuid)
// @Success      200 {object} dto.Master "Мастер найден"
// @Failure      400 {object} response.ErrorResponse "Неверный формат ID"
// @Failure      404 {object} response.ErrorResponse "Мастер не найден"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /masters/user/{userID} [get]
func (h *handler) GetMasterByUserID(w http.ResponseWriter, r *http.Request) {
	const op = "users.handler.GetMasterByUserID"
	log := middleware.LoggerFromContext(r.Context())

	idStr, ok := mux.Vars(r)["userID"]
	if !ok {
		log.Errorf("[%s]: userID is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("[%s]: failed to parse userID from URL: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	master, err := h.service.GetMasterByUserID(r.Context(), userID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			log.Errorf("[%s]: service error: %v", op, err)
		}
		h.handleMasterError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, master)
}

// GetAllMasters godoc
// @Summary      Получение списка мастеров
// @Description  Возвращает список всех мастеров с пагинацией
// @Tags         masters
// @Accept       json
// @Produce      json
// @Param        limit  query    int     false "Количество записей на странице (max: 100)" default(20) minimum(1) maximum(100)
// @Param        offset query    int     false "Смещение для пагинации" default(0) minimum(0)
// @Success      200 {array}  dto.Master "Список мастеров"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /masters [get]
func (h *handler) GetAllMasters(w http.ResponseWriter, r *http.Request) {
	const op = "users.handler.GetAllMasters"
	log := middleware.LoggerFromContext(r.Context())

	limit, offset := parsePagination(r)

	masters, err := h.service.GetAllMasters(r.Context(), limit, offset)
	if err != nil {
		log.Errorf("[%s]: Service error: %v", op, err)
		h.handleMasterError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, masters)
}

// GetMastersByCategory godoc
// @Summary      Получение мастеров по категории
// @Description  Возвращает список мастеров, предоставляющих услуги в указанной категории
// @Tags         masters
// @Accept       json
// @Produce      json
// @Param        id     path    string  true  "UUID категории" format(uuid)
// @Param        limit  query   int     false "Количество записей на странице (max: 100)" default(20) minimum(1) maximum(100)
// @Param        offset query   int     false "Смещение для пагинации" default(0) minimum(0)
// @Success      200 {array}  dto.Master "Список мастеров категории"
// @Failure      400 {object} response.ErrorResponse "Неверный формат ID категории"
// @Failure      404 {object} response.ErrorResponse "Категория не найдена"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /categories/{id}/masters [get]
func (h *handler) GetMastersByCategory(w http.ResponseWriter, r *http.Request) {
	const op = "users.handler.GetMastersByCategory"
	log := middleware.LoggerFromContext(r.Context())

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Errorf("[%s]: category id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("[%s]: Failed to parse category ID from URL: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	limit, offset := parsePagination(r)

	masters, err := h.service.GetMastersByCategory(r.Context(), categoryID, limit, offset)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			log.Errorf("[%s]: Service error: %v", op, err)
		}
		h.handleMasterError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, masters)
}
