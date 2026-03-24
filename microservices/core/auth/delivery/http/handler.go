package http

import (
	"context"

	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/pkg/csrfmanager"
	"github.com/RBS-Team/Okoshki/pkg/jwtmanager"
)

//go:generate mockgen -destination=../../../../mocks/auth/service/service_mock.go -package=mock_auth_service spotify/microservices/auth/delivery/http IService,CSRFManager
type IService interface {
	RegisterNewUser(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
}

type AuthHandler struct {
	service    IService
	jwtManager *jwtmanager.Manager
	csrfManager *csrfmanager.Manager
}

func NewHandler(svc IService, jwtManager *jwtmanager.Manager) *AuthHandler {
	return &AuthHandler{
		service:    svc,
		jwtManager: jwtManager,
	}
}
