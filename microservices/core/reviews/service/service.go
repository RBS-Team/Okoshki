package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	catalogDTO "github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	"github.com/RBS-Team/Okoshki/microservices/core/reviews/dto"
	usersDTO "github.com/RBS-Team/Okoshki/microservices/core/users/dto"
)

type IRepository interface {
	CreateReview(ctx context.Context, review model.Review) (*model.Review, error)
	GetReviewByID(ctx context.Context, id uuid.UUID) (*model.Review, error)
	GetReviewByAppointmentID(ctx context.Context, appointmentID uuid.UUID) (*model.Review, error)
	DeleteReview(ctx context.Context, id uuid.UUID) error
	RecalcMasterRating(ctx context.Context, masterID uuid.UUID) error
	GetReviewsByClientID(ctx context.Context, clientID uuid.UUID, limit, offset uint64) ([]model.Review, error)
	GetReviewsByMasterID(ctx context.Context, masterID uuid.UUID, limit, offset uint64) ([]model.Review, error)
}

type BookingProvider interface {
	GetAppointmentByID(ctx context.Context, id uuid.UUID) (*model.Appointment, error)
}

type UserProvider interface {
	GetClientByUserID(ctx context.Context, userID uuid.UUID) (*usersDTO.Client, error)
}

type CatalogProvider interface {
	GetServiceItemByID(ctx context.Context, id uuid.UUID) (*catalogDTO.ServiceItem, error)
}

type Service interface {
	CreateReview(ctx context.Context, userID uuid.UUID, req dto.CreateReviewRequest) (*dto.ReviewResponse, error)
	DeleteReview(ctx context.Context, actorID uuid.UUID, actorRole string, reviewID uuid.UUID) error
	GetClientReviews(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]dto.ReviewResponse, error)
	GetMasterReviews(ctx context.Context, masterID uuid.UUID, limit, offset uint64) ([]dto.ReviewResponse, error)
}

type service struct {
	repo    IRepository
	booking BookingProvider
	user    UserProvider
	catalog CatalogProvider
}

func New(repo IRepository, booking BookingProvider, user UserProvider, catalog CatalogProvider) Service {
	return &service{repo: repo, booking: booking, user: user, catalog: catalog}
}
