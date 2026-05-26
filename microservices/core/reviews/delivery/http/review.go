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

// @Summary      Оставить отзыв
// @Description  Клиент оставляет отзыв на выполненный приём. Один приём — один отзыв.
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security     CookieAuth
// @Param        request body dto.CreateReviewRequest true "Данные отзыва"
// @Success      201 {object} dto.ReviewResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /reviews [post]
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

// @Summary      Удалить отзыв
// @Description  Автор отзыва или администратор могут удалить отзыв по его UUID.
// @Tags         reviews
// @Produce      json
// @Security     CookieAuth
// @Param        id path string true "UUID отзыва"
// @Success      204
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /reviews/{id} [delete]
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

// @Summary      Мои отзывы
// @Description  Возвращает список отзывов, оставленных текущим клиентом. Поддерживает пагинацию.
// @Tags         reviews
// @Produce      json
// @Security     CookieAuth
// @Param        limit  query int false "Количество записей (1–100, по умолчанию 20)"
// @Param        offset query int false "Смещение (по умолчанию 0)"
// @Success      200 {array}  dto.ReviewResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      403 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /reviews/my [get]
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

// @Summary      Отзывы о мастере
// @Description  Публично возвращает список отзывов о мастере по его UUID. Поддерживает пагинацию.
// @Tags         reviews
// @Produce      json
// @Param        masterID path string true  "UUID мастера"
// @Param        limit    query int    false "Количество записей (1–100, по умолчанию 20)"
// @Param        offset   query int    false "Смещение (по умолчанию 0)"
// @Success      200 {array}  dto.ReviewResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /reviews/master/{masterID} [get]
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
