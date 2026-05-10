package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/RBS-Team/Okoshki/docs" // сгенерированная документация
	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/server"
	userHtpp "github.com/RBS-Team/Okoshki/microservices/core/auth/delivery/http"
	userRepo "github.com/RBS-Team/Okoshki/microservices/core/auth/repository/postgres"
	userService "github.com/RBS-Team/Okoshki/microservices/core/auth/service"
	bookingHttp "github.com/RBS-Team/Okoshki/microservices/core/booking/delivery/http"
	bookingRepo "github.com/RBS-Team/Okoshki/microservices/core/booking/repository/postgres"
	bookingService "github.com/RBS-Team/Okoshki/microservices/core/booking/service"
	catalogHttp "github.com/RBS-Team/Okoshki/microservices/core/catalog/delivery/http"
	catalogDto "github.com/RBS-Team/Okoshki/microservices/core/catalog/dto"
	catalogRepo "github.com/RBS-Team/Okoshki/microservices/core/catalog/repository/postgres"
	catalogService "github.com/RBS-Team/Okoshki/microservices/core/catalog/service"
	"github.com/RBS-Team/Okoshki/pkg/jwtmanager"
	"github.com/RBS-Team/Okoshki/pkg/logger"
	minioPkg "github.com/RBS-Team/Okoshki/pkg/minio"
	"github.com/RBS-Team/Okoshki/pkg/postgres"
)

// masterCreatorAdapter реализует userHtpp.MasterCreator через catalog.Service.
// Изолирует auth домен от прямой зависимости на catalog.
type masterCreatorAdapter struct {
	svc *catalogService.Service
}

func (a *masterCreatorAdapter) CreateMasterProfile(ctx context.Context, userIDStr, name string, bio *string, timezone string, lat, lon *float64) (string, error) {
	master, err := a.svc.CreateMaster(ctx, userIDStr, catalogDto.CreateMasterRequest{
		Name:     name,
		Bio:      bio,
		Timezone: timezone,
		Lat:      lat,
		Lon:      lon,
	})
	if err != nil {
		return "", err
	}
	return master.ID, nil
}

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

	jwtManager := jwtmanager.NewManager(cfg.Auth.HTTP.Auth.JWT.SecretKey, cfg.Auth.HTTP.Auth.JWT.AccessTokenTTL)

	userRepository := userRepo.NewUserRepository(db)
	userService := userService.NewAuthService(userRepository, userRepository)

	minioClient, err := minioPkg.New(cfg.Minio)
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	catalogRepository := catalogRepo.New(db)
	catalogSvc := catalogService.New(catalogRepository, minioClient)
	catalogHandler := catalogHttp.NewHandler(catalogSvc)

	userHandler := userHtpp.NewHandler(userService, jwtManager, &masterCreatorAdapter{svc: catalogSvc})

	bookingRepository := bookingRepo.New(db)
	bookingSvc := bookingService.New(bookingRepository, catalogSvc, userService)
	bookingHandler := bookingHttp.NewHandler(bookingSvc)

	requestLoggerMiddleware := middleware.RequestLoggerMiddleware(appLogger)
	corsMiddleware := middleware.CORS(cfg.Auth.HTTP.CORS)

	csrfMiddleware := csrf.Protect(
		[]byte(cfg.Auth.HTTP.Auth.CSRF.SecretKey),
		csrf.Secure(cfg.Auth.HTTP.Auth.CSRF.Secure),
		csrf.TrustedOrigins(cfg.Auth.HTTP.Auth.CSRF.TrustedOrigins),
	)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	router := mux.NewRouter()

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	api := router.PathPrefix("/api/v1").Subrouter()

	api.Use(requestLoggerMiddleware)
	api.Use(corsMiddleware)

	public := api.PathPrefix("").Subrouter()

	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.AuthMiddleware)

	csrfProtected := protected.PathPrefix("").Subrouter()
	// csrfProtected.Use(csrfMiddleware)
	_ = csrfMiddleware
	catalogHandler.RegisterRoutes(public, protected, csrfProtected)
	userHandler.RegisterRoutes(public, protected, csrfProtected)
	bookingHandler.RegisterRoutes(public, protected, csrfProtected)

	httpServer := server.NewHTTPServer(&cfg.Auth.HTTP, router, appLogger)

	return &App{
		cfg:        cfg,
		logger:     appLogger,
		db:         db,
		httpServer: httpServer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	serverErrors := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.httpServer.Run(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("http server error: %w", err)
		}
	}()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	if err := a.grpcServer.Run(); err != nil {
	// 		serverErrors <- fmt.Errorf("grpc server error: %w", err)
	// 	}
	// }()

	a.logger.Infof("Auth microservice is running...")

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server run failed: %w", err)
	case <-ctx.Done():
		a.logger.Infof("shutting down servers due to context cancellation...")
	}

	if err := a.Stop(); err != nil {
		return fmt.Errorf("failed to gracefully stop application: %w", err)
	}

	wg.Wait()
	a.logger.Infof("All servers stopped, application is shutting down.")
	return nil
}

func (a *App) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.Auth.HTTP.ShutdownTimeout)
	defer cancel()

	errHTTP := a.httpServer.Shutdown(shutdownCtx)
	// a.grpcServer.Stop()

	errDB := a.db.Close()

	errLog := a.logger.Sync()
	if errLog != nil {
		log.Printf("ERROR: failed to sync logger: %v", errLog)
	}

	if errHTTP != nil || errDB != nil {
		return fmt.Errorf("shutdown errors: http=%v, db=%v", errHTTP, errDB)
	}

	a.logger.Infof("Application stopped gracefully.")
	return nil
}
