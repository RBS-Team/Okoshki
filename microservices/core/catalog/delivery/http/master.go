package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/service"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

func (h *Handler) CreateMaster(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.CreateMaster"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	var req dto.CreateMasterRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: failed to unmarshal request: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if req.UserID == "" || req.Name == "" {
		log.Warnf("[%s]: missing required fields: user_id or name", op)
		response.BadRequestJSON(w)
		return
	}

	master, err := h.service.CreateMaster(r.Context(), req)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleMasterError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, master)
}

func (h *Handler) GetMasterByID(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetMasterByID"
	log := middleware.LoggerFromContext(r.Context())

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Errorf("[%s]: id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("[%s]: Failed to parse master ID from URL: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	master, err := h.service.GetMasterByID(r.Context(), id)
	if err != nil {
		if !errors.Is(err, service.ErrNotFound) {
			log.Errorf("[%s]: Service error: %v", op, err)
		}
		h.handleMasterError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, master)
}

func (h *Handler) GetAllMasters(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetAllMasters"
	log := middleware.LoggerFromContext(r.Context())

	limit, offset := parsePagination(r)

	masters, err := h.service.GetAllMasters(r.Context(), limit, offset)
	if err != nil {
		log.Errorf("[%s]: Service error: %v", op, err)
		h.handleMasterError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, masters)
}

func (h *Handler) GetMastersByCategory(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetMastersByCategory"
	log := middleware.LoggerFromContext(r.Context())

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		log.Errorf("[%s]: category id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		log.Warnf("[%s]: Failed to parse category ID from URL: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	limit, offset := parsePagination(r)

	masters, err := h.service.GetMastersByCategory(r.Context(), categoryID, limit, offset)
	if err != nil {
		if !errors.Is(err, service.ErrNotFound) {
			log.Errorf("[%s]: Service error: %v", op, err)
		}
		h.handleMasterError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, masters)
}

func parsePagination(r *http.Request) (uint64, uint64) {
	query := r.URL.Query()

	limit, err := strconv.ParseUint(query.Get("limit"), 10, 64)
	if err != nil || limit == 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.ParseUint(query.Get("offset"), 10, 64)
	if err != nil {
		offset = 0
	}

	return limit, offset
}
