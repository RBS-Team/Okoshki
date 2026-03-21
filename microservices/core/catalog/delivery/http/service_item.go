package http

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/service"
	"github.com/RBS-Team/Okoshki/pkg/response"
	// "github.com/RBS-Team/Okoshki/internal/middleware"
)

func (h *Handler) CreateServiceItem(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.CreateServiceItem"
	// log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		// log.Errorf("[%s]: master id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	masterID, err := uuid.Parse(idStr)
	if err != nil {
		// log.Warnf("[%s]: invalid master id format: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	var req dto.CreateServiceItemRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		// log.Warnf("[%s]: failed to unmarshal request: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if req.Title == "" || req.CategoryID == "" || req.DurationMinutes <= 0 || req.Price < 0 {
		// log.Warnf("[%s]: invalid request fields", op)
		response.BadRequestJSON(w)
		return
	}

	item, err := h.service.CreateServiceItem(r.Context(), masterID, req)
	if err != nil {
		// log.Errorf("[%s]: service error: %v", op, err)
		h.handleServiceItemError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, item)
}

func (h *Handler) GetServiceItemsByMasterID(w http.ResponseWriter, r *http.Request) {
	const op = "catalog.handler.GetServiceItemsByMasterID"
	// log := middleware.LoggerFromContext(r.Context())

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		// log.Errorf("[%s]: master id is missing in URL vars", op)
		response.BadRequestJSON(w)
		return
	}

	masterID, err := uuid.Parse(idStr)
	if err != nil {
		// log.Warnf("[%s]: invalid master id format: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	items, err := h.service.GetServiceItemsByMasterID(r.Context(), masterID)
	if err != nil {
		if !errors.Is(err, service.ErrNotFound) {
			// log.Errorf("[%s]: service error: %v", op, err)
		}
		h.handleServiceItemError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, items)
}

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

func (h *Handler) handleServiceItemError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.NotFoundJSON(w)
	case errors.Is(err, service.ErrConflict):
		response.ConflictJSON(w)
	default:
		response.InternalErrorJSON(w)
	}
}
