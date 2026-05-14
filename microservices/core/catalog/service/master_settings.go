package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
)

// Дефолты для мастера, у которого ещё нет строки настроек.
const (
	DefaultSlotStepMinutes = 30
	DefaultLeadTimeMinutes = 0
)

// AllowedSlotSteps — единственно допустимые значения шага сетки.
var AllowedSlotSteps = []int{5, 10, 15, 20, 30, 60}

// GetMasterSettings возвращает настройки мастера. Если строки нет — отдаёт дефолты.
// Дефолты — внутренняя константа, не из БД.
func (s *Service) GetMasterSettings(ctx context.Context, masterID uuid.UUID) (*dto.MasterSettings, error) {
	const op = "catalog.service.GetMasterSettings"

	settings, err := s.repo.GetMasterSettings(ctx, masterID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return &dto.MasterSettings{
				MasterID:        masterID.String(),
				SlotStepMinutes: DefaultSlotStepMinutes,
				LeadTimeMinutes: DefaultLeadTimeMinutes,
			}, nil
		}
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	return &dto.MasterSettings{
		MasterID:        settings.MasterID.String(),
		SlotStepMinutes: settings.SlotStepMinutes,
		LeadTimeMinutes: settings.LeadTimeMinutes,
	}, nil
}

// UpsertMasterSettings создаёт строку настроек или обновляет указанные поля.
// nil-поля в req не трогают существующие значения.
func (s *Service) UpsertMasterSettings(ctx context.Context, masterID uuid.UUID, req dto.UpsertMasterSettingsRequest) error {
	const op = "catalog.service.UpsertMasterSettings"

	if req.SlotStepMinutes == nil && req.LeadTimeMinutes == nil {
		return fmt.Errorf("[%s]: nothing to update: %w", op, domain.ErrInvalidInput)
	}

	if req.SlotStepMinutes != nil && !isAllowedSlotStep(*req.SlotStepMinutes) {
		return fmt.Errorf("[%s]: slot_step_minutes must be one of %v: %w", op, AllowedSlotSteps, domain.ErrInvalidInput)
	}

	if req.LeadTimeMinutes != nil && *req.LeadTimeMinutes < 0 {
		return fmt.Errorf("[%s]: lead_time_minutes must be >= 0: %w", op, domain.ErrInvalidInput)
	}

	if err := s.repo.UpsertMasterSettings(ctx, masterID, req.SlotStepMinutes, req.LeadTimeMinutes); err != nil {
		return fmt.Errorf("[%s]: %w", op, err)
	}

	return nil
}

func isAllowedSlotStep(v int) bool {
	for _, allowed := range AllowedSlotSteps {
		if v == allowed {
			return true
		}
	}
	return false
}
