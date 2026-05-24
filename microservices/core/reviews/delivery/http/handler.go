package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/microservices/core/reviews/dto"
)

type IService interface {
	CreateReview(ctx context.Context, userID uuid.UUID, req dto.CreateReviewRequest) (*dto.ReviewResponse, error)
	DeleteReview(ctx context.Context, actorID uuid.UUID, actorRole string, reviewID uuid.UUID) error
	GetClientReviews(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]dto.ReviewResponse, error)
	GetMasterReviews(ctx context.Context, masterID uuid.UUID, limit, offset uint64) ([]dto.ReviewResponse, error)
}

type Handler interface {
	RegisterRoutes(public, protected, csrfProtected *mux.Router)
}

type handler struct {
	service IService
}

func NewHandler(service IService) Handler {
	return &handler{service: service}
}
