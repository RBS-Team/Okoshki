package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/users/dto"
)

func (s *Service) RegisterMaster(ctx context.Context, req dto.RegisterMasterRequest) (*dto.RegisterMasterResponse, error) {
	const op = "users.service.RegisterMaster"

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: invalid category_id: %w", op, ErrInvalidInput)
	}

	tz := req.Timezone
	if tz == "" {
		tz = "Europe/Moscow"
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, ErrInvalidTimezone)
	}

	userID, err := s.auth.CreateUser(ctx, req.Email, req.Password, string(model.RoleMaster))
	if err != nil {
		return nil, fmt.Errorf("[%s]: create account: %w", op, err)
	}

	master := model.Master{
		ID:           uuid.New(),
		UserID:       userID,
		CategoryID:   categoryID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Address:      req.Address,
		City:         req.City,
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

	if err := s.repo.CreateMaster(ctx, master); err != nil {
		_ = s.auth.DeleteUserByID(ctx, userID)
		return nil, fmt.Errorf("[%s]: create master profile: %w", op, mapError(err))
	}

	return &dto.RegisterMasterResponse{
		UserID:   userID.String(),
		MasterID: master.ID.String(),
		Role:     string(model.RoleMaster),
	}, nil
}

func (s *Service) GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*dto.Master, error) {
	const op = "catalog.service.GetMasterByUserID"

	masterModel, err := s.repo.GetMasterByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get master by user id: %w", op, mapError(err))
	}

	return s.mapMasterToDTO(masterModel), nil
}

func (s *Service) GetMastersByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]dto.Master, error) {
	const op = "users.service.GetMastersByCategory"

	masterModels, err := s.repo.GetMastersByCategoryID(ctx, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to get masters by category: %w", op, mapError(err))
	}

	if len(masterModels) == 0 {
		return []dto.Master{}, nil
	}

	masterDTOs := make([]dto.Master, 0, len(masterModels))
	for i := range masterModels {
		masterDTOs = append(masterDTOs, *s.mapMasterToDTO(&masterModels[i]))
	}

	return masterDTOs, nil
}

// GetMasterByID возвращает профиль мастера по ID как DTO (для HTTP-хендлеров).
func (s *Service) GetMasterByID(ctx context.Context, id uuid.UUID) (*dto.Master, error) {
	const op = "users.service.GetMasterByID"

	m, err := s.repo.GetMasterByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapError(err))
	}

	return s.mapMasterToDTO(m), nil
}

// GetAllMasters возвращает страницу мастеров как DTO (для HTTP-хендлеров).
func (s *Service) GetAllMasters(ctx context.Context, limit, offset uint64) ([]dto.Master, error) {
	const op = "users.service.GetAllMasters"

	masters, err := s.repo.GetAllMasters(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, mapError(err))
	}

	if len(masters) == 0 {
		return []dto.Master{}, nil
	}

	result := make([]dto.Master, 0, len(masters))
	for i := range masters {
		result = append(result, *s.mapMasterToDTO(&masters[i]))
	}

	return result, nil
}

// Find* методы возвращают model.Master и реализуют catalog/service.MasterProvider (duck typing).
// Названия отличаются от Get*, чтобы не конфликтовать с DTO-возвращающими методами.
func (s *Service) GetMastersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Master, error) {
	return s.repo.GetMastersByIDs(ctx, ids)
}

const usersBucket = "okoshki-users"

func (s *Service) mapMasterToDTO(m *model.Master) *dto.Master {
	d := &dto.Master{
		ID:          m.ID.String(),
		UserID:      m.UserID.String(),
		CategoryID:  m.CategoryID.String(),
		FirstName:   m.FirstName,
		LastName:    m.LastName,
		Phone:       m.Phone,
		Address:     m.Address,
		City:        m.City,
		Bio:         m.Bio,
		Timezone:    m.Timezone,
		Lat:         m.Lat,
		Lon:         m.Lon,
		Rating:      m.Rating,
		ReviewCount: m.ReviewCount,
	}
	if m.AvatarURL != nil {
		url := s.storage.BuildObjectURL(usersBucket, *m.AvatarURL)
		d.AvatarURL = &url
	}
	return d
}
