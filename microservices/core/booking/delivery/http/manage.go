package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

const dateFormat = "2006-01-02"

func parsePagination(r *http.Request) (uint64, uint64) {
	limit, err := strconv.ParseUint(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit == 0 {
		limit = 20
	}
	offset, err := strconv.ParseUint(r.URL.Query().Get("offset"), 10, 64)
	if err != nil {
		offset = 0
	}
	return limit, offset
}

func (h *Handler) getMasterID(r *http.Request) (uuid.UUID, error) {
	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}
	userID, _ := uuid.Parse(userIDStr)
	return h.service.GetMasterIDByUserID(r.Context(), userID)
}

func (h *Handler) getClientID(r *http.Request) (uuid.UUID, error) {
	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}
	return uuid.Parse(userIDStr)
}

// GetMyAppointments godoc
// @Summary      Получение списка записей клиента
// @Description  Возвращает список всех записей авторизованного клиента с пагинацией. Записи отсортированы по дате создания (сначала новые). Требуется роль client.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Param        limit  query int false "Количество записей на страницу (по умолчанию 20, максимум 100)"
// @Param        offset query int false "Смещение для пагинации (по умолчанию 0)"
// @Success      200 {array} dto.ClientAppointmentView "Список записей клиента"
// @Failure      400 {object} response.ErrorResponse "Неверные параметры пагинации"
// @Failure      401 {object} response.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} response.ErrorResponse "Доступ запрещен (требуется роль client)"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /appointments/my [get]
func (h *Handler) GetMyAppointments(w http.ResponseWriter, r *http.Request) {
	const op = "booking.handler.GetMyAppointments"
	log := middleware.LoggerFromContext(r.Context())

	clientID, err := h.getClientID(r)
	if err != nil {
		response.UnauthorizedJSON(w)
		return
	}

	limit, offset := parsePagination(r)

	appts, err := h.service.GetClientAppointments(r.Context(), clientID, limit, offset)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		response.InternalErrorJSON(w)
		return
	}

	response.JSON(w, http.StatusOK, appts)
}

// CancelAppointment godoc
// @Summary      Отмена записи клиентом
// @Description  Отменяет существующую запись от имени авторизованного клиента. Клиент может отменить только свои записи. Статус записи меняется на "cancelled". Требуется роль client.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID записи"
// @Success      200 {object} map[string]string "Запись успешно отменена"
// @Failure      400 {object} response.ErrorResponse "Неверный формат UUID записи или запись не может быть отменена"
// @Failure      401 {object} response.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} response.ErrorResponse "Доступ запрещен (требуется роль client) или попытка отменить чужую запись"
// @Failure      404 {object} response.ErrorResponse "Запись не найдена"
// @Failure      409 {object} response.ErrorResponse "Запись уже отменена или завершена"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /appointments/{id}/cancel [patch]
func (h *Handler) CancelAppointment(w http.ResponseWriter, r *http.Request) {
	const op = "booking.handler.CancelAppointment"
	log := middleware.LoggerFromContext(r.Context())

	clientID, err := h.getClientID(r)
	if err != nil {
		response.UnauthorizedJSON(w)
		return
	}

	apptID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		response.BadRequestJSON(w)
		return
	}

	req := dto.UpdateAppointmentStatusRequest{Status: "cancelled"}

	if err := h.service.UpdateAppointmentStatus(r.Context(), clientID, apptID, req, true); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// GetMasterAppointments godoc
// @Summary      Получение записей мастера за период
// @Description  Возвращает список всех записей мастера в заданном диапазоне дат. Включает записи клиентов и ручные блокировки. Требуется роль master и наличие созданного профиля мастера.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Param        start_date query string true "Начальная дата периода в формате YYYY-MM-DD (например 2026-04-21)"
// @Param        end_date   query string true "Конечная дата периода в формате YYYY-MM-DD (например 2026-04-28)"
// @Success      200 {array} dto.MasterAppointmentView "Список записей мастера"
// @Failure      400 {object} response.ErrorResponse "Отсутствуют обязательные query параметры или неверный формат даты"
// @Failure      401 {object} response.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} response.ErrorResponse "Профиль мастера не создан или недостаточно прав"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /master-appointments [get]
func (h *Handler) GetMasterAppointments(w http.ResponseWriter, r *http.Request) {
	const op = "booking.handler.GetMasterAppointments"
	log := middleware.LoggerFromContext(r.Context())

	masterID, err := h.getMasterID(r)
	if err != nil {
		response.UnauthorizedJSON(w)
		return
	}

	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")

	if startStr == "" || endStr == "" {
		response.BadRequestJSON(w)
		return
	}

	start, err := time.Parse(dateFormat, startStr)
	if err != nil {
		response.BadRequestJSON(w)
		return
	}
	end, err := time.Parse(dateFormat, endStr)
	if err != nil {
		response.BadRequestJSON(w)
		return
	}
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())

	appts, err := h.service.GetMasterAppointments(r.Context(), masterID, start, end)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		response.InternalErrorJSON(w)
		return
	}

	response.JSON(w, http.StatusOK, appts)
}

