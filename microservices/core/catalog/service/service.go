package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

//go:generate mockgen -destination=../../../../mocks/catalog/repository/repository_mock.go -package=mock_catalog_repo github.com/RBS-Team/Okoshki/microservices/core/catalog/service IRepository
type IRepository interface {
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*model.Category, error)
	GetAllCategories(ctx context.Context) ([]model.Category, error)

	CreateMaster(ctx context.Context, master model.Master) error
	GetMasterByID(ctx context.Context, id uuid.UUID) (*model.Master, error)
	GetAllMasters(ctx context.Context, limit, offset uint64) ([]model.Master, error)
	GetMastersByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]model.Master, error)
	GetMastersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Master, error)

	CreateServiceItem(ctx context.Context, item model.ServiceItem) error
	GetServiceItemsByMasterID(ctx context.Context, masterID uuid.UUID) ([]model.ServiceItem, error)
	GetServicesByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]model.ServiceItem, error)
}

type Service struct {
	repo IRepository
}

func New(repo IRepository) *Service {
	return &Service{
		repo: repo,
	}
}
