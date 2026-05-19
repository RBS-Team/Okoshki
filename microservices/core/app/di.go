package app

import (
	"context"
	"database/sql"

	authHttp "github.com/RBS-Team/Okoshki/microservices/core/auth/delivery/http"
	authRepo "github.com/RBS-Team/Okoshki/microservices/core/auth/repository/postgres"
	authService "github.com/RBS-Team/Okoshki/microservices/core/auth/service"
	bookingHttp "github.com/RBS-Team/Okoshki/microservices/core/booking/delivery/http"
	bookingRepo "github.com/RBS-Team/Okoshki/microservices/core/booking/repository/postgres"
	bookingService "github.com/RBS-Team/Okoshki/microservices/core/booking/service"
	catalogHttp "github.com/RBS-Team/Okoshki/microservices/core/catalog/delivery/http"
	catalogRepo "github.com/RBS-Team/Okoshki/microservices/core/catalog/repository/postgres"
	catalogService "github.com/RBS-Team/Okoshki/microservices/core/catalog/service"
	userHttp "github.com/RBS-Team/Okoshki/microservices/core/users/delivery/http"
	userRepo "github.com/RBS-Team/Okoshki/microservices/core/users/repository/postgres"
	userService "github.com/RBS-Team/Okoshki/microservices/core/users/service"
	"github.com/RBS-Team/Okoshki/pkg/jwtmanager"
	"github.com/RBS-Team/Okoshki/pkg/logger"
	minioPkg "github.com/RBS-Team/Okoshki/pkg/minio"
	"github.com/RBS-Team/Okoshki/pkg/postgres"
)

type diContainer struct {
	ctx    context.Context
	cfg    *Config
	logger logger.Logger

	// Инфраструктура
	db          *sql.DB
	jwtManager  *jwtmanager.Manager
	minioClient *minioPkg.Client

	// Репозитории
	authRepo    authRepo.Repository
	userRepo    userRepo.Repository
	catalogRepo catalogRepo.Repository
	bookingRepo bookingRepo.Repository

	// Сервисы
	authSvc    authService.Service
	userSvc    userService.Service
	catalogSvc catalogService.Service
	bookingSvc bookingService.Service

	// Хендлеры
	authHandler    authHttp.Handler
	userHandler    userHttp.Handler
	catalogHandler catalogHttp.Handler
	bookingHandler bookingHttp.Handler
}

func newDIContainer(ctx context.Context, cfg *Config, log logger.Logger) *diContainer {
	return &diContainer{ctx: ctx, cfg: cfg, logger: log}
}

// --- Инфраструктура ---

func (d *diContainer) DB() *sql.DB {
	if d.db == nil {
		db, err := postgres.New(d.ctx, d.cfg.DB)
		if err != nil {
			d.logger.Fatalf("failed to connect to db: %v", err)
		}
		d.db = db
		d.logger.Infof("POSTGRES CONNECTION established")
	}
	return d.db
}

func (d *diContainer) JWTManager() *jwtmanager.Manager {
	if d.jwtManager == nil {
		d.jwtManager = jwtmanager.NewManager(d.cfg.Auth.HTTP.Auth.JWT.SecretKey, d.cfg.Auth.HTTP.Auth.JWT.AccessTokenTTL)
		d.logger.Infof("JWT MANAGER created")
	}
	return d.jwtManager
}

func (d *diContainer) MinioClient() *minioPkg.Client {
	if d.minioClient == nil {
		client, err := minioPkg.New(d.cfg.Minio)
		if err != nil {
			d.logger.Fatalf("failed to init minio client: %v", err)
		}
		d.minioClient = client
		d.logger.Infof("MINIO CLIENT created")
	}
	return d.minioClient
}

// --- Репозитории ---

func (d *diContainer) AuthRepo() authRepo.Repository {
	if d.authRepo == nil {
		d.authRepo = authRepo.New(d.DB())
		d.logger.Infof("AUTH REPOSITORY created")
	}
	return d.authRepo
}

func (d *diContainer) UserRepo() userRepo.Repository {
	if d.userRepo == nil {
		d.userRepo = userRepo.New(d.DB())
		d.logger.Infof("USERS REPOSITORY created")
	}
	return d.userRepo
}

func (d *diContainer) CatalogRepo() catalogRepo.Repository {
	if d.catalogRepo == nil {
		d.catalogRepo = catalogRepo.New(d.DB())
		d.logger.Infof("CATALOG REPOSITORY created")
	}
	return d.catalogRepo
}

func (d *diContainer) BookingRepo() bookingRepo.Repository {
	if d.bookingRepo == nil {
		d.bookingRepo = bookingRepo.New(d.DB())
		d.logger.Infof("BOOKING REPOSITORY created")
	}
	return d.bookingRepo
}

// --- Сервисы ---

func (d *diContainer) AuthSvc() authService.Service {
	if d.authSvc == nil {
		d.authSvc = authService.New(d.AuthRepo(), d.AuthRepo())
		d.logger.Infof("AUTH SERVICE created")
	}
	return d.authSvc
}

func (d *diContainer) UserSvc() userService.Service {
	if d.userSvc == nil {
		d.userSvc = userService.New(d.AuthSvc(), d.UserRepo(), d.MinioClient())
		d.logger.Infof("USERS SERVICE created")
	}
	return d.userSvc
}

func (d *diContainer) CatalogSvc() catalogService.Service {
	if d.catalogSvc == nil {
		d.catalogSvc = catalogService.New(d.CatalogRepo(), d.UserSvc(), d.MinioClient())
		d.logger.Infof("CATALOG SERVICE created")
	}
	return d.catalogSvc
}

func (d *diContainer) BookingSvc() bookingService.Service {
	if d.bookingSvc == nil {
		d.bookingSvc = bookingService.New(d.BookingRepo(), d.CatalogSvc(), d.UserSvc())
		d.logger.Infof("BOOKING SERVICE created")
	}
	return d.bookingSvc
}

// --- Хендлеры ---

func (d *diContainer) AuthHandler() authHttp.Handler {
	if d.authHandler == nil {
		d.authHandler = authHttp.NewHandler(d.AuthSvc(), d.JWTManager())
		d.logger.Infof("AUTH HANDLER created")
	}
	return d.authHandler
}

func (d *diContainer) UserHandler() userHttp.Handler {
	if d.userHandler == nil {
		d.userHandler = userHttp.NewHandler(d.UserSvc(), d.JWTManager())
		d.logger.Infof("USERS HANDLER created")
	}
	return d.userHandler
}

func (d *diContainer) CatalogHandler() catalogHttp.Handler {
	if d.catalogHandler == nil {
		d.catalogHandler = catalogHttp.NewHandler(d.CatalogSvc())
		d.logger.Infof("CATALOG HANDLER created")
	}
	return d.catalogHandler
}

func (d *diContainer) BookingHandler() bookingHttp.Handler {
	if d.bookingHandler == nil {
		d.bookingHandler = bookingHttp.NewHandler(d.BookingSvc())
		d.logger.Infof("BOOKING HANDLER created")
	}
	return d.bookingHandler
}
