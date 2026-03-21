package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/server"
	"github.com/RBS-Team/Okoshki/pkg/jwtmanager"

	httpDelivery "github.com/RBS-Team/Okoshki/microservices/core/auth/delivery/http"
	userRepo "github.com/RBS-Team/Okoshki/microservices/core/auth/repository/postgres"
	userService "github.com/RBS-Team/Okoshki/microservices/core/auth/service"
	"github.com/RBS-Team/Okoshki/pkg/logger"
	"github.com/RBS-Team/Okoshki/pkg/postgres"
	"github.com/gorilla/mux"
)

type App struct {
	cfg        *Config
	logger     logger.Logger
	db         *sql.DB
	httpServer *server.Server
}

func NewApp(ctx context.Context, configPath string) (*App, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	appLogger, err := logger.New(cfg.Auth.Logger.Level, cfg.Auth.Logger.Mode)
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}
	appLogger.Infof("Logger initialized for Auth service")

	db, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		appLogger.Errorf("failed to connect to db: %v", err)
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	appLogger.Infof("Database connection established")

	userRepository := userRepo.NewUserRepository(db)
	userService := userService.NewAuthService(userRepository, userRepository)

	jwtManager := jwtmanager.NewManager(cfg.Auth.HTTP.Auth.JWT.SecretKey, cfg.Auth.HTTP.Auth.JWT.AccessTokenTTL)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	httpHandler := httpDelivery.NewHandler(userService, jwtManager)

	router := mux.NewRouter()

	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.RequestLoggerMiddleware(appLogger))


	public := api.PathPrefix("").Subrouter()
	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.AuthMiddleware)

	httpHandler.RegisterRoutes(public, protected)

	httpServer := server.NewHTTPServer(&cfg.Auth.HTTP, router, appLogger)

	return &App{
		cfg:        cfg,
		logger:     appLogger,
		db:         db,
		httpServer: httpServer,
	}, nil
}
