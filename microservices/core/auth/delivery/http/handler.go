package http

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
	"github.com/RBS-Team/Okoshki/pkg/jwtmanager"
)

//go:generate mockgen -destination=../../../../mocks/auth/service/service_mock.go -package=mock_auth_service github.com/RBS-Team/Okoshki/microservices/core/auth/delivery/http IService
type IService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	DeleteUserByID(ctx context.Context, userID uuid.UUID) error
}

type Handler interface {
	RegisterRoutes(public, protected, csrfProtected *mux.Router)
}

type handler struct {
	service    IService
	jwtManager *jwtmanager.Manager
}

func NewHandler(svc IService, jwtManager *jwtmanager.Manager) Handler {
	return &handler{
		service:    svc,
		jwtManager: jwtManager,
	}
}
