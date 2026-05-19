package service

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/users/dto"
)

func (s *service) RegisterClient(ctx context.Context, req dto.RegisterClientRequest) (*dto.RegisterClientResponse, error) {
	const op = "users.service.RegisterClient"

	if !emailRe.MatchString(req.Email) {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}
	if pwdLen := utf8.RuneCountInString(req.Password); pwdLen < 6 || pwdLen > 100 {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}
	if fnLen := utf8.RuneCountInString(req.FirstName); fnLen < 2 || fnLen > 70 {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	userID, err := s.auth.CreateUser(ctx, req.Email, req.Password, string(model.RoleClient))
	if err != nil {
		return nil, fmt.Errorf("[%s]: create account: %w", op, err)
	}

	client := model.Client{
		ID:        uuid.New(),
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateClient(ctx, client); err != nil {
		_ = s.auth.DeleteUserByID(ctx, userID)
		return nil, fmt.Errorf("[%s]: create client profile: %w", op, err)
	}

	return &dto.RegisterClientResponse{
		UserID:    userID.String(),
		ClientID:  client.ID.String(),
		FirstName: client.FirstName,
		LastName:  client.LastName,
		Role:      string(model.RoleClient),
	}, nil
}

func (s *service) GetClientByUserID(ctx context.Context, userID uuid.UUID) (*dto.Client, error) {
	const op = "users.service.GetClientByUserID"

	client, err := s.repo.GetClientByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	result := s.mapClientToDTO(client)
	return &result, nil
}

func (s *service) GetClientsByIDs(ctx context.Context, ids []uuid.UUID) ([]dto.Client, error) {
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
