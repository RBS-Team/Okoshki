package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/repository/postgres"
	usersDTO "github.com/RBS-Team/Okoshki/microservices/core/users/dto"
)

//go:generate mockgen -destination=../../../../mocks/catalog/repository/repository_mock.go -package=mock_catalog_repo github.com/RBS-Team/Okoshki/microservices/core/catalog/service IRepository
type IRepository interface {
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*model.Category, error)
	GetAllCategories(ctx context.Context) ([]model.Category, error)

	CreateServiceItem(ctx context.Context, item model.ServiceItem) error
	GetServiceItemsByMasterID(ctx context.Context, masterID uuid.UUID) ([]model.ServiceItem, error)
	GetServicesByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]model.ServiceItem, error)
	GetServiceItemByID(ctx context.Context, id uuid.UUID) (*model.ServiceItem, error)

	UpsertWorkingHours(ctx context.Context, masterID uuid.UUID, hours []model.WorkingHours) error
	GetWorkingHoursByMasterID(ctx context.Context, masterID uuid.UUID) ([]model.WorkingHours, error)

	CreateScheduleException(ctx context.Context, exc model.ScheduleException) error
	GetScheduleExceptionByID(ctx context.Context, masterID, exceptionID uuid.UUID) (*model.ScheduleException, error)
	GetScheduleExceptions(ctx context.Context, masterID uuid.UUID, startDate, endDate time.Time) ([]model.ScheduleException, error)
	UpdateScheduleException(ctx context.Context, masterID, exceptionID uuid.UUID, upd postgres.UpdateScheduleExceptionInput) error
	DeleteScheduleException(ctx context.Context, masterID, exceptionID uuid.UUID) error
}

type MasterProvider interface {
	GetMastersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Master, error)
	GetMasterByID(ctx context.Context, id uuid.UUID) (*usersDTO.Master, error)
	GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*usersDTO.Master, error)
}

type Service struct {
	repo    IRepository
	masters MasterProvider
}

func New(repo IRepository, masters MasterProvider) *Service {
	return &Service{
		repo:    repo,
		masters: masters,
	}
}
