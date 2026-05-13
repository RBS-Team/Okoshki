package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/RBS-Team/Okoshki/internal/model"
	"github.com/RBS-Team/Okoshki/microservices/core/auth/dto"
)

// CreateAccount хеширует пароль, создаёт запись user и возвращает новый userID.
// Вызывается users/service при регистрации — auth не знает про профили.
func (a *AuthService) CreateUser(ctx context.Context, email, password, role string) (uuid.UUID, error) {
	const op = "auth.service.CreateAccount"

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("[%s]: hash password: %w", op, err)
	}

	user := model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(passHash),
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := a.usrSaver.CreateUser(ctx, user); err != nil {
		return uuid.Nil, mapRepositoryError(err)
	}

	return user.ID, nil
}

func (a *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	const op = "auth.service.Login"

	user, err := a.usrProvider.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("[%s]: invalid credentials: %w", op, ErrValidation)
	}

	return &dto.LoginResponse{
		ID:   user.ID.String(),
		Role: user.Role,
	}, nil
}

// DeleteUser удаляет учётку по ID. Используется как компенсирующая операция в users/service.
func (a *AuthService) DeleteUserByID(ctx context.Context, id uuid.UUID) error {
	const op = "auth.service.DeleteUser"


	return a.usrSaver.DeleteUserByID(ctx, id)
}

