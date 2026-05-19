package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	catalogDTO "github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
	usersDTO "github.com/RBS-Team/Okoshki/microservices/core/users/dto"
)

type AppointmentRepository interface {
	CreateAppointment(ctx context.Context, appt model.Appointment) error
	GetActiveAppointmentsByMaster(ctx context.Context, masterID uuid.UUID, start, end time.Time) ([]model.Appointment, error)
	GetAppointmentsByClientID(ctx context.Context, clientID uuid.UUID, limit, offset uint64) ([]model.Appointment, error)
	GetAppointmentsByMasterID(ctx context.Context, masterID uuid.UUID, start, end time.Time, status model.AppointmentStatus) ([]model.Appointment, error)
	GetAppointmentByID(ctx context.Context, id uuid.UUID) (*model.Appointment, error)
	UpdateAppointmentStatus(ctx context.Context, id uuid.UUID, status model.AppointmentStatus, masterNote *string) error
	DeleteManualBlock(ctx context.Context, id uuid.UUID, masterID uuid.UUID) error
}

type CatalogProvider interface {
	GetServiceItemByID(ctx context.Context, id uuid.UUID) (*catalogDTO.ServiceItem, error)
	GetMasterSettings(ctx context.Context, masterID uuid.UUID) (*catalogDTO.MasterSettings, error)
	ListWorkIntervals(ctx context.Context, masterID uuid.UUID, fromStr, toStr string) ([]catalogDTO.WorkInterval, error)
}

type UserProvider interface {
	GetClientsByIDs(ctx context.Context, ids []uuid.UUID) ([]usersDTO.Client, error)
	GetMasterByID(ctx context.Context, id uuid.UUID) (*usersDTO.Master, error)
	GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*usersDTO.Master, error)
}

type Service interface {
	GetAvailableSlots(ctx context.Context, serviceID uuid.UUID, fromStr, toStr string) (*dto.GetAvailableSlotsResponse, error)
	CreateAppointment(ctx context.Context, clientID uuid.UUID, req dto.CreateAppointmentRequest) (*dto.AppointmentResponse, error)
	GetMasterIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetClientAppointments(ctx context.Context, clientID uuid.UUID, limit, offset uint64) ([]dto.ClientAppointmentView, error)
	GetMasterAppointments(ctx context.Context, masterID uuid.UUID, start, end time.Time, status model.AppointmentStatus) ([]dto.MasterAppointmentView, error)
	UpdateAppointmentStatus(ctx context.Context, actorID uuid.UUID, appointmentID uuid.UUID, req dto.UpdateAppointmentStatusRequest, isClient bool) error
	CreateManualBlock(ctx context.Context, masterID uuid.UUID, req dto.CreateManualBlockRequest) (*dto.CreateManualBlockResponse, error)
	DeleteManualBlock(ctx context.Context, masterID, blockID uuid.UUID) error
}

type service struct {
	repo    AppointmentRepository
	catalog CatalogProvider
	user    UserProvider
}

func New(repo AppointmentRepository, catalog CatalogProvider, user UserProvider) Service {
	return &service{
		repo:    repo,
		catalog: catalog,
		user:    user,
	}
}