// UpdateAppointmentStatus godoc
// @Summary      Обновление статуса записи мастером
// @Description  Позволяет мастеру изменить статус записи (например, подтвердить, отметить выполненной или отменить). Мастер может обновлять только свои записи. Требуется роль master и наличие созданного профиля мастера.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID записи"
// @Param        request body dto.UpdateAppointmentStatusRequest true "Новый статус записи"
// @Success      200 {object} map[string]string "Статус успешно обновлен"
// @Failure      400 {object} response.ErrorResponse "Неверный формат UUID, отсутствует тело запроса или невалидный статус"
// @Failure      401 {object} response.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} response.ErrorResponse "Профиль мастера не создан, недостаточно прав или попытка изменить чужую запись"
// @Failure      404 {object} response.ErrorResponse "Запись не найдена"
// @Failure      409 {object} response.ErrorResponse "Недопустимый переход статуса (например, нельзя подтвердить уже отмененную запись)"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /appointments/{id}/status [patch]
func (h *Handler) UpdateAppointmentStatus(w http.ResponseWriter, r *http.Request) {
	const op = "booking.handler.UpdateAppointmentStatus"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, err := h.getMasterID(r)
	if err != nil {
		response.UnauthorizedJSON(w)
		return
	}

	apptID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		response.BadRequestJSON(w)
		return
	}

	var req dto.UpdateAppointmentStatusRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		response.BadRequestJSON(w)
		return
	}

	// isClient = false
	if err := h.service.UpdateAppointmentStatus(r.Context(), masterID, apptID, req, false); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// CreateManualBlock godoc
// @Summary      Создание ручной блокировки времени
// @Description  Позволяет мастеру вручную заблокировать временной слот (например, для обеда, перерыва или личных дел). Заблокированное время становится недоступным для записи клиентам. Требуется роль master и наличие созданного профиля мастера.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateManualBlockRequest true "Данные для создания блокировки"
// @Success      201 {object} dto.CreateManualBlockResponse "Блокировка успешно создана"
// @Failure      400 {object} response.ErrorResponse "Неверный формат запроса, невалидные данные или конфликт по времени"
// @Failure      401 {object} response.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} response.ErrorResponse "Профиль мастера не создан или недостаточно прав"
// @Failure      409 {object} response.ErrorResponse "Временной слот уже занят или пересекается с существующей записью"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /appointments/block [post]
func (h *Handler) CreateManualBlock(w http.ResponseWriter, r *http.Request) {
	const op = "booking.handler.CreateManualBlock"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, err := h.getMasterID(r)
	if err != nil {
		response.UnauthorizedJSON(w)
		return
	}

	var req dto.CreateManualBlockRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		response.BadRequestJSON(w)
		return
	}

	resp, err := h.service.CreateManualBlock(r.Context(), masterID, req)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

// DeleteManualBlock godoc
// @Summary      Удаление ручной блокировки времени
// @Description  Удаляет ранее созданную мастером ручную блокировку времени. Мастер может удалять только свои блокировки. Требуется роль master и наличие созданного профиля мастера.
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID блокировки"
// @Success      200 {object} map[string]string "Блокировка успешно удалена"
// @Failure      400 {object} response.ErrorResponse "Неверный формат UUID блокировки"
// @Failure      401 {object} response.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} response.ErrorResponse "Профиль мастера не создан или попытка удалить чужую блокировку"
// @Failure      404 {object} response.ErrorResponse "Блокировка не найдена"
// @Failure      409 {object} response.ErrorResponse "Невозможно удалить блокировку (например, она уже в прошлом)"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /appointments/block/{id} [delete]
func (h *Handler) DeleteManualBlock(w http.ResponseWriter, r *http.Request) {
	const op = "booking.handler.DeleteManualBlock"
	log := middleware.LoggerFromContext(r.Context())

	masterID, err := h.getMasterID(r)
	if err != nil {
		response.UnauthorizedJSON(w)
		return
	}

	blockID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		response.BadRequestJSON(w)
		return
	}

	if err := h.service.DeleteManualBlock(r.Context(), masterID, blockID); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
