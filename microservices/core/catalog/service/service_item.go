package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
)

func (s *Service) CreateServiceItem(ctx context.Context, masterID uuid.UUID, req dto.CreateServiceItemRequest) (*dto.ServiceItem, error) {
	const op = "catalog.service.CreateServiceItem"

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid category id: %w", op, err)
	}

	itemModel := model.ServiceItem{
		ID:                  uuid.New(),
		MasterID:            masterID,
		CategoryID:          categoryID,
		Title:               req.Title,
		Description:         req.Description,
		Price:               req.Price,
		DurationMinutes:     req.DurationMinutes,
		BufferBeforeMinutes: req.BufferBeforeMinutes,
		BufferAfterMinutes:  req.BufferAfterMinutes,
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.repo.CreateServiceItem(ctx, itemModel); err != nil {
		return nil, fmt.Errorf("[%s]: failed to create service item: %w", op, mapError(err))
	}

	return mapServiceItemModelToDTO(&itemModel), nil
}

func (s *Service) GetServiceItemsByMasterID(ctx context.Context, masterID uuid.UUID) ([]dto.ServiceItem, error) {
	const op = "catalog.service.GetServiceItemsByMasterID"

	itemModels, err := s.repo.GetServiceItemsByMasterID(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get service items: %w", op, mapError(err))
	}

	if len(itemModels) == 0 {
		return[]dto.ServiceItem{}, nil
	}

	itemDTOs := make([]dto.ServiceItem, 0, len(itemModels))
	for i := range itemModels {
		itemDTOs = append(itemDTOs, *mapServiceItemModelToDTO(&itemModels[i]))
	}

	return itemDTOs, nil
}

func mapServiceItemModelToDTO(m *model.ServiceItem) *dto.ServiceItem {
	return &dto.ServiceItem{
		ID:                  m.ID.String(),
		MasterID:            m.MasterID.String(),
		CategoryID:          m.CategoryID.String(),
		Title:               m.Title,
		Description:         m.Description,
		Price:               m.Price,
		DurationMinutes:     m.DurationMinutes,
		BufferBeforeMinutes: m.BufferBeforeMinutes,
		BufferAfterMinutes:  m.BufferAfterMinutes,
		IsActive:            m.IsActive,
	}
}