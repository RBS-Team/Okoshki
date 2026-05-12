package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

func parseMasterID(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(mux.Vars(r)["masterID"])
	return id, err == nil
}

// UpsertWorkingHours godoc
// @Summary      Создание расписания мастера
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        masterID path string true "UUID мастера"
// @Param        request body dto.UpdateWorkingHoursBulkRequest true "Рабочие часы"
// @Success      200 {object} map[string]string
// @Router       /masters/{masterID}/working-hours [put]
func (h *Handler) UpsertWorkingHours(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.UpsertWorkingHours"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, ok := parseMasterID(r)
	if !ok {
		response.BadRequestJSON(w)
		return
	}

	var req dto.UpdateWorkingHoursBulkRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: failed to unmarshal request: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if err := h.service.UpsertWorkingHours(r.Context(), masterID, req); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleScheduleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetWorkingHours godoc
// @Summary      Получение расписания мастера
// @Tags         schedule
// @Produce      json
// @Param        masterID path string true "UUID мастера"
// @Success      200 {array} dto.WorkingHours
// @Router       /masters/{masterID}/working-hours [get]
func (h *Handler) GetWorkingHours(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetWorkingHours"
	log := middleware.LoggerFromContext(r.Context())

	masterID, ok := parseMasterID(r)
	if !ok {
		response.BadRequestJSON(w)
		return
	}

	hours, err := h.service.GetWorkingHours(r.Context(), masterID)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleScheduleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, hours)
}

// CreateScheduleException godoc
// @Summary      Создание исключения в расписании
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        masterID path string true "UUID мастера"
// @Param        request body dto.CreateScheduleExceptionRequest true "Данные исключения"
// @Success      201 {object} dto.ScheduleException
// @Router       /masters/{masterID}/schedule-exceptions [post]
func (h *Handler) CreateScheduleException(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.CreateScheduleException"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, ok := parseMasterID(r)
	if !ok {
		response.BadRequestJSON(w)
		return
	}

	var req dto.CreateScheduleExceptionRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: failed to unmarshal request: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	exc, err := h.service.CreateScheduleException(r.Context(), masterID, req)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleScheduleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, exc)
}

// UpdateScheduleException godoc
// @Summary      Обновление исключения в расписании
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        masterID path string true "UUID мастера"
// @Param        id path string true "UUID исключения"
// @Param        request body dto.UpdateScheduleExceptionRequest true "Данные"
// @Success      200 {object} map[string]string
// @Router       /masters/{masterID}/schedule-exceptions/{id} [put]
func (h *Handler) UpdateScheduleException(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.UpdateScheduleException"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, ok := parseMasterID(r)
	if !ok {
		response.BadRequestJSON(w)
		return
	}

	exceptionID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		response.BadRequestJSON(w)
		return
	}

	var req dto.UpdateScheduleExceptionRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: failed to unmarshal request: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if err := h.service.UpdateScheduleException(r.Context(), masterID, exceptionID, req); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleScheduleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DeleteScheduleException godoc
// @Summary      Удаление исключения из расписания
// @Tags         schedule
// @Produce      json
// @Param        masterID path string true "UUID мастера"
// @Param        id path string true "UUID исключения"
// @Success      200 {object} map[string]string
// @Router       /masters/{masterID}/schedule-exceptions/{id} [delete]
func (h *Handler) DeleteScheduleException(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.DeleteScheduleException"
	log := middleware.LoggerFromContext(r.Context())

	masterID, ok := parseMasterID(r)
	if !ok {
		response.BadRequestJSON(w)
		return
	}

	exceptionID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		response.BadRequestJSON(w)
		return
	}

	if err := h.service.DeleteScheduleException(r.Context(), masterID, exceptionID); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleScheduleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetScheduleExceptions godoc
// @Summary      Получение исключений в расписании
// @Tags         schedule
// @Produce      json
// @Param        masterID path string true "UUID мастера"
// @Param        start_date query string true "YYYY-MM-DD"
// @Param        end_date query string true "YYYY-MM-DD"
// @Success      200 {array} dto.ScheduleException
// @Router       /masters/{masterID}/schedule-exceptions [get]
func (h *Handler) GetScheduleExceptions(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetScheduleExceptions"
	log := middleware.LoggerFromContext(r.Context())

	masterID, ok := parseMasterID(r)
	if !ok {
		response.BadRequestJSON(w)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if startDate == "" || endDate == "" {
		log.Warnf("[%s]: missing start_date or end_date", op)
		response.BadRequestJSON(w)
		return
	}

	exceptions, err := h.service.GetScheduleExceptions(r.Context(), masterID, startDate, endDate)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleScheduleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, exceptions)
}
