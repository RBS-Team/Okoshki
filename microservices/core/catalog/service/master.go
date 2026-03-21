package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
)

func (s *Service) CreateMaster(ctx context.Context, req dto.CreateMasterRequest) (*dto.Master, error) {
	const op = "catalog.service.CreateMaster"

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid user id: %w", op, err)
	}

	tz := req.Timezone
	if tz == "" {
		tz = "Europe/Moscow"
	}

	if _, err := time.LoadLocation(tz); err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, ErrInvalidTimezone)
	}

	masterModel := model.Master{
		ID:           uuid.New(),
		UserID:       userID,
		Name:         req.Name,
		Bio:          req.Bio,
		Timezone:     tz,
		Lat:          req.Lat,
		Lon:          req.Lon,
		Rating:       0,
		ReviewCount:  0,
		ReportsCount: 0,
		IsBlocked:    false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateMaster(ctx, masterModel); err != nil {
		return nil, fmt.Errorf("[%s]: failed to create master: %w", op, mapError(err))
	}

	return mapMasterModelToDTO(&masterModel), nil
}

func (s *Service) GetMasterByID(ctx context.Context, id uuid.UUID) (*dto.Master, error) {
	const op = "catalog.service.GetMasterByID"

	masterModel, err := s.repo.GetMasterByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get master: %w", op, mapError(err))
	}

	return mapMasterModelToDTO(masterModel), nil
}

func (s *Service) GetAllMasters(ctx context.Context, limit, offset uint64) ([]dto.Master, error) {
	const op = "catalog.service.GetAllMasters"

	masterModels, err := s.repo.GetAllMasters(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get masters: %w", op, mapError(err))
	}

	if len(masterModels) == 0 {
		return []dto.Master{}, nil
	}

	masterDTOs := make([]dto.Master, 0, len(masterModels))
	for i := range masterModels {
		masterDTOs = append(masterDTOs, *mapMasterModelToDTO(&masterModels[i]))
	}

	return masterDTOs, nil
}

func (s *Service) GetMastersByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]dto.Master, error) {
	const op = "catalog.service.GetMastersByCategory"

	_, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to validate category: %w", op, mapError(err))
	}

	masterModels, err := s.repo.GetMastersByCategoryID(ctx, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get masters by category: %w", op, mapError(err))
	}

	if len(masterModels) == 0 {
		return []dto.Master{}, nil
	}

	masterDTOs := make([]dto.Master, 0, len(masterModels))
	for i := range masterModels {
		masterDTOs = append(masterDTOs, *mapMasterModelToDTO(&masterModels[i]))
	}

	return masterDTOs, nil
}

func mapMasterModelToDTO(m *model.Master) *dto.Master {
	return &dto.Master{
		ID:          m.ID.String(),
		UserID:      m.UserID.String(),
		Name:        m.Name,
		Bio:         m.Bio,
		AvatarURL:   m.AvatarURL,
		Timezone:    m.Timezone,
		Lat:         m.Lat,
		Lon:         m.Lon,
		Rating:      m.Rating,
		ReviewCount: m.ReviewCount,
	}
}
