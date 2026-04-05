package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	catalogDTO "github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
)

type AppointmentRepository interface {
	CreateAppointment(ctx context.Context, appt model.Appointment) error
	GetActiveAppointmentsByMaster(ctx context.Context, masterID uuid.UUID, start, end time.Time) ([]model.Appointment, error)
}

type CatalogProvider interface {
	GetServiceItemByID(ctx context.Context, id uuid.UUID) (*catalogDTO.ServiceItem, error)
	GetMasterByID(ctx context.Context, id uuid.UUID) (*catalogDTO.Master, error)
	GetWorkingHours(ctx context.Context, masterID uuid.UUID) ([]catalogDTO.WorkingHours, error)
	GetScheduleExceptions(ctx context.Context, masterID uuid.UUID, startDateStr, endDateStr string) ([]catalogDTO.ScheduleException, error)
}

type Service struct {
	repo    AppointmentRepository
	catalog CatalogProvider
}

func New(repo AppointmentRepository, catalog CatalogProvider) *Service {
	return &Service{
		repo:    repo,
		catalog: catalog,
	}
}
