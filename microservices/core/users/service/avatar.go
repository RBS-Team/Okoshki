package service

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/users/dto"
	minioPkg "github.com/RBS-Team/Okoshki/pkg/minio"
)

func (s *Service) UploadMasterAvatar(ctx context.Context, userIDStr, masterIDStr string, file io.Reader, size int64, contentType string) (string, error) {
	const op = "users.service.UploadMasterAvatar"

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	masterID, err := uuid.Parse(masterIDStr)
	if err != nil {
		return "", fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	master, err := s.repo.GetMasterByID(ctx, masterID)
	if err != nil {
		return "", fmt.Errorf("[%s]: get master: %w", op, err)
	}

	if master.UserID != userID {
		return "", fmt.Errorf("[%s]: %w", op, domain.ErrForbidden)
	}

	objectName, err := s.storage.Upload(ctx, minioPkg.ObjectInfo{
		BucketName:  usersBucket,
		Reader:      file,
		Size:        size,
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("[%s]: upload: %w", op, err)
	}

	if err := s.repo.UpdateMasterAvatarURL(ctx, masterID, objectName); err != nil {
		_ = s.storage.Remove(ctx, usersBucket, objectName)
		return "", fmt.Errorf("[%s]: update avatar url: %w", op, err)
	}

	if master.AvatarURL != nil {
		_ = s.storage.Remove(ctx, usersBucket, *master.AvatarURL)
	}

	return s.storage.BuildObjectURL(usersBucket, objectName), nil
}

func (s *Service) UploadClientAvatar(ctx context.Context, userIDStr string, file io.Reader, size int64, contentType string) (string, error) {
	const op = "users.service.UploadClientAvatar"

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("[%s]: %w", op, domain.ErrInvalidInput)
	}

	client, err := s.repo.GetClientByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("[%s]: get client: %w", op, err)
	}

	objectName, err := s.storage.Upload(ctx, minioPkg.ObjectInfo{
		BucketName:  usersBucket,
		Reader:      file,
		Size:        size,
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("[%s]: upload: %w", op, err)
	}

	if err := s.repo.UpdateClientAvatarURL(ctx, client.ID, objectName); err != nil {
		_ = s.storage.Remove(ctx, usersBucket, objectName)
		return "", fmt.Errorf("[%s]: update avatar url: %w", op, err)
	}

	if client.AvatarURL != nil {
		_ = s.storage.Remove(ctx, usersBucket, *client.AvatarURL)
	}

	return s.storage.BuildObjectURL(usersBucket, objectName), nil
}

func (s *Service) mapClientToDTO(c *model.Client) dto.Client {
	d := dto.Client{
		ID:        c.ID.String(),
		UserID:    c.UserID.String(),
		FirstName: c.FirstName,
		LastName:  c.LastName,
		Phone:     c.Phone,
	}
	if c.AvatarURL != nil {
		url := s.storage.BuildObjectURL(usersBucket, *c.AvatarURL)
		d.AvatarURL = &url
	}
	return d
}
