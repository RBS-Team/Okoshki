package http

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/repository/postgres"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

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

		response.InternalErrorJSON(w)
		return
	}

	response.JSON(w, http.StatusCreated, appt)
}
