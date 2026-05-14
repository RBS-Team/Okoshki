package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	minioPkg "github.com/RBS-Team/Okoshki/pkg/minio"
)

// AccountCreator определён здесь и реализуется модулем auth (duck typing).
// users/service не импортирует auth/service — зависимость только через интерфейс.
type AccountCreator interface {
	CreateUser(ctx context.Context, email, password, role string) (uuid.UUID, error)
	DeleteUserByID(ctx context.Context, id uuid.UUID) error
	
}

//go:generate mockgen -destination=../../../../mocks/users/repository/repository_mock.go -package=mock_users_repo github.com/RBS-Team/Okoshki/microservices/core/users/service IRepository
type IRepository interface {
	GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*model.Master, error)
	GetMasterByID(ctx context.Context, id uuid.UUID) (*model.Master, error)
	CreateMaster(ctx context.Context, master model.Master) error
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

type IStorage interface {
	Upload(ctx context.Context, obj minioPkg.ObjectInfo) (string, error)
	BuildObjectURL(bucket, objectName string) string
	Remove(ctx context.Context, bucket, objectName string) error
}

type Service struct {
	auth    AccountCreator
	repo    IRepository
	storage IStorage
}

func New(auth AccountCreator, repo IRepository, storage IStorage) *Service {
	return &Service{
		auth:    auth,
		repo:    repo,
		storage: storage,
	}
}
