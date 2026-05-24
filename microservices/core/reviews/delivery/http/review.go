package http

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/microservices/core/reviews/dto"
	"github.com/RBS-Team/Okoshki/pkg/response"
)

func (h *handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	const op = "reviews.handler.CreateReview"
	log := middleware.LoggerFromContext(r.Context())
	defer r.Body.Close()

	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok || userIDStr == "" {
		log.Errorf("[%s]: missing user id in context", op)
		response.UnauthorizedJSON(w)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warnf("[%s]: invalid user id format: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	var req dto.CreateReviewRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		log.Warnf("[%s]: failed to unmarshal request: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	rv, err := h.service.CreateReview(r.Context(), userID, req)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, rv)
}

func (h *handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	const op = "reviews.handler.DeleteReview"
	log := middleware.LoggerFromContext(r.Context())

	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok || userIDStr == "" {
		log.Errorf("[%s]: missing user id in context", op)
		response.UnauthorizedJSON(w)
		return
	}

	actorID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warnf("[%s]: invalid user id format: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	actorRole, _ := middleware.GetUserRole(r.Context())

	reviewID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		log.Warnf("[%s]: invalid review id: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	if err := h.service.DeleteReview(r.Context(), actorID, actorRole, reviewID); err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.NoContentJSON(w)
}

func (h *handler) GetClientReviews(w http.ResponseWriter, r *http.Request) {
	const op = "reviews.handler.GetClientReviews"
	log := middleware.LoggerFromContext(r.Context())

	userIDStr, ok := middleware.GetUserID(r.Context())
	if !ok || userIDStr == "" {
		log.Errorf("[%s]: missing user id in context", op)
		response.UnauthorizedJSON(w)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warnf("[%s]: invalid user id format: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	limit, offset := parsePagination(r)

	reviews, err := h.service.GetClientReviews(r.Context(), userID, limit, offset)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, reviews)
}

func (h *handler) GetMasterReviews(w http.ResponseWriter, r *http.Request) {
	const op = "reviews.handler.GetMasterReviews"
	log := middleware.LoggerFromContext(r.Context())

	masterID, err := uuid.Parse(mux.Vars(r)["masterID"])
	if err != nil {
		log.Warnf("[%s]: invalid master id: %v", op, err)
		response.BadRequestJSON(w)
		return
	}

	limit, offset := parsePagination(r)

	reviews, err := h.service.GetMasterReviews(r.Context(), masterID, limit, offset)
	if err != nil {
		log.Errorf("[%s]: service error: %v", op, err)
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, reviews)
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
