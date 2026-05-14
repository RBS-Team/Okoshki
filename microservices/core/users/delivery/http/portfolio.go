package http

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/users/dto"
	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

const (
	maxUploadMemory = 32 << 20 // 32 MB суммарно в памяти
	maxFileSize     = 10 << 20 // 10 MB на файл
	maxFilesCount   = 10
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// UploadPortfolioPhotos godoc
// @Summary      Загрузка фотографий портфолио мастера
// @Description  Загружает до 10 фотографий (JPEG/PNG/WebP, до 10 МБ каждая) в портфолио мастера. Доступно только владельцу профиля.
// @Tags         portfolio
// @Accept       multipart/form-data
// @Produce      json
// @Param        masterID path     string true  "UUID мастера" format(uuid)
// @Param        files    formData []file   true  "Файлы изображений (поле files[])"
// @Success      201      {array}  dto.PortfolioPhoto
// @Failure      400      {object} response.ErrorResponse
// @Failure      401      {object} response.ErrorResponse
// @Failure      403      {object} response.ErrorResponse
// @Failure      500      {object} response.ErrorResponse
// @Security     CookieAuth
// @Router       /masters/{masterID}/portfolio [post]
func (h *Handler) UploadPortfolioPhotos(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.UploadPortfolioPhotos"
	log := middleware.LoggerFromContext(r.Context())

	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok || userIDStr == "" {
		log.Errorf("[%s]: missing user id in context", op)
		response.UnauthorizedJSON(w)
		return
	}

	masterIDStr, ok := mux.Vars(r)["masterID"]
	if !ok {
		log.Errorf("[%s]: masterID missing in URL", op)
		response.BadRequestJSON(w)
		return
	}

	if err := r.ParseMultipartForm(maxUploadMemory); err != nil {
		log.Warnf("[%s]: parse multipart form: %v", op, err)
		response.BadRequestJSON(w)
		return
	}
	defer r.MultipartForm.RemoveAll()

	fileHeaders := r.MultipartForm.File["files"]
	if len(fileHeaders) == 0 {
		log.Warnf("[%s]: no files provided", op)
		response.BadRequestJSON(w)
		return
	}
	if len(fileHeaders) > maxFilesCount {
		log.Warnf("[%s]: too many files: %d", op, len(fileHeaders))
		response.BadRequestJSON(w)
		return
	}

	uploads := make([]dto.FileUpload, 0, len(fileHeaders))
	closers := make([]func(), 0, len(fileHeaders))
	defer func() {
		for _, close := range closers {
			close()
		}
	}()

	for _, fh := range fileHeaders {
		if fh.Size > maxFileSize {
			log.Warnf("[%s]: file %s exceeds size limit", op, fh.Filename)
			response.BadRequestJSON(w)
			return
		}

		ct := fh.Header.Get("Content-Type")
		if !allowedMimeTypes[ct] {
			log.Warnf("[%s]: unsupported content type: %s", op, ct)
			response.BadRequestJSON(w)
			return
		}

		f, err := fh.Open()
		if err != nil {
			log.Errorf("[%s]: open file %s: %v", op, fh.Filename, err)
			response.InternalErrorJSON(w)
			return
		}
		closers = append(closers, func() { f.Close() })

		uploads = append(uploads, dto.FileUpload{
			Reader:      f,
			Size:        fh.Size,
			ContentType: ct,
			Name:        fh.Filename,
		})
	}

	photos, err := h.service.UploadPortfolioPhotos(r.Context(), userIDStr, masterIDStr, uploads)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handlePortfolioError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, photos)
}

// GetPortfolioPhotos godoc
// @Summary      Получение портфолио мастера
// @Description  Возвращает список фотографий портфолио мастера с presigned URL (действительны 1 час)
// @Tags         portfolio
// @Produce      json
// @Param        masterID path     string true "UUID мастера" format(uuid)
// @Success      200      {array}  dto.PortfolioPhoto
// @Failure      400      {object} response.ErrorResponse
// @Failure      404      {object} response.ErrorResponse
// @Failure      500      {object} response.ErrorResponse
// @Router       /masters/{masterID}/portfolio [get]
func (h *Handler) GetPortfolioPhotos(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetPortfolioPhotos"
	log := middleware.LoggerFromContext(r.Context())

	masterIDStr, ok := mux.Vars(r)["masterID"]
	if !ok {
		log.Errorf("[%s]: masterID missing in URL", op)
		response.BadRequestJSON(w)
		return
	}

	photos, err := h.service.GetPortfolioPhotos(r.Context(), masterIDStr)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handlePortfolioError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, photos)
}

// DeletePortfolioPhoto godoc
// @Summary      Удаление фотографии портфолио
// @Description  Удаляет фотографию из портфолио мастера. Доступно только владельцу профиля.
// @Tags         portfolio
// @Produce      json
// @Param        masterID path     string true "UUID мастера" format(uuid)
// @Param        photoID  path     string true "UUID фотографии" format(uuid)
// @Success      204
// @Failure      400      {object} response.ErrorResponse
// @Failure      401      {object} response.ErrorResponse
// @Failure      403      {object} response.ErrorResponse
// @Failure      404      {object} response.ErrorResponse
// @Failure      500      {object} response.ErrorResponse
// @Security     CookieAuth
// @Router       /masters/{masterID}/portfolio/{photoID} [delete]
func (h *Handler) DeletePortfolioPhoto(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.DeletePortfolioPhoto"
	log := middleware.LoggerFromContext(r.Context())

	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok || userIDStr == "" {
		log.Errorf("[%s]: missing user id in context", op)
		response.UnauthorizedJSON(w)
		return
	}

	vars := mux.Vars(r)
	masterIDStr := vars["masterID"]
	photoIDStr := vars["photoID"]

	if masterIDStr == "" || photoIDStr == "" {
		log.Errorf("[%s]: masterID or photoID missing in URL", op)
		response.BadRequestJSON(w)
		return
	}

	if err := h.service.DeletePortfolioPhoto(r.Context(), userIDStr, masterIDStr, photoIDStr); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handlePortfolioError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handlePortfolioError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		response.ForbiddenJSON(w)
	case errors.Is(err, domain.ErrNotFound):
		response.NotFoundJSON(w)
	case errors.Is(err, domain.ErrInvalidInput):
		response.BadRequestJSON(w)
	default:
		response.InternalErrorJSON(w)
	}
}
