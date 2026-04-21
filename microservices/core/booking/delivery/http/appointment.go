package http

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/repository/postgres"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/service"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

// GetAvailableSlots godoc
// @Summary      Получение доступных слотов для записи
// @Description  Возвращает список доступных временных слотов для указанной услуги в заданном диапазоне дат. Учитывает рабочие часы мастера, исключения в расписании и существующие записи. Время возвращается в формате HH:MM с учетом буферного времени до начала услуги.
// @Tags         booking
// @Accept       json
// @Produce      json
// @Param        service_id   query string true  "UUID услуги"
// @Param        start_date   query string true  "Начальная дата в формате YYYY-MM-DD (например 2026-04-21)"
// @Param        end_date     query string true  "Конечная дата в формате YYYY-MM-DD (например 2026-04-28)"
// @Success      200 {object} dto.GetAvailableSlotsResponse "Список доступных слотов по дням"
// @Failure      400 {object} response.ErrorResponse "Отсутствуют обязательные query параметры или неверный формат service_id"
// @Failure      404 {object} response.ErrorResponse "Услуга или мастер не найдены"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /available-slots [get]
func (h *Handler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	const op = "booking.handler.GetAvailableSlots"
	log := middleware.LoggerFromContext(r.Context())

	serviceIDStr := r.URL.Query().Get("service_id")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if serviceIDStr == "" || startDateStr == "" || endDateStr == "" {
		log.Warnf("[%s]: missing required query parameters", op)
		response.BadRequestJSON(w)
		return
	}

	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		log.Warnf("[%s]: invalid service_id format: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	slots, err := h.service.GetAvailableSlots(r.Context(), serviceID, startDateStr, endDateStr)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		response.InternalErrorJSON(w)
		return
	}

	response.JSON(w, http.StatusOK, slots)
}

// CreateAppointment godoc
// @Summary      Создание новой записи
// @Description  Создаёт новую запись на услугу для авторизованного клиента. Требуется роль client. Проверяет доступность временного слота и создаёт запись в статусе "pending".
// @Tags         booking
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateAppointmentRequest true "Данные для создания записи"
// @Success      201 {object} dto.AppointmentResponse "Запись успешно создана"
// @Failure      400 {object} response.ErrorResponse "Неверный формат запроса, невалидные данные или неверный формат UUID клиента"
// @Failure      401 {object} response.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} response.ErrorResponse "Доступ запрещен (требуется роль client)"
// @Failure      409 {object} response.ErrorResponse "Временной слот уже занят"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /appointments [post]
func (h *Handler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	const op = "booking.handler.CreateAppointment"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok || userIDStr == "" {
		log.Errorf("[%s]: missing user id in context", op)
		response.UnauthorizedJSON(w)
		return
	}

	clientID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warnf("[%s]: invalid client id format: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	var req dto.CreateAppointmentRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: failed to unmarshal request: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	appt, err := h.service.CreateAppointment(r.Context(), clientID, req)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)

		if errors.Is(err, postgres.ErrTimeConflict) {
			response.JSON(w, http.StatusConflict, response.ErrorResponse{Error: "time slot is already booked"})
			return
		}

		if errors.Is(err, service.ErrValidation) {
			response.BadRequestJSON(w)
			return
		}

		response.InternalErrorJSON(w)
		return
	}

	response.JSON(w, http.StatusCreated, appt)
}
