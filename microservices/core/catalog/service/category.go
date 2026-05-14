package service

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	minioPkg "github.com/RBS-Team/Okoshki/pkg/minio"
)

const categoryBucket = "okoshki-categories"

func (s *Service) GetCategoryByID(ctx context.Context, id uuid.UUID) (*dto.Category, error) {
	const op = "catalog.service.GetCategoryByID"

	catModel, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get category: %w", op, err)
	}

	return s.mapCategoryToDTO(catModel), nil
}

func (s *Service) GetAllCategories(ctx context.Context) ([]*dto.Category, error) {
	const op = "catalog.service.GetAllCategories"

	catModels, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get categories from repo: %w", op, err)
	}

	categories := make([]*dto.Category, 0, len(catModels))
	for i := range catModels {
		categories = append(categories, s.mapCategoryToDTO(&catModels[i]))
	}

	return categories, nil
}

func (s *Service) UploadCategoryAvatar(ctx context.Context, categoryIDStr string, file io.Reader, size int64, contentType string) error {
	const op = "catalog.service.UploadCategoryAvatar"

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	cat, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("[%s]: get category: %w", op, err)
	}

	objectName, err := s.storage.Upload(ctx, minioPkg.ObjectInfo{
		BucketName:  categoryBucket,
		Reader:      file,
		Size:        size,
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("[%s]: upload: %w", op, err)
	}

	if err := s.repo.UpdateCategoryAvatarURL(ctx, categoryID, objectName); err != nil {
		_ = s.storage.Remove(ctx, categoryBucket, objectName)
		return fmt.Errorf("[%s]: update avatar url: %w", op, err)
	}

	if cat.AvatarURL != nil {
		_ = s.storage.Remove(ctx, categoryBucket, *cat.AvatarURL)
	}

	return nil
}

func (s *Service) mapCategoryToDTO(cat *model.Category) *dto.Category {
	d := &dto.Category{
		ID:           cat.ID.String(),
		Name:         cat.Name,
		Description:  cat.Description,
		MastersCount: cat.MastersCount,
	}
	if cat.AvatarURL != nil {
		url := s.storage.BuildObjectURL(categoryBucket, *cat.AvatarURL)
		d.AvatarURL = &url
	}
	return d
}
