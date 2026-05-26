package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
)

const (
	dateFormat = "2006-01-02"
	timeFormat = "15:04"
)

// CreateWorkInterval создаёт один интервал. Дата/время валидируются.
// Если на эту дату уже есть пересекающийся интервал — domain.ErrIntervalOverlap.
func (s *service) CreateWorkInterval(ctx context.Context, masterID uuid.UUID, req dto.CreateWorkIntervalRequest) (*dto.WorkInterval, error) {
	const op = "catalog.service.CreateWorkInterval"

	workDate, err := time.Parse(dateFormat, req.Date)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid date (expected YYYY-MM-DD): %w", op, domain.ErrInvalidInput)
	}

	if err := validateTimeRange(req.StartTime, req.EndTime); err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	wi := model.WorkInterval{
		ID:        uuid.New(),
		MasterID:  masterID,
		WorkDate:  workDate,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateWorkInterval(ctx, wi); err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	return mapWorkIntervalToDTO(&wi), nil
}

// DeleteWorkInterval удаляет интервал.
// Если внутри удаляемого интервала есть активные записи (pending/confirmed) — domain.ErrIntervalHasAppointments.
func (s *service) DeleteWorkInterval(ctx context.Context, masterID, intervalID uuid.UUID) error {
	const op = "catalog.service.DeleteWorkInterval"

	wi, err := s.repo.GetWorkIntervalByID(ctx, masterID, intervalID)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	masterLoc, err := s.loadMasterTZ(ctx, masterID)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	startUTC, endUTC, err := intervalToUTCRange(wi.WorkDate, wi.StartTime, wi.EndTime, masterLoc)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	hasAppts, err := s.repo.HasActiveAppointmentsInRange(ctx, masterID, startUTC, endUTC)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}
	if hasAppts {
		return fmt.Errorf("[%s]: %w", op, domain.ErrIntervalHasAppointments)
	}

	if err := s.repo.DeleteWorkInterval(ctx, masterID, intervalID); err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	return nil
}

// ListWorkIntervals возвращает все интервалы мастера в диапазоне [from, to] включительно.
func (s *service) ListWorkIntervals(ctx context.Context, masterID uuid.UUID, fromStr, toStr string) ([]dto.WorkInterval, error) {
	const op = "catalog.service.ListWorkIntervals"

	from, err := time.Parse(dateFormat, fromStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid from date: %w", op, domain.ErrInvalidInput)
	}
	to, err := time.Parse(dateFormat, toStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid to date: %w", op, domain.ErrInvalidInput)
	}
	if to.Before(from) {
		return nil, fmt.Errorf("[%s]: to must be >= from: %w", op, domain.ErrInvalidInput)
	}

	intervals, err := s.repo.GetWorkIntervalsByMasterRange(ctx, masterID, from, to)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	result := make([]dto.WorkInterval, 0, len(intervals))
	for i := range intervals {
		result = append(result, *mapWorkIntervalToDTO(&intervals[i]))
	}

	return result, nil
}

