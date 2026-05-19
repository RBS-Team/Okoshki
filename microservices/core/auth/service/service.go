package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
)

type UserSaver interface {
	CreateUser(ctx context.Context, user model.User) error
	DeleteUserByID(ctx context.Context, id uuid.UUID) error
}

type UserProvider interface {
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.User, error)
}

type Service interface {
	CreateUser(ctx context.Context, email, password, role string) (uuid.UUID, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	DeleteUserByID(ctx context.Context, id uuid.UUID) error
}

type service struct {
	usrSaver    UserSaver
	usrProvider UserProvider
}

func New(userSaver UserSaver, userProvider UserProvider) Service {
	return &service{
		usrSaver:    userSaver,
		usrProvider: userProvider,
	}
}
