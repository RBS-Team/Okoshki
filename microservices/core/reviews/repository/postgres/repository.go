package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

type Repository interface {
	CreateReview(ctx context.Context, review model.Review) (*model.Review, error)
	GetReviewByID(ctx context.Context, id uuid.UUID) (*model.Review, error)
	GetReviewByAppointmentID(ctx context.Context, appointmentID uuid.UUID) (*model.Review, error)
	DeleteReview(ctx context.Context, id uuid.UUID) error
	RecalcMasterRating(ctx context.Context, masterID uuid.UUID) error
	GetReviewsByClientID(ctx context.Context, clientID uuid.UUID, limit, offset uint64) ([]model.Review, error)
	GetReviewsByMasterID(ctx context.Context, masterID uuid.UUID, limit, offset uint64) ([]model.Review, error)
}

type repository struct {
	db *sql.DB
}

func New(db *sql.DB) Repository {
	return &repository{db: db}
}