// ReplaceWorkIntervalsForDate атомарно заменяет все интервалы мастера на конкретную дату.
// Каждая активная запись (pending/confirmed) на эту дату должна полностью попадать
// хотя бы в один из новых интервалов; иначе — domain.ErrIntervalHasAppointments.
func (s *service) ReplaceWorkIntervalsForDate(ctx context.Context, masterID uuid.UUID, req dto.ReplaceWorkIntervalsForDateRequest) error {
	const op = "catalog.service.ReplaceWorkIntervalsForDate"

	workDate, err := time.Parse(dateFormat, req.Date)
	if err != nil {
		return fmt.Errorf("[%s]: invalid date (expected YYYY-MM-DD): %w", op, domain.ErrInvalidInput)
	}

	for i, ii := range req.Intervals {
		if err := validateTimeRange(ii.StartTime, ii.EndTime); err != nil {
			return fmt.Errorf("[%s]: interval[%d]: %w", op, i, err)
		}
	}

	if err := assertNoOverlap(req.Intervals); err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	masterLoc, err := s.loadMasterTZ(ctx, masterID)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	dayStartUTC := time.Date(workDate.Year(), workDate.Month(), workDate.Day(), 0, 0, 0, 0, masterLoc).UTC()
	dayEndUTC := dayStartUTC.Add(24 * time.Hour)

	hasAppts, err := s.repo.HasActiveAppointmentsInRange(ctx, masterID, dayStartUTC, dayEndUTC)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}
	if hasAppts {
		// Простая защитная политика: запрещаем замену дня, если на нём вообще есть записи.
		// Иначе пришлось бы проверять, что каждая запись лежит внутри какого-то нового интервала.
		return fmt.Errorf("[%s]: %w", op, domain.ErrIntervalHasAppointments)
	}

	models := make([]model.WorkInterval, 0, len(req.Intervals))
	now := time.Now()
	for _, ii := range req.Intervals {
		models = append(models, model.WorkInterval{
			ID:        uuid.New(),
			MasterID:  masterID,
			WorkDate:  workDate,
			StartTime: ii.StartTime,
			EndTime:   ii.EndTime,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	if err := s.repo.ReplaceWorkIntervalsForDate(ctx, masterID, workDate, models); err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	return nil
}

// loadMasterTZ — общий хелпер чтения таймзоны мастера.
func (s *service) loadMasterTZ(ctx context.Context, masterID uuid.UUID) (*time.Location, error) {
	master, err := s.masters.GetMasterByID(ctx, masterID)
	if err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation(master.Timezone)
	if err != nil {
		return time.UTC, nil
	}
	return loc, nil
}

// validateTimeRange проверяет формат и порядок HH:MM.
func validateTimeRange(start, end string) error {
	st, err := time.Parse(timeFormat, start)
	if err != nil {
		return fmt.Errorf("invalid start_time (HH:MM): %w", domain.ErrInvalidInput)
	}
	en, err := time.Parse(timeFormat, end)
	if err != nil {
		return fmt.Errorf("invalid end_time (HH:MM): %w", domain.ErrInvalidInput)
	}
	if !st.Before(en) {
		return fmt.Errorf("start_time must be < end_time: %w", domain.ErrInvalidInput)
	}
	return nil
}

// assertNoOverlap проверяет, что в одном запросе на замену дня нет пересекающихся интервалов.
// БД защитит дополнительно (EXCLUDE), но мы хотим осмысленную ошибку до похода в транзакцию.
func assertNoOverlap(intervals []dto.IntervalInput) error {
	type rng struct{ start, end time.Time }
	parsed := make([]rng, 0, len(intervals))
	for _, ii := range intervals {
		s, _ := time.Parse(timeFormat, ii.StartTime)
		e, _ := time.Parse(timeFormat, ii.EndTime)
		parsed = append(parsed, rng{s, e})
	}
	for i := 0; i < len(parsed); i++ {
		for j := i + 1; j < len(parsed); j++ {
			if parsed[i].start.Before(parsed[j].end) && parsed[j].start.Before(parsed[i].end) {
				return fmt.Errorf("intervals %d and %d overlap: %w", i, j, domain.ErrIntervalOverlap)
			}
		}
	}
	return nil
}

// intervalToUTCRange конвертирует локальные start/end в UTC-полуоткрытый диапазон.
func intervalToUTCRange(workDate time.Time, startStr, endStr string, loc *time.Location) (time.Time, time.Time, error) {
	st, err := time.ParseInLocation(timeFormat, startStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse start: %w", err)
	}
	en, err := time.ParseInLocation(timeFormat, endStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parse end: %w", err)
	}
	startLocal := time.Date(workDate.Year(), workDate.Month(), workDate.Day(), st.Hour(), st.Minute(), 0, 0, loc)
	endLocal := time.Date(workDate.Year(), workDate.Month(), workDate.Day(), en.Hour(), en.Minute(), 0, 0, loc)
	return startLocal.UTC(), endLocal.UTC(), nil
}

func mapWorkIntervalToDTO(wi *model.WorkInterval) *dto.WorkInterval {
	return &dto.WorkInterval{
		ID:        wi.ID.String(),
		MasterID:  wi.MasterID.String(),
		Date:      wi.WorkDate.Format(dateFormat),
		StartTime: trimTimeToHHMM(wi.StartTime),
		EndTime:   trimTimeToHHMM(wi.EndTime),
	}
}

// trimTimeToHHMM приводит "15:04:00" → "15:04".
// Postgres возвращает TIME как HH:MM:SS, в DTO мы хотим HH:MM.
func trimTimeToHHMM(t string) string {
	if len(t) >= 5 {
		return t[:5]
	}
	return t
}
