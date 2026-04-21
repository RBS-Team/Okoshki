package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	authDTO "github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/microservices/core/booking/dto"
)

var validTransitions = map[model.AppointmentStatus][]model.AppointmentStatus{
	model.StatusPending:   {model.StatusConfirmed, model.StatusRejected, model.StatusCancelled},
	model.StatusConfirmed: {model.StatusCompleted, model.StatusCancelled},
}

func isValidTransition(current, next model.AppointmentStatus) bool {
	allowed, ok := validTransitions[current]
	if !ok {
		return false
	}
	for _, status := range allowed {
		if status == next {
			return true
		}
	}
	return false
}

func (s *Service) UpdateAppointmentStatus(ctx context.Context, actorID uuid.UUID, appointmentID uuid.UUID, req dto.UpdateAppointmentStatusRequest, isClient bool) error {
	const op = "booking.service.UpdateAppointmentStatus"

	appt, err := s.repo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	if isClient {
		if appt.ClientID != actorID {
			return fmt.Errorf("[%s]: access denied", op)
		}
		if req.Status != string(model.StatusCancelled) {
			return fmt.Errorf("[%s]: client can only cancel appointments", op)
		}
	} else {
		// actorID здесь - это MasterID
		if appt.MasterID != actorID {
			return fmt.Errorf("[%s]: access denied", op)
		}
	}

	newStatus := model.AppointmentStatus(req.Status)

	if !isValidTransition(appt.Status, newStatus) {
		return fmt.Errorf("[%s]: invalid status transition from %s to %s", op, appt.Status, newStatus)
	}

	if err := s.repo.UpdateAppointmentStatus(ctx, appointmentID, newStatus, req.MasterNote); err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	return nil
}

func (s *Service) CreateManualBlock(ctx context.Context, masterID uuid.UUID, req dto.CreateManualBlockRequest) (*dto.CreateManualBlockResponse, error) {
	const op = "booking.service.CreateManualBlock"

	master, err := s.catalog.GetMasterByID(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	masterLoc, err := time.LoadLocation(master.Timezone)
	if err != nil {
		masterLoc = time.UTC
	}

	startAt, err := time.ParseInLocation(dateTimeFormat, req.StartAt, masterLoc)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid start_at format", op)
	}

	endAt, err := time.ParseInLocation(dateTimeFormat, req.EndAt, masterLoc)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid end_at format", op)
	}

	if !startAt.Before(endAt) {
		return nil, fmt.Errorf("[%s]: start_at must be before end_at", op)
	}

	appt := model.Appointment{
		ID:            uuid.New(),
		MasterID:      masterID,
		ClientID:      uuid.Nil,
		ServiceID:     uuid.Nil,
		StartAt:       startAt.UTC(),
		EndAt:         endAt.UTC(),
		Status:        model.StatusConfirmed,
		IsManualBlock: true,
		MasterNote:    req.Note,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.repo.CreateAppointment(ctx, appt); err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	return &dto.CreateManualBlockResponse{
		ID:      appt.ID.String(),
		StartAt: startAt,
		EndAt:   endAt,
		Note:    req.Note,
	}, nil
}

func (s *Service) DeleteManualBlock(ctx context.Context, masterID, blockID uuid.UUID) error {
	const op = "booking.service.DeleteManualBlock"
	if err := s.repo.DeleteManualBlock(ctx, blockID, masterID); err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}
	return nil
}

func (s *Service) GetClientAppointments(ctx context.Context, clientID uuid.UUID, limit, offset uint64) ([]dto.ClientAppointmentView, error) {
	const op = "booking.service.GetClientAppointments"

	appts, err := s.repo.GetAppointmentsByClientID(ctx, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	if len(appts) == 0 {
		return []dto.ClientAppointmentView{}, nil
	}

	var views []dto.ClientAppointmentView
	for _, a := range appts {
		master, err := s.catalog.GetMasterByID(ctx, a.MasterID)
		if err != nil {
			continue
		}
		service, err := s.catalog.GetServiceItemByID(ctx, a.ServiceID)
		if err != nil {
			continue
		}

		views = append(views, dto.ClientAppointmentView{
			ID:            a.ID.String(),
			MasterID:      a.MasterID.String(),
			MasterName:    master.Name,
			MasterAvatar:  master.AvatarURL,
			MasterLat:     master.Lat,
			MasterLon:     master.Lon,
			ServiceID:     a.ServiceID.String(),
			ServiceTitle:  service.Title,
			Price:         service.Price,
			Duration:      service.DurationMinutes,
			StartAt:       a.StartAt,
			EndAt:         a.EndAt,
			Status:        string(a.Status),
			ClientComment: a.ClientComment,
			MasterNote:    a.MasterNote,
		})
	}

	return views, nil
}

func (s *Service) GetMasterAppointments(ctx context.Context, masterID uuid.UUID, start, end time.Time) ([]dto.MasterAppointmentView, error) {
	const op = "booking.service.GetMasterAppointments"

	appts, err := s.repo.GetAppointmentsByMasterID(ctx, masterID, start, end)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	if len(appts) == 0 {
		return []dto.MasterAppointmentView{}, nil
	}

	var clientIDs []uuid.UUID
	for _, a := range appts {
		if !a.IsManualBlock && a.ClientID != uuid.Nil {
			clientIDs = append(clientIDs, a.ClientID)
		}
	}

	usersInfo, err := s.user.GetUsersInfo(ctx, clientIDs)
	if err != nil {
		usersInfo = []authDTO.UserInfo{}
	}
	userMap := make(map[string]authDTO.UserInfo)
	for _, u := range usersInfo {
		userMap[u.ID] = u
	}

	var views []dto.MasterAppointmentView
	for _, a := range appts {
		view := dto.MasterAppointmentView{
			ID:            a.ID.String(),
			StartAt:       a.StartAt,
			EndAt:         a.EndAt,
			Status:        string(a.Status),
			IsManualBlock: a.IsManualBlock,
			ClientComment: a.ClientComment,
			MasterNote:    a.MasterNote,
		}

		if !a.IsManualBlock {
			cID := a.ClientID.String()
			sID := a.ServiceID.String()
			view.ClientID = &cID
			view.ServiceID = &sID

			if ui, ok := userMap[cID]; ok {
				view.ClientEmail = &ui.Email
				view.ClientAvatar = &ui.AvatarURL
			}

			service, err := s.catalog.GetServiceItemByID(ctx, a.ServiceID)
			if err == nil {
				view.ServiceTitle = &service.Title
				view.Price = &service.Price
				view.Duration = &service.DurationMinutes
			}
		}

		views = append(views, view)
	}

	return views, nil
}

func (s *Service) GetMasterIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	master, err := s.catalog.GetMasterByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(master.ID)
}
