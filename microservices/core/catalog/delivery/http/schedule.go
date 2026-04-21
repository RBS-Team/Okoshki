package http

import (
	"context"
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

func (h *Handler) resolveMasterID(ctx context.Context) (uuid.UUID, error) {
	userIDStr, ok := middleware.GetUserID(ctx)
	if !ok || userIDStr == "" {
		return uuid.Nil, errors.New("unauthorized")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid user id format")
	}

	master, err := h.service.GetMasterByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(master.ID)
}

// UpsertWorkingHours godoc
// @Summary      Обновление рабочих часов мастера
// @Description  Создаёт или обновляет расписание рабочих часов мастера на неделю. Требуется роль master и наличие созданного профиля мастера.
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        request body dto.UpdateWorkingHoursBulkRequest true "Массив рабочих часов по дням недели"
// @Success      200 {object} map[string]string "Расписание успешно обновлено"
// @Failure      400 {object} response.ErrorResponse "Неверный формат запроса или невалидные данные"
// @Failure      401 {object} response.ErrorResponse "Пользователь не авторизован"
// @Failure      403 {object} response.ErrorResponse "Профиль мастера не создан"
// @Failure      500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /working-hours [put]
func (h *Handler) UpsertWorkingHours(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.UpsertWorkingHours"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, err := h.resolveMasterID(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.JSON(w, http.StatusForbidden, response.ErrorResponse{Error: "master profile not created"})
			return
		}
		log.Errorf("[%s]: failed to resolve master ID: %v", op, err)
		response.UnauthorizedJSON(w)
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

func (h *Handler) GetWorkingHours(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetWorkingHours"
	log := middleware.LoggerFromContext(r.Context())

	masterID, err := h.resolveMasterID(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.JSON(w, http.StatusForbidden, response.ErrorResponse{Error: "master profile not created"})
			return
		}
		log.Errorf("[%s]: failed to resolve master ID: %v", op, err)
		response.UnauthorizedJSON(w)
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

func (h *Handler) CreateScheduleException(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.CreateScheduleException"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, err := h.resolveMasterID(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.JSON(w, http.StatusForbidden, response.ErrorResponse{Error: "master profile not created"})
			return
		}
		log.Errorf("[%s]: failed to resolve master ID: %v", op, err)
		response.UnauthorizedJSON(w)
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

func (h *Handler) UpdateScheduleException(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.UpdateScheduleException"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, err := h.resolveMasterID(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.JSON(w, http.StatusForbidden, response.ErrorResponse{Error: "master profile not created"})
			return
		}
		log.Errorf("[%s]: failed to resolve master ID: %v", op, err)
		response.UnauthorizedJSON(w)
		return
	}

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Errorf("[%s]: id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	exceptionID, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("[%s]: invalid exception id format: %v", op, err)
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

func (h *Handler) DeleteScheduleException(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.DeleteScheduleException"
	log := middleware.LoggerFromContext(r.Context())

	masterID, err := h.resolveMasterID(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.JSON(w, http.StatusForbidden, response.ErrorResponse{Error: "master profile not created"})
			return
		}
		log.Errorf("[%s]: failed to resolve master ID: %v", op, err)
		response.UnauthorizedJSON(w)
		return
	}

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Errorf("[%s]: id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	exceptionID, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("[%s]: invalid exception id format: %v", op, err)
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

func (h *Handler) GetScheduleExceptions(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetScheduleExceptions"
	log := middleware.LoggerFromContext(r.Context())

	masterID, err := h.resolveMasterID(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.JSON(w, http.StatusForbidden, response.ErrorResponse{Error: "master profile not created"})
			return
		}
		log.Errorf("[%s]: failed to resolve master ID: %v", op, err)
		response.UnauthorizedJSON(w)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" || endDate == "" {
		log.Warnf("[%s]: missing start_date or end_date query params", op)
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
