package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
	catalogDTO "github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
)

const (
	dateFormat = "2006-01-02"
	timeFormat = "15:04"
)

func (s *Service) GetAvailableSlots(ctx context.Context, serviceID uuid.UUID, startDateStr, endDateStr string) (*dto.GetAvailableSlotsResponse, error) {
	const op = "booking.service.GetAvailableSlots"

	startDate, err := time.Parse(dateFormat, startDateStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid start_date: %w", op, err)
	}
	endDate, err := time.Parse(dateFormat, endDateStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid end_date: %w", op, err)
	}
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("[%s]: end_date must be after start_date", op)
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

	workingHours, err := s.catalog.GetWorkingHours(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}
	whMap := make(map[int]catalogDTO.WorkingHours)
	for _, wh := range workingHours {
		whMap[wh.DayOfWeek] = wh
	}

	exceptions, err := s.catalog.GetScheduleExceptions(ctx, masterID, startDateStr, endDateStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}
	excMap := make(map[string]catalogDTO.ScheduleException)
	for _, exc := range exceptions {
		excMap[exc.ExceptionDate] = exc
	}

	queryStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, masterLoc).UTC()
	queryEnd := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, masterLoc).UTC()

	appointments, err := s.repo.GetActiveAppointmentsByMaster(ctx, masterID, queryStart, queryEnd)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	apptsByDate := make(map[string][]model.Appointment)
	for _, a := range appointments {
		a.StartAt = a.StartAt.In(masterLoc)
		a.EndAt = a.EndAt.In(masterLoc)
		dateStr := a.StartAt.Format(dateFormat)
		apptsByDate[dateStr] = append(apptsByDate[dateStr], a)
	}

	totalDuration := time.Duration(serviceItem.BufferBeforeMinutes+serviceItem.DurationMinutes+serviceItem.BufferAfterMinutes) * time.Minute
	bufferBeforeDuration := time.Duration(serviceItem.BufferBeforeMinutes) * time.Minute

	result := &dto.GetAvailableSlotsResponse{
		Slots: make(map[string][]string),
	}

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format(dateFormat)
		result.Slots[dateStr] = []string{}

		var workStartStr, workEndStr string
		var isWorking bool

		if exc, ok := excMap[dateStr]; ok {
			if exc.IsWorking && exc.StartTime != nil && exc.EndTime != nil {
				isWorking = true
				workStartStr = *exc.StartTime
				workEndStr = *exc.EndTime
			}
		} else {
			dayOfWeek := int(d.Weekday())
			if wh, ok := whMap[dayOfWeek]; ok && !wh.IsDayOff && wh.StartTime != nil && wh.EndTime != nil {
				isWorking = true
				workStartStr = *wh.StartTime
				workEndStr = *wh.EndTime
			}
		}

		if !isWorking {
			continue
		}

		workStart, _ := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+workStartStr, masterLoc)
		workEnd, _ := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+workEndStr, masterLoc)

		dayAppts := apptsByDate[dateStr]

		currentTime := workStart

		for currentTime.Add(totalDuration).Before(workEnd) || currentTime.Add(totalDuration).Equal(workEnd) {
			slotStart := currentTime
			slotEnd := currentTime.Add(totalDuration)

			hasIntersection := false
			for _, appt := range dayAppts {
				if slotStart.Before(appt.EndAt) && appt.StartAt.Before(slotEnd) {
					hasIntersection = true
					currentTime = appt.EndAt
					break
				}
			}

			if !hasIntersection {
				clientArrivalTime := slotStart.Add(bufferBeforeDuration)
				result.Slots[dateStr] = append(result.Slots[dateStr], clientArrivalTime.Format(timeFormat))

				currentTime = slotEnd
			}
		}
	}

	return result, nil
}
