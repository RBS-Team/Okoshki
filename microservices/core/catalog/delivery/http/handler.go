package http

import (
	"context"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
)

//go:generate mockgen -destination=../../../../mocks/catalog/service/http/service_mock.go -package=mock_catalog_service github.com/RBS-Team/Okoshki/microservices/core/catalog/delivery/http IService
type IService interface {
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*dto.Category, error)
	GetAllCategories(ctx context.Context) ([]*dto.Category, error)

	GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*dto.Master, error)
	CreateMaster(ctx context.Context, userIDStr string, req dto.CreateMasterRequest) (*dto.Master, error)
	GetMasterByID(ctx context.Context, id uuid.UUID) (*dto.Master, error)
	GetAllMasters(ctx context.Context, limit, offset uint64) ([]dto.Master, error)
	GetMastersByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]dto.Master, error)

	CreateServiceItem(ctx context.Context, masterID uuid.UUID, req dto.CreateServiceItemRequest) (*dto.ServiceItem, error)
	GetServiceItemsByMasterID(ctx context.Context, masterID uuid.UUID) ([]dto.ServiceItem, error)
	GetServicesByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]dto.ServiceWithMaster, error)

	UpsertWorkingHours(ctx context.Context, masterID uuid.UUID, req dto.UpdateWorkingHoursBulkRequest) error
	GetWorkingHours(ctx context.Context, masterID uuid.UUID) ([]dto.WorkingHours, error)

	CreateScheduleException(ctx context.Context, masterID uuid.UUID, req dto.CreateScheduleExceptionRequest) (*dto.ScheduleException, error)
	UpdateScheduleException(ctx context.Context, masterID, exceptionID uuid.UUID, req dto.UpdateScheduleExceptionRequest) error
	DeleteScheduleException(ctx context.Context, masterID, exceptionID uuid.UUID) error
	GetScheduleExceptions(ctx context.Context, masterID uuid.UUID, startDateStr, endDateStr string) ([]dto.ScheduleException, error)
}

type Handler struct {
	service IService
}

func NewHandler(service IService) *Handler {
	return &Handler{
		service: service,
	}
}
