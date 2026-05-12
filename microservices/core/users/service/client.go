package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/users/dto"
)

func (s *Service) RegisterClient(ctx context.Context, req dto.RegisterClientRequest) (*dto.RegisterClientResponse, error) {
	const op = "users.service.RegisterClient"

	if req.FirstName == "" {
		return nil, fmt.Errorf("[%s]: %w", op, ErrInvalidInput)
	}

	userID, err := s.auth.CreateAccount(ctx, req.Email, req.Password, string(model.RoleClient))
	if err != nil {
		return nil, fmt.Errorf("[%s]: create account: %w", op, err)
	}

	client := model.Client{
		ID:        uuid.New(),
		UserID:    userID,
		FirstName: req.FirstName,
		Phone:     req.Phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateClient(ctx, client); err != nil {
		_ = s.auth.DeleteAccount(ctx, userID)
		return nil, fmt.Errorf("[%s]: create client profile: %w", op, mapError(err))
	}

	return &dto.RegisterClientResponse{
		ID:    userID.String(),
		Email: req.Email,
		Role:  string(model.RoleClient),
	}, nil
}

func (s *Service) GetClientsByIDs(ctx context.Context, ids []uuid.UUID) ([]dto.Client, error) {
	const op = "users.service.GetClientsByIDs"
	clients, err := s.repo.GetClientsByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	result := make([]dto.Client, 0, len(clients))
	for i := range clients {
		result = append(result, s.mapClientToDTO(&clients[i]))
	}

	return result, nil
}
