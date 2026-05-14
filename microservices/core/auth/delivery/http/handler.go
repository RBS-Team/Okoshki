package http

import (
	"context"

	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/pkg/jwtmanager"
	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../../../mocks/auth/service/service_mock.go -package=mock_auth_service github.com/RBS-Team/Okoshki/microservices/core/auth/delivery/http IService
type IService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	DeleteUserByID(ctx context.Context, userID uuid.UUID) error
}

type AuthHandler struct {
	service    IService
	jwtManager *jwtmanager.Manager
}

func NewHandler(svc IService, jwtManager *jwtmanager.Manager) *AuthHandler {
	return &AuthHandler{
		service:    svc,
		jwtManager: jwtManager,
	}
}
