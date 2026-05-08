package http

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/service"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

// CreateServiceItem godoc
// @Summary      Создание услуги мастера
// @Description  Добавляет новую услугу для указанного мастера
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        id      path      string                       true  "UUID мастера" format(uuid)
// @Param        request body      dto.CreateServiceItemRequest true  "Данные услуги"
// @Success      201     {object}  dto.ServiceItem    			"Услуга успешно создана"
// @Failure      400     {object}  response.ErrorResponse       "Неверный формат ID или тела запроса"
// @Failure      404     {object}  response.ErrorResponse       "Мастер не найден"
// @Failure      409     {object}  response.ErrorResponse       "Конфликт (например, услуга уже существует)"
// @Failure      500     {object}  response.ErrorResponse       "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /masters/{id}/services [post]
func (h *Handler) CreateServiceItem(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.CreateServiceItem"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Errorf("[%s]: master id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	masterID, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("[%s]: invalid master id format: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	var req dto.CreateServiceItemRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: failed to unmarshal request: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if req.Title == "" || req.CategoryID == "" || req.DurationMinutes <= 0 || req.Price < 0 || req.Address == ""{
		log.Warnf("[%s]: invalid request fields", op)
		response.BadRequestJSON(w)
		return
	}

	item, err := h.service.CreateServiceItem(r.Context(), masterID, req)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleServiceItemError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, item)
}
// func (h *Handler) UploadServiceItemImages(w http.ResponseWriter, r *http.Request) {
// 	const op = "catalog.handler.UploadServiceItemImages"
// 	log:=middleware.LoggerFromContext(r.Context())

// 	// Проверка авторизации (убеждаемся, что это мастер)
// 	userRole, ok := middleware.GetUserRole(r.Context())
// 	if !ok || userRole != "Master" {
// 		response.UnauthorizedJSON(w)
// 		return
// 	}

// 	// Получаем ID услуги из URL
// 	serviceIDRaw := mux.Vars(r)["srvсid"]
// 	serviceID, err := uuid.Parse(serviceIDRaw)
// 	if err != nil {
// 		log.Errorf("[%s]: invalid service id: %v", op, err)
// 		response.BadRequestJSON(w)
// 		return
// 	}

// 	file, header, err := r.FormFile("avatar")
// 	if err != nil {
// 		log.Errorf("[%s]: failed to get file: %v", op, err)
// 		response.BadRequestJSON(w)
// 		return
// 	}
// 	defer file.Close()

// 	if err := h.validatePlaylistAvatar(header.Header.Get("Content-Type"), header.Size); err != nil {
// 		log.Errorf("[%s] validation error: %v", op, err)
// 		response.BadRequestJSON(w)
// 		return
// 	}

// 	res, err := h.service.UploadPlaylistAvatar(r.Context(), dto.UploadServiceItemImagesRequest{
// 		PlaylistID:  playlistID,
// 		File:        file,
// 		Size:        header.Size,
// 		ContentType: header.Header.Get("Content-Type"),
// 	})
// 	if err != nil {
// 		log.Errorf("[%s]: service error: %v", op, err)
// 		h.handleError(w, err)
// 		return
// 	}

// 	response.JSON(w, http.StatusOK, res)
// }

// func (h *Handler) DeletePlaylistAvatar(w http.ResponseWriter, r *http.Request) {
// 	const op = "handler.DeletePlaylistAvatar"
// 	log := middleware.LoggerFromContext(r.Context())

// 	userID, ok := middleware.GetUserID(r.Context())
// 	if !ok || userID == "" {
// 		response.UnauthorizedJSON(w)
// 		return
// 	}

// 	rawID := mux.Vars(r)["id"]
// 	playlistID, err := uuid.Parse(rawID)
// 	if err != nil {
// 		log.Errorf("[%s] invalid playlist id: %v", op, err)
// 		response.BadRequestJSON(w)
// 		return
// 	}

// 	req := dto.DeletePlaylistAvatarRequest{
// 		PlaylistID: playlistID,
// 	}
// 	if err := h.service.DeletePlaylistAvatar(r.Context(), req); err != nil {
// 		log.Errorf("[%s]: service error: %v", op, err)
// 		h.handleError(w, err)
// 		return
// 	}

// 	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
// }

// GetServiceItemsByMasterID godoc
// @Summary      Получение услуг мастера
// @Description  Возвращает список всех услуг, предоставляемых мастером
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "UUID мастера" format(uuid)
// @Success      200  {array}   dto.ServiceItem "Список услуг"
// @Failure      400  {object}  response.ErrorResponse  "Неверный формат ID"
// @Failure      404  {object}  response.ErrorResponse  "Мастер не найден"
// @Failure      500  {object}  response.ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /masters/{id}/services [get]
func (h *Handler) GetServiceItemsByMasterID(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetServiceItemsByMasterID"
	log := middleware.LoggerFromContext(r.Context())

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Errorf("[%s]: master id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	masterID, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("[%s]: invalid master id format: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	items, err := h.service.GetServiceItemsByMasterID(r.Context(), masterID)
	if err != nil {
		if !errors.Is(err, service.ErrNotFound) {
			log.Errorf("[%s]: service error: %v", op, err)
		}
		h.handleServiceItemError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// GetServicesByCategory godoc
// @Summary      Получение услуг по категории
// @Description  Возвращает список услуг, отфильтрованных по категории с пагинацией
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        id      path    string  true  "UUID категории" format(uuid)
// @Param        limit   query   int     false "Количество записей на странице (max: 100)" default(20) minimum(1) maximum(100)
// @Param        offset  query   int     false "Смещение для пагинации" default(0) minimum(0)
// @Success      200     {array} dto.ServiceWithMaster "Список услуг"
// @Failure      400     {object} response.ErrorResponse "Неверный формат ID категории"
// @Failure      404     {object} response.ErrorResponse "Категория не найдена"
// @Failure      500     {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /categories/{id}/services [get]
func (h *Handler) GetServicesByCategory(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetServicesByCategory"

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		response.BadRequestJSON(w)
		return
	}

	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequestJSON(w)
		return
	}

	limit, offset := parsePagination(r)

	items, err := h.service.GetServicesByCategory(r.Context(), categoryID, limit, offset)
	if err != nil {
		if !errors.Is(err, service.ErrNotFound) {
		}
		h.handleServiceItemError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, items)
}
