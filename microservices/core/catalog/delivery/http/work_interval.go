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

// ListWorkIntervals godoc
// @Summary      Список рабочих интервалов мастера в диапазоне дат
// @Description  Возвращает все интервалы (несколько на день возможно) в [from, to] включительно. Доступно публично.
// @Tags         schedule
// @Produce      json
// @Param        masterID path string true "UUID мастера"
// @Param        from query string true "YYYY-MM-DD"
// @Param        to query string true "YYYY-MM-DD"
// @Success      200 {object} dto.WorkIntervalList
// @Failure      400 {object} response.ErrorResponse
// @Router       /masters/{masterID}/work-intervals [get]
func (h *handler) ListWorkIntervals(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.ListWorkIntervals"
	log := middleware.LoggerFromContext(r.Context())

	masterID, ok := parseMasterIDFromURL(r)
	if !ok {
		response.BadRequestJSON(w)
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		log.Warnf("[%s]: missing from/to", op)
		response.BadRequestJSON(w)
		return
	}

	intervals, err := h.service.ListWorkIntervals(r.Context(), masterID, from, to)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, &dto.WorkIntervalList{Intervals: intervals})
}

// CreateWorkInterval godoc
// @Summary      Создать один рабочий интервал
// @Description  Создаёт интервал на конкретную дату. При пересечении с существующим — 409.
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateWorkIntervalRequest true "Интервал"
// @Success      201 {object} dto.WorkInterval
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /me/work-intervals [post]
func (h *handler) CreateWorkInterval(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.CreateWorkInterval"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, ok := middleware.MasterIDFromContext(r.Context())
	if !ok {
		log.Errorf("[%s]: master id missing in context", op)
		response.ForbiddenJSON(w)
		return
	}

	var req dto.CreateWorkIntervalRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: unmarshal: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	wi, err := h.service.CreateWorkInterval(r.Context(), masterID, req)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, wi)
}

// ReplaceWorkIntervalsForDate godoc
// @Summary      Атомарная замена всех интервалов на конкретную дату
// @Description  Удаляет все существующие интервалы мастера на дату и вставляет переданные. Пустой список = очистить день. Если на дату есть активные записи — 409.
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        request body dto.ReplaceWorkIntervalsForDateRequest true "Новый набор интервалов"
// @Success      200 {object} map[string]string
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /me/work-intervals [put]
func (h *handler) ReplaceWorkIntervalsForDate(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.ReplaceWorkIntervalsForDate"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, ok := middleware.MasterIDFromContext(r.Context())
	if !ok {
		log.Errorf("[%s]: master id missing in context", op)
		response.ForbiddenJSON(w)
		return
	}

	var req dto.ReplaceWorkIntervalsForDateRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: unmarshal: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if err := h.service.ReplaceWorkIntervalsForDate(r.Context(), masterID, req); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DeleteWorkInterval godoc
// @Summary      Удалить один интервал
// @Description  Если внутри интервала есть активные записи — 409.
// @Tags         schedule
// @Produce      json
// @Param        intervalID path string true "UUID интервала"
// @Success      200 {object} map[string]string
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /me/work-intervals/{intervalID} [delete]
func (h *handler) DeleteWorkInterval(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.DeleteWorkInterval"
	log := middleware.LoggerFromContext(r.Context())

	masterID, ok := middleware.MasterIDFromContext(r.Context())
	if !ok {
		log.Errorf("[%s]: master id missing in context", op)
		response.ForbiddenJSON(w)
		return
	}

	intervalID, err := uuid.Parse(mux.Vars(r)["intervalID"])
	if err != nil {
		response.BadRequestJSON(w)
		return
	}

	if err := h.service.DeleteWorkInterval(r.Context(), masterID, intervalID); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
