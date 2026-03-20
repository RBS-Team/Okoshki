package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
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

func (s *Service) GetAllCategories(ctx context.Context) ([]dto.Category, error) {
	const op = "catalog.service.GetAllCategories"

	catModels, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get categories: %w", op, mapError(err))
	}

	if len(catModels) == 0 {
		return[]dto.Category{}, nil
	}

	catDTOs := make([]dto.Category, 0, len(catModels))
	for _, catModel := range catModels {
		cat := dto.Category{
			ID:   catModel.ID.String(),
			Name: catModel.Name,
		}

		if catModel.ParentID != nil {
			parentIDStr := catModel.ParentID.String()
			cat.ParentID = &parentIDStr
		}

		if catModel.Description != nil {
			cat.Description = catModel.Description
		}

		catDTOs = append(catDTOs, cat)
	}

	return catDTOs, nil
}