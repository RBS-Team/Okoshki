package service

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	usersDTO "github.com/RBS-Team/Okoshki/microservices/core/users/dto"
	minioPkg "github.com/RBS-Team/Okoshki/pkg/minio"
)

//go:generate mockgen -destination=../../../../mocks/catalog/repository/repository_mock.go -package=mock_catalog_repo github.com/RBS-Team/Okoshki/microservices/core/catalog/service IRepository
type IRepository interface {
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*model.Category, error)
	GetAllCategories(ctx context.Context) ([]model.Category, error)
	UpdateCategoryAvatarURL(ctx context.Context, id uuid.UUID, objectName string) error

	CreateServiceItem(ctx context.Context, item model.ServiceItem) error
	GetServiceItemsByMasterID(ctx context.Context, masterID uuid.UUID) ([]model.ServiceItem, error)
	GetServicesByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]model.ServiceItem, error)
	GetServiceItemByID(ctx context.Context, id uuid.UUID) (*model.ServiceItem, error)

	GetMasterSettings(ctx context.Context, masterID uuid.UUID) (*model.MasterSettings, error)
	UpsertMasterSettings(ctx context.Context, masterID uuid.UUID, slotStep, leadTime *int) error

	CreateWorkInterval(ctx context.Context, wi model.WorkInterval) error
	DeleteWorkInterval(ctx context.Context, masterID, intervalID uuid.UUID) error
	GetWorkIntervalByID(ctx context.Context, masterID, intervalID uuid.UUID) (*model.WorkInterval, error)
	GetWorkIntervalsByMasterRange(ctx context.Context, masterID uuid.UUID, from, to time.Time) ([]model.WorkInterval, error)
	ReplaceWorkIntervalsForDate(ctx context.Context, masterID uuid.UUID, workDate time.Time, intervals []model.WorkInterval) error
	HasActiveAppointmentsInRange(ctx context.Context, masterID uuid.UUID, startUTC, endUTC time.Time) (bool, error)
}

type MasterProvider interface {
	GetMastersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Master, error)
	GetMasterByID(ctx context.Context, id uuid.UUID) (*usersDTO.Master, error)
	GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*usersDTO.Master, error)
}

type IStorage interface {
	Upload(ctx context.Context, obj minioPkg.ObjectInfo) (string, error)
	BuildObjectURL(bucket, objectName string) string
	Remove(ctx context.Context, bucket, objectName string) error
}

type Service interface {
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*dto.Category, error)
	GetAllCategories(ctx context.Context) ([]*dto.Category, error)
	UploadCategoryAvatar(ctx context.Context, categoryIDStr string, file io.Reader, size int64, contentType string) error

	GetMasterSettings(ctx context.Context, masterID uuid.UUID) (*dto.MasterSettings, error)
	UpsertMasterSettings(ctx context.Context, masterID uuid.UUID, req dto.UpsertMasterSettingsRequest) error

	CreateServiceItem(ctx context.Context, masterID uuid.UUID, req dto.CreateServiceItemRequest) (*dto.ServiceItem, error)
	GetServiceItemsByMasterID(ctx context.Context, masterID uuid.UUID) ([]dto.ServiceItem, error)
	GetServicesByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]dto.ServiceWithMaster, error)
	GetServiceItemByID(ctx context.Context, id uuid.UUID) (*dto.ServiceItem, error)

	CreateWorkInterval(ctx context.Context, masterID uuid.UUID, req dto.CreateWorkIntervalRequest) (*dto.WorkInterval, error)
	DeleteWorkInterval(ctx context.Context, masterID, intervalID uuid.UUID) error
	ListWorkIntervals(ctx context.Context, masterID uuid.UUID, fromStr, toStr string) ([]dto.WorkInterval, error)
	ReplaceWorkIntervalsForDate(ctx context.Context, masterID uuid.UUID, req dto.ReplaceWorkIntervalsForDateRequest) error
}

type service struct {
	repo    IRepository
	masters MasterProvider
	storage IStorage
}

func New(repo IRepository, masters MasterProvider, storage IStorage) Service {
	return &service{
		repo:    repo,
		masters: masters,
		storage: storage,
	}
}
