package http

import (
	"context"

	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/pkg/jwtmanager"
)

//go:generate mockgen -destination=../../../../mocks/auth/service/service_mock.go -package=mock_auth_service github.com/RBS-Team/Okoshki/microservices/core/auth/delivery/http IService
type IService interface {
	RegisterNewUser(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	DeleteUser(ctx context.Context, userID string) error
}

// MasterCreator создаёт профиль мастера в catalog домене.
// Реализуется через адаптер в app.go, чтобы auth не импортировал catalog.
type MasterCreator interface {
	CreateMasterProfile(ctx context.Context, userIDStr, name string, bio *string, timezone string, lat, lon *float64) (masterID string, err error)
}

type AuthHandler struct {
	service       IService
	jwtManager    *jwtmanager.Manager
	masterCreator MasterCreator
}

func NewHandler(svc IService, jwtManager *jwtmanager.Manager, masterCreator MasterCreator) *AuthHandler {
	return &AuthHandler{
		service:       svc,
		jwtManager:    jwtManager,
		masterCreator: masterCreator,
	}
}
