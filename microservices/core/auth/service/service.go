package service

import (
	"context"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/google/uuid"
)

// Это интерйфесы для работы с базой данных.
// Этот интерфейс на запись
type UserSaver interface {
	CreateUser(ctx context.Context, user model.User) error
}

// Этот интерфейс на чтение
type UserProvider interface {
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]model.User, error)
	// Мб check role сделать ?
	// IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AuthService struct {
	usrSaver    UserSaver
	usrProvider UserProvider
}

// New returns a new instance of the Auth service
func NewAuthService(userSaver UserSaver, userProvider UserProvider) *AuthService {
	return &AuthService{
		usrSaver:    userSaver,
		usrProvider: userProvider,
	}
}
