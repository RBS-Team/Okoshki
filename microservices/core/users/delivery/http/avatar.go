package http

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

const avatarMaxFileSize = 5 << 20 // 5 MB

type avatarResponse struct {
	AvatarURL string `json:"avatar_url"`
}

// UploadMasterAvatar godoc
// @Summary      Загрузить аватар мастера
// @Description  Загружает изображение (JPEG/PNG/WebP, до 5 МБ) как аватар мастера. Доступно только владельцу профиля.
// @Tags         masters
// @Accept       multipart/form-data
// @Produce      json
// @Param        masterID path     string true "UUID мастера" format(uuid)
// @Param        file     formData file   true "Файл изображения"
// @Success      200 {object} avatarResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Security     CookieAuth
// @Router       /masters/{masterID}/avatar [put]
func (h *handler) UploadMasterAvatar(w http.ResponseWriter, r *http.Request) {
	const op = "users.handler.UploadMasterAvatar"
	log := middleware.LoggerFromContext(r.Context())

	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok || userIDStr == "" {
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

	fileHeaders := r.MultipartForm.File["file"]
	if len(fileHeaders) == 0 {
		log.Warnf("[%s]: no file provided", op)
		response.BadRequestJSON(w)
		return
	}

	header := fileHeaders[0]
	if header.Size > avatarMaxFileSize {
		log.Warnf("[%s]: file exceeds size limit", op)
		response.BadRequestJSON(w)
		return
	}

	ct := header.Header.Get("Content-Type")
	if !allowedMimeTypes[ct] {
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

	url, err := h.service.UploadMasterAvatar(r.Context(), userIDStr, masterIDStr, fh, header.Size, ct)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			log.Errorf("[%s]: service error: %v", op, err)
		}
		h.handleUsersError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, avatarResponse{AvatarURL: url})
}

// UploadClientAvatar godoc
// @Summary      Загрузить аватар клиента
// @Description  Загружает изображение (JPEG/PNG/WebP, до 5 МБ) как аватар текущего клиента.
// @Tags         clients
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Файл изображения"
// @Success      200 {object} avatarResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Security     CookieAuth
// @Router       /clients/me/avatar [put]
func (h *handler) UploadClientAvatar(w http.ResponseWriter, r *http.Request) {
	const op = "users.handler.UploadClientAvatar"
	log := middleware.LoggerFromContext(r.Context())

	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok || userIDStr == "" {
		response.UnauthorizedJSON(w)
		return
	}

	if err := r.ParseMultipartForm(maxUploadMemory); err != nil {
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
	if header.Size > avatarMaxFileSize {
		log.Warnf("[%s]: file exceeds size limit", op)
		response.BadRequestJSON(w)
		return
	}

	ct := header.Header.Get("Content-Type")
	if !allowedMimeTypes[ct] {
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

	url, err := h.service.UploadClientAvatar(r.Context(), userIDStr, fh, header.Size, ct)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			log.Errorf("[%s]: service error: %v", op, err)
		}
		h.handleUsersError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, avatarResponse{AvatarURL: url})
}
