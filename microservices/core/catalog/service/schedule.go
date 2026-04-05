package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/repository/postgres"
)

const (
	timeFormat      = "15:04:00"
	timeFormatShort = "15:04"
	dateFormat      = "2006-01-02"
)

func (s *Service) UpsertWorkingHours(ctx context.Context, masterID uuid.UUID, req dto.UpdateWorkingHoursBulkRequest) error {
	const op = "catalog.service.UpsertWorkingHours"

	if len(req.Days) != 7 {
		return fmt.Errorf("[%s]: expected exactly 7 days, got %d", op, len(req.Days))
	}

	hoursModels := make([]model.WorkingHours, 0, 7)
	seenDays := make(map[int]bool)

	for _, d := range req.Days {
		if d.DayOfWeek < 0 || d.DayOfWeek > 6 {
			return fmt.Errorf("[%s]: invalid day_of_week %d", op, d.DayOfWeek)
		}
		if seenDays[d.DayOfWeek] {
			return fmt.Errorf("[%s]: duplicate day_of_week %d", op, d.DayOfWeek)
		}
		seenDays[d.DayOfWeek] = true

		if d.IsDayOff {
			d.StartTime = nil
			d.EndTime = nil
		} else {
			if d.StartTime == nil || d.EndTime == nil {
				return fmt.Errorf("[%s]: start_time and end_time are required for working day %d", op, d.DayOfWeek)
			}
			if err := validateTimeFormat(*d.StartTime); err != nil {
				return fmt.Errorf("[%s]: day %d start_time: %w", op, d.DayOfWeek, err)
			}
			if err := validateTimeFormat(*d.EndTime); err != nil {
				return fmt.Errorf("[%s]: day %d end_time: %w", op, d.DayOfWeek, err)
			}
			if *d.StartTime >= *d.EndTime {
				return fmt.Errorf("[%s]: start_time must be before end_time for day %d", op, d.DayOfWeek)
			}
		}

		hoursModels = append(hoursModels, model.WorkingHours{
			MasterID:  masterID,
			DayOfWeek: d.DayOfWeek,
			StartTime: d.StartTime,
			EndTime:   d.EndTime,
			IsDayOff:  d.IsDayOff,
		})
	}

	if err := s.repo.UpsertWorkingHours(ctx, masterID, hoursModels); err != nil {
		return fmt.Errorf("[%s]: %w", op, mapError(err))
	}

	return nil
}

func (s *Service) GetWorkingHours(ctx context.Context, masterID uuid.UUID) ([]dto.WorkingHours, error) {
	const op = "catalog.service.GetWorkingHours"

	hours, err := s.repo.GetWorkingHoursByMasterID(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapError(err))
	}

	dtos := make([]dto.WorkingHours, 0, len(hours))
	for _, h := range hours {
		dtos = append(dtos, dto.WorkingHours{
			ID:        h.ID.String(),
			MasterID:  h.MasterID.String(),
			DayOfWeek: h.DayOfWeek,
			StartTime: formatTimeFromDB(h.StartTime),
			EndTime:   formatTimeFromDB(h.EndTime),
			IsDayOff:  h.IsDayOff,
		})
	}

	return dtos, nil
}

func (s *Service) CreateScheduleException(ctx context.Context, masterID uuid.UUID, req dto.CreateScheduleExceptionRequest) (*dto.ScheduleException, error) {
	const op = "catalog.service.CreateScheduleException"

	parsedDate, err := time.Parse(dateFormat, req.ExceptionDate)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid exception_date format: %w", op, err)
	}

	if !req.IsWorking {
		req.StartTime = nil
		req.EndTime = nil
	} else {
		if req.StartTime == nil || req.EndTime == nil {
			return nil, fmt.Errorf("[%s]: start_time and end_time are required when is_working is true", op)
		}
		if err := validateTimeFormat(*req.StartTime); err != nil {
			return nil, fmt.Errorf("[%s]: start_time: %w", op, err)
		}
		if err := validateTimeFormat(*req.EndTime); err != nil {
			return nil, fmt.Errorf("[%s]: end_time: %w", op, err)
		}
		if *req.StartTime >= *req.EndTime {
			return nil, fmt.Errorf("[%s]: start_time must be before end_time", op)
		}
	}

	excModel := model.ScheduleException{
		ID:            uuid.New(),
		MasterID:      masterID,
		ExceptionDate: parsedDate,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		IsWorking:     req.IsWorking,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateScheduleException(ctx, excModel); err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapError(err))
	}

	return mapExceptionModelToDTO(&excModel), nil
}

