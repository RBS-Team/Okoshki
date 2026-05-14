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

// parseMasterIDFromURL — для публичных эндпоинтов, где master_id задаётся в URL.
func parseMasterIDFromURL(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(mux.Vars(r)["masterID"])
	return id, err == nil
}

// GetMasterSettings godoc
// @Summary      Получение настроек расписания текущего мастера
// @Description  Возвращает шаг сетки и lead time. Если у мастера ещё нет своих настроек — отдаёт дефолты.
// @Tags         schedule
// @Produce      json
// @Success      200 {object} dto.MasterSettings
// @Failure      401 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Router       /me/settings [get]
func (h *Handler) GetMasterSettings(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetMasterSettings"
	log := middleware.LoggerFromContext(r.Context())

	masterID, ok := middleware.MasterIDFromContext(r.Context())
	if !ok {
		log.Errorf("[%s]: master id missing in context", op)
		response.ForbiddenJSON(w)
		return
	}

	settings, err := h.service.GetMasterSettings(r.Context(), masterID)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, settings)
}

// UpsertMasterSettings godoc
// @Summary      Обновление настроек расписания текущего мастера
// @Description  Поля с null не меняются. slot_step_minutes допустимые значения: 5, 10, 15, 20, 30, 60.
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        request body dto.UpsertMasterSettingsRequest true "Настройки"
// @Success      200 {object} map[string]string
// @Failure      400 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Router       /me/settings [put]
func (h *Handler) UpsertMasterSettings(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.UpsertMasterSettings"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	masterID, ok := middleware.MasterIDFromContext(r.Context())
	if !ok {
		log.Errorf("[%s]: master id missing in context", op)
		response.ForbiddenJSON(w)
		return
	}

	var req dto.UpsertMasterSettingsRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: unmarshal: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if err := h.service.UpsertMasterSettings(r.Context(), masterID, req); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
