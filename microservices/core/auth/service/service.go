package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/RBS-Team/Okoshki/internal/model"
)

type UserSaver interface {
	CreateUser(ctx context.Context, user model.User) error
	DeleteUserByID(ctx context.Context, id uuid.UUID) error
}

type UserProvider interface {
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.User, error)
}

type AuthService struct {
	usrSaver    UserSaver
	usrProvider UserProvider
}

func New(userSaver UserSaver, userProvider UserProvider) *AuthService {
	return &AuthService{
		usrSaver:    userSaver,
		usrProvider: userProvider,
	}
}
