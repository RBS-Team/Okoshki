package http

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

const (
	categoryMaxUploadMemory = 32 << 20 // 32 MB
	categoryMaxFileSize     = 5 << 20  // 5 MB
)

var categoryAllowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// UploadCategoryAvatar godoc
// @Summary      Загрузить аватар категории
// @Description  Загружает изображение (JPEG/PNG/WebP, до 5 МБ) и сохраняет как аватар категории. Доступно только администраторам.
// @Tags         categories
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path     string true "UUID категории" format(uuid)
// @Param        file formData file   true "Файл изображения"
// @Success      204
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Security     CookieAuth
// @Router       /categories/{id}/avatar [put]
func (h *Handler) UploadCategoryAvatar(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.UploadCategoryAvatar"
	log := middleware.LoggerFromContext(r.Context())

	categoryIDStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Errorf("[%s]: id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	if err := r.ParseMultipartForm(categoryMaxUploadMemory); err != nil {
		log.Warnf("[%s]: parse multipart form: %v", op, err)
		response.BadRequestJSON(w)
		return
	}
	defer r.MultipartForm.RemoveAll()

	fileHeaders := r.MultipartForm.File["file"]
	if len(fileHeaders) == 0 {
		log.Warnf("[%s]: no file provided", op)
		response.BadRequestJSON(w)
		return
	}

	header := fileHeaders[0]
	if header.Size > categoryMaxFileSize {
		log.Warnf("[%s]: file exceeds size limit", op)
		response.BadRequestJSON(w)
		return
	}

	ct := header.Header.Get("Content-Type")
	if !categoryAllowedMimeTypes[ct] {
		log.Warnf("[%s]: unsupported content type: %s", op, ct)
		response.BadRequestJSON(w)
		return
	}

	fh, err := header.Open()
	if err != nil {
		log.Errorf("[%s]: open file: %v", op, err)
		response.InternalErrorJSON(w)
		return
	}
	defer fh.Close()

	if err := h.service.UploadCategoryAvatar(r.Context(), categoryIDStr, fh, header.Size, ct); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			log.Errorf("[%s]: service error: %v", op, err)
		}
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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
		if !errors.Is(err, domain.ErrNotFound) {
			log.Errorf("[%s]: Service error: %v", op, err)
		}
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, category)
}
