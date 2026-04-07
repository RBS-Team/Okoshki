package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
)

const dateTimeFormat = "2006-01-02 15:04"

func (s *Service) CreateAppointment(ctx context.Context, clientID uuid.UUID, req dto.CreateAppointmentRequest) (*dto.AppointmentResponse, error) {
	const op = "booking.service.CreateAppointment"

	serviceID, err := uuid.Parse(req.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid service id: %w", op, err)
	}

	serviceItem, err := s.catalog.GetServiceItemByID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	masterID, err := uuid.Parse(serviceItem.MasterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid master id in service: %w", op, err)
	}

	master, err := s.catalog.GetMasterByID(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	masterLoc, err := time.LoadLocation(master.Timezone)
	if err != nil {
		masterLoc = time.UTC
	}

	clientArrivalTime, err := time.ParseInLocation(dateTimeFormat, req.StartAt, masterLoc)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid start_at format (expected YYYY-MM-DD HH:MM): %w", op, err)
	}

	dbStartAt := clientArrivalTime.Add(-time.Duration(serviceItem.BufferBeforeMinutes) * time.Minute).UTC()
	dbEndAt := clientArrivalTime.Add(time.Duration(serviceItem.DurationMinutes+serviceItem.BufferAfterMinutes) * time.Minute).UTC()

	now := time.Now().UTC()
	if dbStartAt.Before(now) {
		return nil, fmt.Errorf("[%s]: cannot book an appointment in the past: %w", op, ErrValidation)
	}

	dateStr := clientArrivalTime.Format(dateFormat)
	timeStr := clientArrivalTime.Format("15:04")

	availableSlotsResp, err := s.GetAvailableSlots(ctx, serviceID, dateStr, dateStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to validate schedule: %w", op, err)
	}

	isValidSlot := false
	if slots, ok := availableSlotsResp.Slots[dateStr]; ok {
		for _, slot := range slots {
			if slot == timeStr {
				isValidSlot = true
				break
			}
		}
	}

	if !isValidSlot {
		return nil, fmt.Errorf("[%s]: time slot %s is not available in master schedule: %w", op, timeStr, ErrValidation)
	}

	status := model.StatusPending
	if serviceItem.IsAutoConfirm {
		status = model.StatusConfirmed
	}

	appt := model.Appointment{
		ID:            uuid.New(),
		ClientID:      clientID,
		MasterID:      masterID,
		ServiceID:     serviceID,
		StartAt:       dbStartAt,
		EndAt:         dbEndAt,
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
		StartAt:       clientArrivalTime,
		EndAt:         clientArrivalTime.Add(time.Duration(serviceItem.DurationMinutes) * time.Minute),
		Status:        string(appt.Status),
		ClientComment: appt.ClientComment,
	}, nil
}
