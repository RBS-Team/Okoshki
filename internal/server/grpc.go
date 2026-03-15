package server

import (
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type GRPCServer struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func NewGRPCServer(log *slog.Logger, port int, registerServices func(server *grpc.Server)) *GRPCServer {
	gRPCServer := grpc.NewServer()

	if registerServices != nil {
		registerServices(gRPCServer)
	}

	return &GRPCServer{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (s *GRPCServer) MustRun() {
	if err := s.Run(); err != nil {
		panic(err)
	}
}

func (s *GRPCServer) Run() error {
	// константа op это метка текущей функции, чтобы понимать, откуда пришла ошибка, чтобы потом в логах было проще понять, где поломались
	const op = "grpcserver.Run"
	log := s.log.With(
		slog.String("op", op),
		slog.Int("port", s.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	log.Info("gRPC server is running", slog.String("addr", l.Addr().String()))

	if err := s.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (s *GRPCServer) Stop() {
	const op = "grpcserver.Stop"

	s.log.With(slog.String("op", op))

	s.gRPCServer.GracefulStop()
}
