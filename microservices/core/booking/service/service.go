package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	authDTO "github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	catalogDTO "github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
)

type AppointmentRepository interface {
	CreateAppointment(ctx context.Context, appt model.Appointment) error
	GetActiveAppointmentsByMaster(ctx context.Context, masterID uuid.UUID, start, end time.Time) ([]model.Appointment, error)
	GetAppointmentsByClientID(ctx context.Context, clientID uuid.UUID, limit, offset uint64) ([]model.Appointment, error)
	GetAppointmentsByMasterID(ctx context.Context, masterID uuid.UUID, start, end time.Time) ([]model.Appointment, error)
	GetAppointmentByID(ctx context.Context, id uuid.UUID) (*model.Appointment, error)
	UpdateAppointmentStatus(ctx context.Context, id uuid.UUID, status model.AppointmentStatus, masterNote *string) error
	DeleteManualBlock(ctx context.Context, id uuid.UUID, masterID uuid.UUID) error
}

type CatalogProvider interface {
	GetServiceItemByID(ctx context.Context, id uuid.UUID) (*catalogDTO.ServiceItem, error)
	GetMasterByID(ctx context.Context, id uuid.UUID) (*catalogDTO.Master, error)
	GetWorkingHours(ctx context.Context, masterID uuid.UUID) ([]catalogDTO.WorkingHours, error)
	GetScheduleExceptions(ctx context.Context, masterID uuid.UUID, startDateStr, endDateStr string) ([]catalogDTO.ScheduleException, error)
	GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*catalogDTO.Master, error)
}

type UserProvider interface {
	GetUsersInfo(ctx context.Context, ids []uuid.UUID) ([]authDTO.UserInfo, error)
}

type Service struct {
	repo    AppointmentRepository
	catalog CatalogProvider
	user    UserProvider
}

func New(repo AppointmentRepository, catalog CatalogProvider, user UserProvider) *Service {
	return &Service{
		repo:    repo,
		catalog: catalog,
		user:    user,
	}
}
