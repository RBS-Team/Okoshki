package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/RBS-Team/Okoshki/daemon/sso/domain/model"
	"github.com/RBS-Team/Okoshki/daemon/sso/repository/postgresql"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, PassHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (model.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (model.App, error)
}

// New returns a new instance of the Auth service
func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const op = "auth.Login"
	log := a.log.With(slog.String("op", op))
	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) {
			a.log.Warn("user not found", slog.String("error", err.Error()))

			return "", fmt.Errorf("%s:%w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s:%w", op, err)

	}
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(slog.String("op", op))
	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user registered")

	return id, nil

}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	panic("not implemented")
}
