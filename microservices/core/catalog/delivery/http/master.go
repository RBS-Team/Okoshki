package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/service"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

// CreateMaster godoc
// @Summary      Создание мастера
// @Description  Создаёт нового мастера в каталоге
// @Tags         masters
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateMasterRequest true "Данные мастера"
// @Success      201 {object} dto.Master "Мастер успешно создан"
// @Failure      400 {object} response.ErrorResponse "Неверный формат запроса или отсутствуют обязательные поля"
// @Failure      409 {object} response.ErrorResponse "Мастер с таким user_id уже существует"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /masters [post]
func (h *Handler) CreateMaster(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.CreateMaster"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	var req dto.CreateMasterRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: failed to unmarshal request: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if req.UserID == "" || req.Name == "" {
		log.Warnf("[%s]: missing required fields: user_id or name", op)
		response.BadRequestJSON(w)
		return
	}

	master, err := h.service.CreateMaster(r.Context(), req)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleMasterError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, master)
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
func (h *Handler) GetMasterByID(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetMasterByID"
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
		if !errors.Is(err, service.ErrNotFound) {
			log.Errorf("[%s]: Service error: %v", op, err)
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
func (h *Handler) GetAllMasters(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetAllMasters"
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
func (h *Handler) GetMastersByCategory(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetMastersByCategory"
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
		if !errors.Is(err, service.ErrNotFound) {
			log.Errorf("[%s]: Service error: %v", op, err)
		}
		h.handleMasterError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, masters)
}

func parsePagination(r *http.Request) (uint64, uint64) {
	query := r.URL.Query()

	limit, err := strconv.ParseUint(query.Get("limit"), 10, 64)
	if err != nil || limit == 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.ParseUint(query.Get("offset"), 10, 64)
	if err != nil {
		offset = 0
	}

	return limit, offset
}
