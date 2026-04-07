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
