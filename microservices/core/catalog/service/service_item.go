package service

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	usersDTO "github.com/RBS-Team/Okoshki/microservices/core/users/dto"
)

func (s *service) CreateServiceItem(ctx context.Context, masterID uuid.UUID, req dto.CreateServiceItemRequest) (*dto.ServiceItem, error) {
	const op = "catalog.service.CreateServiceItem"

	if titleLen := utf8.RuneCountInString(req.Title); titleLen < 3 || titleLen > 50 {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}
	if addrLen := utf8.RuneCountInString(req.Address); addrLen < 2 || addrLen > 300 {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}
	if cityLen := utf8.RuneCountInString(req.City); cityLen < 2 || cityLen > 100 {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid category id: %w", op, err)
	}

	isAutoConfirm := true
	if req.IsAutoConfirm != nil {
		isAutoConfirm = *req.IsAutoConfirm
	}

	itemModel := model.ServiceItem{
		ID:              uuid.New(),
		MasterID:        masterID,
		CategoryID:      categoryID,
		Title:           req.Title,
		Address:         req.Address,
		City:            req.City,
		Description:     req.Description,
		Price:           req.Price,
		DurationMinutes: req.DurationMinutes,
		Lat:             req.Lat,
		Lon:             req.Lon,
		IsActive:        true,
		IsAutoConfirm:   isAutoConfirm,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.CreateServiceItem(ctx, itemModel); err != nil {
		return nil, fmt.Errorf("[%s]: failed to create service item: %w", op, err)
	}

	return mapServiceItemModelToDTO(&itemModel), nil
}

func (s *service) GetServiceItemsByMasterID(ctx context.Context, masterID uuid.UUID) ([]dto.ServiceItem, error) {
	const op = "catalog.service.GetServiceItemsByMasterID"

	itemModels, err := s.repo.GetServiceItemsByMasterID(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get service items: %w", op, err)
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

func (s *service) GetServicesByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]dto.ServiceWithMaster, error) {
	const op = "catalog.service.GetServicesByCategory"

	_, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to validate category: %w", op, err)
	}

	serviceModels, err := s.repo.GetServicesByCategoryID(ctx, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get services: %w", op, err)
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

	masters, err := s.masters.GetMastersByIDs(ctx, masterIDs)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get masters for services: %w", op, err)
	}

	mastersMap := make(map[uuid.UUID]usersDTO.Master, len(masters))
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
			ID:              srv.ID.String(),
			CategoryID:      srv.CategoryID.String(),
			Title:           srv.Title,
			Address:         srv.Address,
			City:            srv.City,
			Description:     srv.Description,
			Price:           srv.Price,
			DurationMinutes: srv.DurationMinutes,
			Lat:             srv.Lat,
			Lon:             srv.Lon,
			IsActive:        srv.IsActive,
			IsAutoConfirm:   srv.IsAutoConfirm,
			MasterID:        masterDTO.ID,
			FirstName:       masterDTO.FirstName,
			LastName:        masterDTO.LastName,
			Phone:           masterDTO.Phone,
			MasterAddress:   masterDTO.Address,
			MasterCity:      masterDTO.City,
			Bio:             masterDTO.Bio,
			AvatarURL:       masterDTO.AvatarURL,
			Timezone:        masterDTO.Timezone,
			MasterLat:       masterDTO.Lat,
			MasterLon:       masterDTO.Lon,
			Rating:          masterDTO.Rating,
			ReviewCount:     masterDTO.ReviewCount,
		})
	}

	return dtos, nil
}

func (s *service) GetServiceItemByID(ctx context.Context, id uuid.UUID) (*dto.ServiceItem, error) {
	const op = "catalog.service.GetServiceItemByID"

	itemModel, err := s.repo.GetServiceItemByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	return mapServiceItemModelToDTO(itemModel), nil
}

func mapMasterModelToDTO(m *model.Master) *usersDTO.Master {
	return &usersDTO.Master{
		ID:          m.ID.String(),
		UserID:      m.UserID.String(),
		CategoryID:  m.CategoryID.String(),
		FirstName:   m.FirstName,
		LastName:    m.LastName,
		Phone:       m.Phone,
		Address:     m.Address,
		City:        m.City,
		Bio:         m.Bio,
		AvatarURL:   m.AvatarURL,
		Timezone:    m.Timezone,
		Lat:         m.Lat,
		Lon:         m.Lon,
		Rating:      m.Rating,
		ReviewCount: m.ReviewCount,
	}
}

func mapServiceItemModelToDTO(m *model.ServiceItem) *dto.ServiceItem {
	return &dto.ServiceItem{
		ID:              m.ID.String(),
		MasterID:        m.MasterID.String(),
		CategoryID:      m.CategoryID.String(),
		Title:           m.Title,
		Address:         m.Address,
		City:            m.City,
		Description:     m.Description,
		Price:           m.Price,
		DurationMinutes: m.DurationMinutes,
		Lat:             m.Lat,
		Lon:             m.Lon,
		IsActive:        m.IsActive,
		IsAutoConfirm:   m.IsAutoConfirm,
	}
}
