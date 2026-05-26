package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
	catalogDTO "github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
)

const (
	dateFormat = "2006-01-02"
	timeFormat = "15:04"
)

// GetAvailableSlots — основной алгоритм генерации доступных слотов записи на услугу.
//
// Источник истины:
//   - master_work_intervals (несколько интервалов на день; "выходной" = отсутствие интервалов).
//   - master_settings.slot_step_minutes (шаг сетки) и lead_time_minutes.
//   - appointments (активные: pending/confirmed) — занимают пересекающиеся слоты.
//
// Длительность услуги НЕ обязана быть кратной шагу сетки. Слот валиден, если:
//   - помещается целиком в один из рабочих интервалов;
//   - не пересекается ни с одной активной записью;
//   - его начало (UTC) >= now + lead_time.
func (s *service) GetAvailableSlots(ctx context.Context, serviceID uuid.UUID, fromStr, toStr string) (*dto.GetAvailableSlotsResponse, error) {
	const op = "booking.service.GetAvailableSlots"

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

	intervals, err := s.catalog.ListWorkIntervals(ctx, masterID, fromStr, toStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	queryStartUTC := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, masterLoc).UTC()
	queryEndUTC := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, masterLoc).Add(24 * time.Hour).UTC()

	appointments, err := s.repo.GetActiveAppointmentsByMaster(ctx, masterID, queryStartUTC, queryEndUTC)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	stepDur := time.Duration(settings.SlotStepMinutes) * time.Minute
	leadDur := time.Duration(settings.LeadTimeMinutes) * time.Minute
	serviceDur := time.Duration(serviceItem.DurationMinutes) * time.Minute
	leadCutoffUTC := time.Now().UTC().Add(leadDur)

	intervalsByDate := groupIntervalsByDate(intervals)
	apptsByDate := groupAppointmentsByDate(appointments, masterLoc)

	result := &dto.GetAvailableSlotsResponse{Slots: make(map[string][]string)}

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format(dateFormat)

		dayIntervals, ok := intervalsByDate[dateStr]
		if !ok {
			continue
		}
		dayAppts := apptsByDate[dateStr]

		var daySlots []string
		for _, wi := range dayIntervals {
			daySlots = append(daySlots, generateSlotsForInterval(d, wi, dayAppts, serviceDur, stepDur, masterLoc, leadCutoffUTC)...)
		}

		if len(daySlots) > 0 {
			result.Slots[dateStr] = daySlots
		}
	}

	return result, nil
}

// generateSlotsForInterval перебирает кандидатные слоты внутри одного интервала с шагом step.
// Возвращает HH:MM в локальной таймзоне мастера.
func generateSlotsForInterval(date time.Time, wi catalogDTO.WorkInterval, dayAppts []model.Appointment, serviceDur, step time.Duration, loc *time.Location, leadCutoffUTC time.Time) []string {
	intervalStart, intervalEnd, ok := parseIntervalLocal(date, wi.StartTime, wi.EndTime, loc)
	if !ok {
		return nil
	}

	slots := make([]string, 0)
	for cur := intervalStart; !cur.Add(serviceDur).After(intervalEnd); cur = cur.Add(step) {
		slotStartUTC := cur.UTC()
		slotEndUTC := cur.Add(serviceDur).UTC()

		if slotStartUTC.Before(leadCutoffUTC) {
			continue
		}
		if intersectsAny(slotStartUTC, slotEndUTC, dayAppts) {
			continue
		}
		slots = append(slots, cur.Format(timeFormat))
	}
	return slots
}

// intersectsAny — true, если [start, end) пересекается хотя бы с одной записью.
func intersectsAny(startUTC, endUTC time.Time, appts []model.Appointment) bool {
	for _, a := range appts {
		if startUTC.Before(a.EndAt) && a.StartAt.Before(endUTC) {
			return true
		}
	}
	return false
}

// parseIntervalLocal превращает (date, "HH:MM", "HH:MM", loc) в локальные time.Time границы интервала.
func parseIntervalLocal(date time.Time, startStr, endStr string, loc *time.Location) (time.Time, time.Time, bool) {
	st, err := time.Parse(timeFormat, startStr)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	en, err := time.Parse(timeFormat, endStr)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	start := time.Date(date.Year(), date.Month(), date.Day(), st.Hour(), st.Minute(), 0, 0, loc)
	end := time.Date(date.Year(), date.Month(), date.Day(), en.Hour(), en.Minute(), 0, 0, loc)
	return start, end, true
}

func groupIntervalsByDate(intervals []catalogDTO.WorkInterval) map[string][]catalogDTO.WorkInterval {
	out := make(map[string][]catalogDTO.WorkInterval, len(intervals))
	for _, wi := range intervals {
		out[wi.Date] = append(out[wi.Date], wi)
	}
	return out
}

// groupAppointmentsByDate раскладывает записи по датам в локальной таймзоне мастера.
// Считаем, что одна запись лежит в одном дне (start_at). Это допустимо — записи не должны
// пересекать полночь. Если такое произойдёт, end будет учтён при пересечении в соседнем дне.
func groupAppointmentsByDate(appts []model.Appointment, loc *time.Location) map[string][]model.Appointment {
	out := make(map[string][]model.Appointment, len(appts))
	for _, a := range appts {
		dateStr := a.StartAt.In(loc).Format(dateFormat)
		out[dateStr] = append(out[dateStr], a)
	}
	return out
}
