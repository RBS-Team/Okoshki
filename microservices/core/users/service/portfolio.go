package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/users/dto"
	minioPkg "github.com/RBS-Team/Okoshki/pkg/minio"
)

const portfolioBucket = "okoshki-portfolio"

func (s *service) UploadPortfolioPhotos(
	ctx context.Context,
	userIDStr, masterIDStr string,
	files []dto.FileUpload,
) ([]dto.PortfolioPhoto, error) {
	const op = "catalog.service.UploadPortfolioPhotos"

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	masterID, err := uuid.Parse(masterIDStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	authedMaster, err := s.repo.GetMasterByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: get master: %w", op, err)
	}

	if authedMaster.ID != masterID {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrForbidden)
	}

	now := time.Now()
	photos := make([]model.PortfolioPhoto, 0, len(files))

	for _, f := range files {
		objName, err := s.storage.Upload(ctx, minioPkg.ObjectInfo{
			BucketName:   portfolioBucket,
			Reader:       f.Reader,
			Size:         f.Size,
			ContentType:  f.ContentType,
			OriginalName: f.Name,
		})
		if err != nil {
			return nil, fmt.Errorf("[%s]: upload: %w", op, err)
		}

		photos = append(photos, model.PortfolioPhoto{
			ID:         uuid.New(),
			MasterID:   masterID,
			ObjectName: objName,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	if err := s.repo.SavePortfolioPhotos(ctx, photos); err != nil {
		return nil, fmt.Errorf("[%s]: save: %w", op, err)
	}

	return s.buildPhotoDTOs(photos), nil
}

func (s *service) GetPortfolioPhotos(ctx context.Context, masterIDStr string) ([]dto.PortfolioPhoto, error) {
	const op = "catalog.service.GetPortfolioPhotos"

	masterID, err := uuid.Parse(masterIDStr)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	photos, err := s.repo.GetPortfolioPhotosByMasterID(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: get photos: %w", op, err)
	}

	if len(photos) == 0 {
		return []dto.PortfolioPhoto{}, nil
	}

	return s.buildPhotoDTOs(photos), nil
}

func (s *service) DeletePortfolioPhoto(ctx context.Context, userIDStr, masterIDStr, photoIDStr string) error {
	const op = "catalog.service.DeletePortfolioPhoto"

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	masterID, err := uuid.Parse(masterIDStr)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		return fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	authedMaster, err := s.repo.GetMasterByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("[%s]: get master: %w", op, err)
	}

	if authedMaster.ID != masterID {
		return fmt.Errorf("[%s]: %w", op, domain.ErrForbidden)
	}

	photo, err := s.repo.GetPortfolioPhotoByID(ctx, photoID)
	if err != nil {
		return fmt.Errorf("[%s]: get photo: %w", op, err)
	}

	if photo.MasterID != masterID {
		return fmt.Errorf("[%s]: %w", op, domain.ErrForbidden)
	}

	if err := s.storage.Remove(ctx, portfolioBucket, photo.ObjectName); err != nil {
		return fmt.Errorf("[%s]: remove from storage: %w", op, err)
	}

	if err := s.repo.DeletePortfolioPhotoByID(ctx, photoID); err != nil {
		return fmt.Errorf("[%s]: delete from db: %w", op, err)
	}

	return nil
}

func (s *service) buildPhotoDTOs(photos []model.PortfolioPhoto) []dto.PortfolioPhoto {
	result := make([]dto.PortfolioPhoto, 0, len(photos))

	for _, p := range photos {
		result = append(result, dto.PortfolioPhoto{
			ID:       p.ID.String(),
			MasterID: p.MasterID.String(),
			URL:      s.storage.BuildObjectURL(portfolioBucket, p.ObjectName),
		})
	}

	return result
}
