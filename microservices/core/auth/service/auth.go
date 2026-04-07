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

func (a *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	const op = "service.Login"
	// log := a.log.With(slog.String("op", op))
	// log.Info("attempting to login user")

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

func (a *AuthService) RegisterNewUser(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	const op = "auth.RegisterNewUser"
	// log := a.log.With(slog.String("op", op))
	// log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("[%s]: failed to hash password: %w", op, err)
	}
	user := model.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(passHash),
		Role:         req.Role,
		AvatarURL:    "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := a.usrSaver.CreateUser(ctx, user); err != nil {
		return nil, mapRepositoryError(err)
	}
	return &dto.RegisterResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (a *AuthService) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	panic("not implemented")
}

func (a *AuthService) GetUsersInfo(ctx context.Context, ids []uuid.UUID) ([]dto.UserInfo, error) {
	const op = "auth.service.GetUsersInfo"

	users, err := a.usrProvider.GetUsersByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("[%s]: %w", op, err)
	}

	result := make([]dto.UserInfo, 0, len(users))
	for _, u := range users {
		result = append(result, dto.UserInfo{
			ID:        u.ID.String(),
			Email:     u.Email,
			AvatarURL: u.AvatarURL,
		})
	}
	return result, nil
}
