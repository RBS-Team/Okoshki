package app

import (
	"log/slog"
	"time"

	"google.golang.org/grpc"

	auth "github.com/RBS-Team/Okoshki/daemon/sso/delivery/grpc"
	"github.com/RBS-Team/Okoshki/internal/server"
)

type App struct {
	GRPCServer *server.GRPCServer
}

func New(log *slog.Logger, grpcPort int, tokenTTL time.Duration) *App {
	// TODO: инициализация хранилища (repository)

	// TODO: инициализация auth сервиса

	grpcServer := server.NewGRPCServer(log, grpcPort, func(s *grpc.Server) { auth.Register(s, authService) })
	return &App{
		GRPCServer: grpcServer,
	}
}
