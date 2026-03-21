package service

import (
	"context"
	"fmt"

	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	"github.com/google/uuid"
)

func (s *Service) GetCategoryByID(ctx context.Context, id uuid.UUID) (*dto.Category, error) {
	const op = "catalog.service.GetCategoryByID"

	catModel, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get category: %w", op, mapError(err))
	}

	catDTO := &dto.Category{
		ID:   catModel.ID.String(),
		Name: catModel.Name,
	}

	if catModel.ParentID != nil {
		parentIDStr := catModel.ParentID.String()
		catDTO.ParentID = &parentIDStr
	}

	if catModel.Description != nil {
		catDTO.Description = catModel.Description
	}

	return catDTO, nil
}

func (s *Service) GetAllCategories(ctx context.Context) ([]*dto.Category, error) {
	const op = "catalog.service.GetAllCategories"

	catModels, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get categories from repo: %w", op, mapError(err))
	}

	if len(catModels) == 0 {
		return []*dto.Category{}, nil
	}

	categoryMap := make(map[uuid.UUID]*dto.Category, len(catModels))
	for _, catModel := range catModels {
		dtoCat := &dto.Category{
			ID:          catModel.ID.String(),
			Name:        catModel.Name,
			Description: catModel.Description,
			Children:    []*dto.Category{},
		}
		if catModel.ParentID != nil {
			parentIDStr := catModel.ParentID.String()
			dtoCat.ParentID = &parentIDStr
		}
		categoryMap[catModel.ID] = dtoCat
	}

	rootCategories := make([]*dto.Category, 0)
	for _, catModel := range catModels {
		if catModel.ParentID == nil {
			rootCategories = append(rootCategories, categoryMap[catModel.ID])
		} else {
			if parentDTO, ok := categoryMap[*catModel.ParentID]; ok {
				parentDTO.Children = append(parentDTO.Children, categoryMap[catModel.ID])
			}
		}
	}

	return rootCategories, nil
}
