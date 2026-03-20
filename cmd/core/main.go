package main

import (
	"context"
	"log"

	"github.com/RBS-Team/Okoshki/microservices/core/app"
)

func main() {
	application, err := app.NewApp(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application stopped with error: %v", err)
	}
}