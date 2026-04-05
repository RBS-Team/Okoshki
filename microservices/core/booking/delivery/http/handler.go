package http

import (
	"context"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
)

type IService interface {
	GetAvailableSlots(ctx context.Context, serviceID uuid.UUID, startDateStr, endDateStr string) (*dto.GetAvailableSlotsResponse, error)
	CreateAppointment(ctx context.Context, clientID uuid.UUID, req dto.CreateAppointmentRequest) (*dto.AppointmentResponse, error)
}

type Handler struct {
	service IService
}

func NewHandler(service IService) *Handler {
	return &Handler{
		service: service,
	}
}
