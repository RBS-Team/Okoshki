// @title           Okoshki API
// @version         1.0
// @description     API для сервиса Okoshki (каталог мастеров и услуг)

// @host      localhost:8080
// @BasePath  /api/v1

package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/RBS-Team/Okoshki/microservices/core/app"
)

func main() {
	configPath := parseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	auth, err := app.NewApp(ctx, configPath)
	if err != nil {
		log.Fatalf("application init failed: %v", err)
	}

	if err := auth.Run(ctx); err != nil {
		log.Printf("application run failed: %v", err)
	}
}

func parseFlags() string {
	var configPath string
	flag.StringVar(&configPath, "f", "config", "path to config directory")
	flag.Parse()
	return configPath
}
