package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

type Repository interface {
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

type repository struct {
	db *sql.DB
}

func New(db *sql.DB) Repository {
	return &repository{db: db}
}
