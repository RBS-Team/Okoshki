package http

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/service"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

// GetAllCategories godoc
// @Summary      Получить все категории
// @Description  Возвращает список всех категорий услуг
// @Tags         categories
// @Accept       json
// @Produce      json
// @Success      200 {array} dto.Category "Список категорий"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /categories [get]
func (h *Handler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetAllCategories"
	log := middleware.LoggerFromContext(r.Context())

	categories, err := h.service.GetAllCategories(r.Context())
	if err != nil {
		log.Errorf("[%s]: Service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, categories)
}

// GetCategoryByID godoc
// @Summary      Получить категорию по ID
// @Description  Возвращает информацию о категории по её уникальному идентификатору
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID категории" format(uuid)
// @Success      200 {object} dto.Category "Категория найдена"
// @Failure      400 {object} response.ErrorResponse "Неверный формат ID"
// @Failure      404 {object} response.ErrorResponse "Категория не найдена"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /categories/{id} [get]
func (h *Handler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetCategoryByID"
	log := middleware.LoggerFromContext(r.Context())

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Errorf("[%s]: id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("[%s]: Failed to parse category ID from URL: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	category, err := h.service.GetCategoryByID(r.Context(), id)
	if err != nil {
		if !errors.Is(err, service.ErrNotFound) {
			log.Errorf("[%s]: Service error: %v", op, err)
		}
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, category)
}
