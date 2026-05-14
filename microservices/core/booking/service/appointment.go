package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
)

const dateTimeFormat = "2006-01-02 15:04"

func (s *Service) CreateAppointment(ctx context.Context, clientID uuid.UUID, req dto.CreateAppointmentRequest) (*dto.AppointmentResponse, error) {
	const op = "booking.service.CreateAppointment"

	serviceID, err := uuid.Parse(req.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid service id: %w", op, domain.ErrInvalidInput)
	}

	serviceItem, err := s.catalog.GetServiceItemByID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	masterID, err := uuid.Parse(serviceItem.MasterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid master id in service: %w", op, err)
	}

	master, err := s.user.GetMasterByID(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	masterLoc, err := time.LoadLocation(master.Timezone)
	if err != nil {
		masterLoc = time.UTC
	}

	settings, err := s.catalog.GetMasterSettings(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	clientArrival, err := time.ParseInLocation(dateTimeFormat, req.StartAt, masterLoc)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid start_at format (expected YYYY-MM-DD HH:MM): %w", op, domain.ErrInvalidInput)
	}

	startUTC := clientArrival.UTC()
	endUTC := clientArrival.Add(time.Duration(serviceItem.DurationMinutes) * time.Minute).UTC()

	leadCutoffUTC := time.Now().UTC().Add(time.Duration(settings.LeadTimeMinutes) * time.Minute)
	if startUTC.Before(leadCutoffUTC) {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrLeadTimeViolation)
	}

	dateStr := clientArrival.Format(dateFormat)
	timeStr := clientArrival.Format(timeFormat)

	availableSlots, err := s.GetAvailableSlots(ctx, serviceID, dateStr, dateStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to validate schedule: %w", op, err)
	}

	if !slices.Contains(availableSlots.Slots[dateStr], timeStr) {
		return nil, fmt.Errorf("[%s]: time slot %s: %w", op, timeStr, domain.ErrSlotNotAvailable)
	}

	status := model.StatusPending
	if serviceItem.IsAutoConfirm {
		status = model.StatusConfirmed
	}

	now := time.Now().UTC()
	appt := model.Appointment{
		ID:            uuid.New(),
		ClientID:      clientID,
		MasterID:      masterID,
		ServiceID:     serviceID,
		StartAt:       startUTC,
		EndAt:         endUTC,
		Status:        status,
		IsManualBlock: false,
		ClientComment: req.ClientComment,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateAppointment(ctx, appt); err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	return &dto.AppointmentResponse{
		ID:            appt.ID.String(),
		ClientID:      appt.ClientID.String(),
		MasterID:      appt.MasterID.String(),
		ServiceID:     appt.ServiceID.String(),
		StartAt:       clientArrival,
		EndAt:         clientArrival.Add(time.Duration(serviceItem.DurationMinutes) * time.Minute),
		Status:        string(appt.Status),
		ClientComment: appt.ClientComment,
	}, nil
}

