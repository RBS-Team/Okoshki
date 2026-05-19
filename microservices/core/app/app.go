package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/RBS-Team/Okoshki/docs"
	"github.com/RBS-Team/Okoshki/internal/domain"
	"github.com/RBS-Team/Okoshki/internal/middleware"
	"github.com/RBS-Team/Okoshki/internal/server"
	"github.com/RBS-Team/Okoshki/pkg/closer"
	"github.com/RBS-Team/Okoshki/pkg/logger"
)

type App struct {
	cfg        *Config
	logger     logger.Logger
	closer     *closer.Closer
	di         *diContainer
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

	appCloser := closer.New(appLogger)
	appCloser.Add("logger", func(_ context.Context) error {
		return appLogger.Sync()
	})

	a := &App{
		cfg:    cfg,
		logger: appLogger,
		closer: appCloser,
		di:     newDIContainer(ctx, cfg, appLogger, appCloser),
	}

	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func() error{
		func() error { return a.initEnsureAdmin(ctx) },
		a.initHTTPServer,
	}
	for _, fn := range inits {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initEnsureAdmin(ctx context.Context) error {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		return nil
	}

	_, err := a.di.AuthSvc().CreateUser(ctx, email, password, "admin")
	if err == nil {
		a.logger.Infof("Admin account created: %s", email)
		return nil
	}
	if errors.Is(err, domain.ErrConflict) {
		return nil
	}
	a.logger.Warnf("Failed to create admin account: %v", err)
	return nil
}

func (a *App) initHTTPServer() error {
	requestLoggerMiddleware := middleware.RequestLoggerMiddleware(a.logger)
	corsMiddleware := middleware.CORS(a.cfg.Auth.HTTP.CORS)

	csrfMiddleware := csrf.Protect(
		[]byte(a.cfg.Auth.HTTP.Auth.CSRF.SecretKey),
		csrf.Secure(a.cfg.Auth.HTTP.Auth.CSRF.Secure),
		csrf.TrustedOrigins(a.cfg.Auth.HTTP.Auth.CSRF.TrustedOrigins),
	)
	authMiddleware := middleware.NewAuthMiddleware(a.di.JWTManager())

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

	masterCtx := middleware.MasterContext(a.di.UserSvc())

	a.di.CatalogHandler().RegisterRoutes(public, protected, csrfProtected, masterCtx)
	a.di.AuthHandler().RegisterRoutes(public, protected, csrfProtected)
	a.di.UserHandler().RegisterRoutes(public, protected, csrfProtected)
	a.di.BookingHandler().RegisterRoutes(public, protected, csrfProtected)

	a.httpServer = server.NewHTTPServer(&a.cfg.Auth.HTTP, router, a.logger)
	return nil
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

	a.logger.Infof("Core microservice is running...")

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

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Errorf("http server shutdown error: %v", err)
	}

	closerCtx, closerCancel := context.WithTimeout(context.Background(), a.cfg.Auth.HTTP.ShutdownTimeout)
	defer closerCancel()

	return a.closer.CloseAll(closerCtx)
}
