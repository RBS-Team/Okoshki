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
		return []dto.ServiceItem{}, nil
	}

	itemDTOs := make([]dto.ServiceItem, 0, len(itemModels))
	for i := range itemModels {
		itemDTOs = append(itemDTOs, *mapServiceItemModelToDTO(&itemModels[i]))
	}

	return itemDTOs, nil
}

func (s *Service) GetServicesByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]dto.ServiceWithMaster, error) {
	const op = "catalog.service.GetServicesByCategory"

	_, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to validate category: %w", op, mapError(err))
	}

	serviceModels, err := s.repo.GetServicesByCategoryID(ctx, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get services: %w", op, mapError(err))
	}

	if len(serviceModels) == 0 {
		return []dto.ServiceWithMaster{}, nil
	}

	uniqueMasterIDs := make(map[uuid.UUID]bool)
	var masterIDs []uuid.UUID
	for _, srv := range serviceModels {
		if !uniqueMasterIDs[srv.MasterID] {
			uniqueMasterIDs[srv.MasterID] = true
			masterIDs = append(masterIDs, srv.MasterID)
		}
	}

	masters, err := s.repo.GetMastersByIDs(ctx, masterIDs)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get masters for services: %w", op, mapError(err))
	}

	mastersMap := make(map[uuid.UUID]dto.Master, len(masters))
	for i := range masters {
		mastersMap[masters[i].ID] = *mapMasterModelToDTO(&masters[i])
	}

	dtos := make([]dto.ServiceWithMaster, 0, len(serviceModels))
	for _, srv := range serviceModels {
		masterDTO, ok := mastersMap[srv.MasterID]
		if !ok {
			continue
		}

		dtos = append(dtos, dto.ServiceWithMaster{
			ID:                  srv.ID.String(),
			CategoryID:          srv.CategoryID.String(),
			Title:               srv.Title,
			Description:         srv.Description,
			Price:               srv.Price,
			DurationMinutes:     srv.DurationMinutes,
			BufferBeforeMinutes: srv.BufferBeforeMinutes,
			BufferAfterMinutes:  srv.BufferAfterMinutes,
			IsActive:            srv.IsActive,
			Master:              masterDTO,
		})
	}

	return dtos, nil
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
