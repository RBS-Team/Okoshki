package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

type Repository interface {
	CreateUser(ctx context.Context, user model.User) error
	DeleteUserByID(ctx context.Context, id uuid.UUID) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.User, error)
}

type repository struct {
	db *sql.DB
}

func New(db *sql.DB) Repository {
	return &repository{db: db}
}
