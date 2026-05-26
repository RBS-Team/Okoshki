package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/reviews/dto"
)

func (s *service) CreateReview(ctx context.Context, userID uuid.UUID, req dto.CreateReviewRequest) (*dto.ReviewResponse, error) {
	const op = "reviews.service.CreateReview"

	if req.Rating < 1 || req.Rating > 5 {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}
	if req.Comment == "" {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	appointmentID, err := uuid.Parse(req.AppointmentID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid appointment_id: %w", op, domain.ErrInvalidInput)
	}

	appt, err := s.booking.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}
	if appt.Status != model.StatusCompleted {
		return nil, fmt.Errorf("[%s]: appointment is not completed: %w", op, domain.ErrInvalidInput)
	}

	client, err := s.user.GetClientByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	clientID, err := uuid.Parse(client.ID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid client id: %w", op, domain.ErrInternal)
	}

	if appt.ClientID.String() != client.UserID {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrForbidden)
	}

	serviceItem, err := s.catalog.GetServiceItemByID(ctx, appt.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	created, err := s.repo.CreateReview(ctx, model.Review{
		ID:               uuid.New(),
		ClientID:         clientID,
		MasterID:         appt.MasterID,
		AppointmentID:    appointmentID,
		Rating:           req.Rating,
		Comment:          req.Comment,
		ServiceItemTitle: serviceItem.Title,
	})
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	if err := s.repo.RecalcMasterRating(ctx, created.MasterID); err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	return &dto.ReviewResponse{
		ID:               created.ID.String(),
		ClientID:         created.ClientID.String(),
		MasterID:         created.MasterID.String(),
		AppointmentID:    created.AppointmentID.String(),
		Rating:           created.Rating,
		Comment:          created.Comment,
		ServiceItemTitle: created.ServiceItemTitle,
		CreatedAt:        created.CreatedAt,
	}, nil
}

func (s *service) DeleteReview(ctx context.Context, actorID uuid.UUID, actorRole string, reviewID uuid.UUID) error {
	const op = "reviews.service.DeleteReview"

	rv, err := s.repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	if actorRole != string(model.RoleAdmin) {
		client, err := s.user.GetClientByUserID(ctx, actorID)
		if err != nil {
			return fmt.Errorf("[%s]: %w", op, err)
		}
		clientID, err := uuid.Parse(client.ID)
		if err != nil {
			return fmt.Errorf("[%s]: invalid client id: %w", op, domain.ErrInternal)
		}
		if rv.ClientID != clientID {
			return fmt.Errorf("[%s]: %w", op, domain.ErrForbidden)
		}
	}

	masterID := rv.MasterID
	if err := s.repo.DeleteReview(ctx, reviewID); err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	if err := s.repo.RecalcMasterRating(ctx, masterID); err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	return nil
}

func (s *service) GetClientReviews(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]dto.ReviewResponse, error) {
	const op = "reviews.service.GetClientReviews"

	client, err := s.user.GetClientByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	clientID, err := uuid.Parse(client.ID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid client id: %w", op, domain.ErrInternal)
	}

	reviews, err := s.repo.GetReviewsByClientID(ctx, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	result := make([]dto.ReviewResponse, 0, len(reviews))
	for _, rv := range reviews {
		result = append(result, dto.ReviewResponse{
			ID:               rv.ID.String(),
			ClientID:         rv.ClientID.String(),
			MasterID:         rv.MasterID.String(),
			AppointmentID:    rv.AppointmentID.String(),
			Rating:           rv.Rating,
			Comment:          rv.Comment,
			ServiceItemTitle: rv.ServiceItemTitle,
			CreatedAt:        rv.CreatedAt,
		})
	}
	return result, nil
}

func (s *service) GetMasterReviews(ctx context.Context, masterID uuid.UUID, limit, offset uint64) ([]dto.ReviewResponse, error) {
	const op = "reviews.service.GetMasterReviews"

	reviews, err := s.repo.GetReviewsByMasterID(ctx, masterID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	result := make([]dto.ReviewResponse, 0, len(reviews))
	for _, rv := range reviews {
		result = append(result, dto.ReviewResponse{
			ID:               rv.ID.String(),
			ClientID:         rv.ClientID.String(),
			MasterID:         rv.MasterID.String(),
			AppointmentID:    rv.AppointmentID.String(),
			Rating:           rv.Rating,
			Comment:          rv.Comment,
			ServiceItemTitle: rv.ServiceItemTitle,
			CreatedAt:        rv.CreatedAt,
		})
	}
	return result, nil
}
