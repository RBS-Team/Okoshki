package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/RBS-Team/Okoshki/pkg/logger"
	"github.com/RBS-Team/Okoshki/pkg/postgres"

	catalogHttp "github.com/RBS-Team/Okoshki/microservices/core/catalog/delivery/http"
	catalogRepo "github.com/RBS-Team/Okoshki/microservices/core/catalog/repository/postgres"
	catalogSvc "github.com/RBS-Team/Okoshki/microservices/core/catalog/service"
)

type App struct {
	server *http.Server
	logger logger.Logger
}

func NewApp(ctx context.Context) (*App, error) {
	log, err := logger.New("debug", "dev")
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}
	log.Infof("Logger initialized")

	dbConfig := postgres.Config{
		Host:            "localhost",
		Port:            5432,
		User:            "dev_user",
		Password:        "dev_password",
		DBName:          "okoshki_db",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 0,
	}

	db, err := postgres.New(ctx, dbConfig)
	if err != nil {
		log.Errorf("Database connection failed: %v", err)
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	log.Infof("Connected to PostgreSQL okoshki_db")

	catalogRepository := catalogRepo.New(db)
	catalogService := catalogSvc.New(catalogRepository)
	catalogHandler := catalogHttp.NewHandler(catalogService)

	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	catalogHandler.RegisterRoutes(apiRouter)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	return &App{
		server: srv,
		logger: log,
	}, nil
}

func (a *App) Run() error {
	a.logger.Infof("Core Service is running on http://localhost%s", a.server.Addr)
	return a.server.ListenAndServe()
}
