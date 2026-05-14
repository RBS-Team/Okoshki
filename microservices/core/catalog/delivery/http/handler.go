package http

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
)

//go:generate mockgen -destination=../../../../mocks/catalog/service/http/service_mock.go -package=mock_catalog_service github.com/RBS-Team/Okoshki/microservices/core/catalog/delivery/http IService
type IService interface {
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*dto.Category, error)
	GetAllCategories(ctx context.Context) ([]*dto.Category, error)
	UploadCategoryAvatar(ctx context.Context, categoryIDStr string, file io.Reader, size int64, contentType string) error

	CreateServiceItem(ctx context.Context, masterID uuid.UUID, req dto.CreateServiceItemRequest) (*dto.ServiceItem, error)
	GetServiceItemsByMasterID(ctx context.Context, masterID uuid.UUID) ([]dto.ServiceItem, error)
	GetServicesByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]dto.ServiceWithMaster, error)

	GetMasterSettings(ctx context.Context, masterID uuid.UUID) (*dto.MasterSettings, error)
	UpsertMasterSettings(ctx context.Context, masterID uuid.UUID, req dto.UpsertMasterSettingsRequest) error

	CreateWorkInterval(ctx context.Context, masterID uuid.UUID, req dto.CreateWorkIntervalRequest) (*dto.WorkInterval, error)
	DeleteWorkInterval(ctx context.Context, masterID, intervalID uuid.UUID) error
	ListWorkIntervals(ctx context.Context, masterID uuid.UUID, fromStr, toStr string) ([]dto.WorkInterval, error)
	ReplaceWorkIntervalsForDate(ctx context.Context, masterID uuid.UUID, req dto.ReplaceWorkIntervalsForDateRequest) error
}

type Handler struct {
	service IService
}

func NewHandler(service IService) *Handler {
	return &Handler{
		service: service,
	}
}
