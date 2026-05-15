package http

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/microservices/core/users/dto"
	"github.com/RBS-Team/Okoshki/pkg/jwtmanager"
)

//go:generate mockgen -destination=../../../../mocks/users/service/service_mock.go -package=mock_users_service github.com/RBS-Team/Okoshki/microservices/core/users/delivery/http IService
type IService interface {
	RegisterMaster(ctx context.Context, req dto.RegisterMasterRequest) (*dto.RegisterMasterResponse, error)
	RegisterClient(ctx context.Context, req dto.RegisterClientRequest) (*dto.RegisterClientResponse, error)

	GetMasterByUserID(ctx context.Context, userID uuid.UUID) (*dto.Master, error)
	GetMasterByID(ctx context.Context, id uuid.UUID) (*dto.Master, error)
	GetAllMasters(ctx context.Context, limit, offset uint64) ([]dto.Master, error)
	GetMastersByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset uint64) ([]dto.Master, error)

	GetClientByUserID(ctx context.Context, userID uuid.UUID) (*dto.Client, error)

	UploadMasterAvatar(ctx context.Context, userIDStr, masterIDStr string, file io.Reader, size int64, contentType string) (string, error)
	UploadClientAvatar(ctx context.Context, userIDStr string, file io.Reader, size int64, contentType string) (string, error)

	UploadPortfolioPhotos(ctx context.Context, userIDStr, masterIDStr string, files []dto.FileUpload) ([]dto.PortfolioPhoto, error)
	GetPortfolioPhotos(ctx context.Context, masterIDStr string) ([]dto.PortfolioPhoto, error)
	DeletePortfolioPhoto(ctx context.Context, userIDStr, masterIDStr, photoIDStr string) error
}

type Handler struct {
	service    IService
	jwtManager *jwtmanager.Manager
}

func NewHandler(service IService, jwtManager *jwtmanager.Manager) *Handler {
	return &Handler{
		service:    service,
		jwtManager: jwtManager,
	}
}
