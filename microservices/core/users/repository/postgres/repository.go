package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

type Repository interface {
	CreateMaster(ctx context.Context, master model.Master) error
	GetMasterByID(ctx context.Context, id uuid.UUID) (*model.Master, error)
	GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*model.Master, error)
	GetAllMasters(ctx context.Context, limit, offset uint64) ([]model.Master, error)
	GetMastersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Master, error)
	GetMastersByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]model.Master, error)
	UpdateMasterAvatarURL(ctx context.Context, id uuid.UUID, objectName string) error

	CreateClient(ctx context.Context, client model.Client) error
	GetClientByUserID(ctx context.Context, userID uuid.UUID) (*model.Client, error)
	GetClientsByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Client, error)
	UpdateClientAvatarURL(ctx context.Context, id uuid.UUID, objectName string) error

	SavePortfolioPhotos(ctx context.Context, photos []model.PortfolioPhoto) error
	GetPortfolioPhotosByMasterID(ctx context.Context, masterID uuid.UUID) ([]model.PortfolioPhoto, error)
	GetPortfolioPhotoByID(ctx context.Context, photoID uuid.UUID) (*model.PortfolioPhoto, error)
	DeletePortfolioPhotoByID(ctx context.Context, photoID uuid.UUID) error
}

type repository struct {
	db *sql.DB
}

func New(db *sql.DB) Repository {
	return &repository{db: db}
}
