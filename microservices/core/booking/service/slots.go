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

	// Проверка корректности дат и валидности диапазона
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

	// Идет в сервис catalog в service_item и по UID сервиса получаем serviceItem
	serviceItem, err := s.catalog.GetServiceItemByID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	// Достаем из нашего serviceItem UID мастера
	masterID, err := uuid.Parse(serviceItem.MasterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid master id in service: %w", op, err)
	}

	// По UID мастера мы получаем DTO мастера со всей инфой о нем
	master, err := s.catalog.GetMasterByID(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	// Достаем из DTO инфу о тайм зоне мастера и получаем инфу о часовом поясе мастера это нужно для синхронизации часов с разных таймзон
	masterLoc, err := time.LoadLocation(master.Timezone)
	if err != nil {
		masterLoc = time.UTC
	}

	// Получаем рабочие часы мастера, это слайс из 7 дней, в которых прописан диапазон работы мастера или является ли день выходным
	workingHours, err := s.catalog.GetWorkingHours(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	// МАПА РАБОЧИХ ЧАСОВ. Преобразовываем полученный слайс в мапу. типа  0: и сам этот workingHour со всеми полями
	whMap := make(map[int]catalogDTO.WorkingHours)
	for _, wh := range workingHours {
		whMap[wh.DayOfWeek] = wh
	}

	// Получаем слайс исключений расписания мастера за интересующий нас период 
	exceptions, err := s.catalog.GetScheduleExceptions(ctx, masterID, startDateStr, endDateStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	// МАПА ИСКЛЮЧЕНИЙ. Преобразовываем полученный слайс в мапу. Типа "2026-04-21" : и само исключение со своими полями. 
	excMap := make(map[string]catalogDTO.ScheduleException)
	for _, exc := range exceptions {
		excMap[exc.ExceptionDate] = exc
	}

	// Подготовка временных границ с 00:00:00 по  23:59:59
	// Границы создаются с интерпретацией временной локации мастера, после чего перевод в UTC для запроса в БД
	queryStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, masterLoc).UTC()
	queryEnd := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, masterLoc).UTC()
	
	// Получаем слайс всех встречи мастера в нашем временном диапазоне
	appointments, err := s.repo.GetActiveAppointmentsByMaster(ctx, masterID, queryStart, queryEnd)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	// Создаем мапу записей. Мы локализуем в часовой пояс мастера время. Ключ дата "2006-01-02" : значения это записи Appointment на эту дату 
	apptsByDate := make(map[string][]model.Appointment)
	for _, a := range appointments {
		a.StartAt = a.StartAt.In(masterLoc)
		a.EndAt = a.EndAt.In(masterLoc)
		// Приводим дату начала записи к формату такого шаблона "2006-01-02"
		dateStr := a.StartAt.Format(dateFormat)
		// Засовывает запись(a типа Appointment)  на какую то дату 
		apptsByDate[dateStr] = append(apptsByDate[dateStr], a)
	}
	
	// Расчитываем полную длительность Слота с учетом Buffers - время на уборку/подготовку к следующему клиенту
	totalDuration := time.Duration(serviceItem.BufferBeforeMinutes+serviceItem.DurationMinutes+serviceItem.BufferAfterMinutes) * time.Minute
	bufferBeforeDuration := time.Duration(serviceItem.BufferBeforeMinutes) * time.Minute

	// Мапа для складирования доступных слотов
	result := &dto.GetAvailableSlotsResponse{
		Slots: make(map[string][]string),
	}
	
	// Цикл по датам. Шаг итерации ровно 1 день.
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format(dateFormat)
		// Инициализация слайса для конкретной даты dateStr в мапе result.Slots
		result.Slots[dateStr] = []string{}

		var workStartStr, workEndStr string
		var isWorking bool
		// Проверяем есть ли исключения в мапе excMap на проверяемую дату
		if exc, ok := excMap[dateStr]; ok {
			
			// Зайдем внутрь если в исключение мы работаем и заданы старт и конец исключения
			if exc.IsWorking && exc.StartTime != nil && exc.EndTime != nil {
				isWorking = true
				workStartStr = *exc.StartTime
				workEndStr = *exc.EndTime
			}
		// Если нет исключений на проверяемую дату
		} else {
			// Превращаем день недели в число, помним что 0 - воскресенье, 1 понедельник и тд
			dayOfWeek := int(d.Weekday())
			// Проверка настроек мастера
			if wh, ok := whMap[dayOfWeek]; ok && !wh.IsDayOff && wh.StartTime != nil && wh.EndTime != nil {
				isWorking = true
				workStartStr = *wh.StartTime
				workEndStr = *wh.EndTime
			}
		}

		// Пропускаем нерабочие дни
		if !isWorking {
			continue
		}
		
		// Конкатинируем к дате dateStr, по которой итерируемся, еще время часов с учетом локации мастера
		workStart, _ := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+workStartStr, masterLoc)
		workEnd, _ := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+workEndStr, masterLoc)

		// Тут лежат записи Appointment на дату dateStr
		dayAppts := apptsByDate[dateStr]
		currentTime := workStart

		// Пытаемся втиснуть услугу totalDuration в оставшееся рабочее время
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
