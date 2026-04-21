package http

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
)

type IService interface {
	GetAvailableSlots(ctx context.Context, serviceID uuid.UUID, startDateStr, endDateStr string) (*dto.GetAvailableSlotsResponse, error)
	CreateAppointment(ctx context.Context, clientID uuid.UUID, req dto.CreateAppointmentRequest) (*dto.AppointmentResponse, error)

	GetMasterIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetClientAppointments(ctx context.Context, clientID uuid.UUID, limit, offset uint64) ([]dto.ClientAppointmentView, error)
	GetMasterAppointments(ctx context.Context, masterID uuid.UUID, start, end time.Time) ([]dto.MasterAppointmentView, error)
	UpdateAppointmentStatus(ctx context.Context, actorID uuid.UUID, appointmentID uuid.UUID, req dto.UpdateAppointmentStatusRequest, isClient bool) error
	CreateManualBlock(ctx context.Context, masterID uuid.UUID, req dto.CreateManualBlockRequest) (*dto.CreateManualBlockResponse, error)
	DeleteManualBlock(ctx context.Context, masterID, blockID uuid.UUID) error
}

type Handler struct {
	service IService
}

func NewHandler(service IService) *Handler {
	return &Handler{
		service: service,
	}
}