func (s *Service) UpdateScheduleException(ctx context.Context, masterID, exceptionID uuid.UUID, req dto.UpdateScheduleExceptionRequest) error {
	const op = "catalog.service.UpdateScheduleException"

	existing, err := s.repo.GetScheduleExceptionByID(ctx, masterID, exceptionID)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, mapError(err))
	}

	isWorking := existing.IsWorking
	if req.IsWorking != nil {
		isWorking = *req.IsWorking
	}

	startTime := existing.StartTime
	if req.StartTime != nil {
		startTime = req.StartTime
	}

	endTime := existing.EndTime
	if req.EndTime != nil {
		endTime = req.EndTime
	}

	if !isWorking {
		startTime = nil
		endTime = nil
	} else {
		if startTime == nil || endTime == nil {
			return fmt.Errorf("[%s]: start_time and end_time are required for working exception", op)
		}

		if req.StartTime != nil {
			if err := validateTimeFormat(*startTime); err != nil {
				return fmt.Errorf("[%s]: start_time: %w", op, err)
			}
		}
		if req.EndTime != nil {
			if err := validateTimeFormat(*endTime); err != nil {
				return fmt.Errorf("[%s]: end_time: %w", op, err)
			}
		}

		if *startTime >= *endTime {
			return fmt.Errorf("[%s]: start_time must be before end_time", op)
		}
	}

	upd := postgres.UpdateScheduleExceptionInput{
		StartTime: startTime,
		EndTime:   endTime,
		IsWorking: &isWorking,
	}

	if err := s.repo.UpdateScheduleException(ctx, masterID, exceptionID, upd); err != nil {
		return fmt.Errorf("[%s]: %w", op, mapError(err))
	}

	return nil
}

func (s *Service) DeleteScheduleException(ctx context.Context, masterID, exceptionID uuid.UUID) error {
	const op = "catalog.service.DeleteScheduleException"

	if err := s.repo.DeleteScheduleException(ctx, masterID, exceptionID); err != nil {
		return fmt.Errorf("[%s]: %w", op, mapError(err))
	}

	return nil
}

func (s *Service) GetScheduleExceptions(ctx context.Context, masterID uuid.UUID, startDateStr, endDateStr string) ([]dto.ScheduleException, error) {
	const op = "catalog.service.GetScheduleExceptions"

	startDate, err := time.Parse(dateFormat, startDateStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid start_date format: %w", op, err)
	}

	endDate, err := time.Parse(dateFormat, endDateStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid end_date format: %w", op, err)
	}

	exceptions, err := s.repo.GetScheduleExceptions(ctx, masterID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapError(err))
	}

	dtos := make([]dto.ScheduleException, 0, len(exceptions))
	for _, e := range exceptions {
		dtos = append(dtos, *mapExceptionModelToDTO(&e))
	}

	return dtos, nil
}

func validateTimeFormat(t string) error {
	_, err := time.Parse(timeFormatShort, t)
	if err == nil {
		return nil
	}
	_, err = time.Parse(timeFormat, t)
	return err
}

func formatTimeFromDB(t *string) *string {
	if t == nil {
		return nil
	}
	parsed, err := time.Parse(timeFormat, *t)
	if err != nil {
		return t
	}
	formatted := parsed.Format(timeFormatShort)
	return &formatted
}

func mapExceptionModelToDTO(m *model.ScheduleException) *dto.ScheduleException {
	return &dto.ScheduleException{
		ID:            m.ID.String(),
		MasterID:      m.MasterID.String(),
		ExceptionDate: m.ExceptionDate.Format(dateFormat),
		StartTime:     formatTimeFromDB(m.StartTime),
		EndTime:       formatTimeFromDB(m.EndTime),
		IsWorking:     m.IsWorking,
	}
}
