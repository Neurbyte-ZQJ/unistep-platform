package main

import (
	"log"

	"unistep-platform/backend/internal/config"
	"unistep-platform/backend/internal/database"
	"unistep-platform/backend/internal/router"
	"unistep-platform/backend/internal/storage"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := database.Seed(db); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	uploader, err := storage.NewClient(storage.Config{
		Endpoint:        cfg.MinioEndpoint,
		AccessKey:       cfg.MinioAccessKey,
		SecretKey:       cfg.MinioSecretKey,
		Bucket:          cfg.MinioBucket,
		UseSSL:          cfg.MinioUseSSL,
		PublicURLPrefix: cfg.MinioPublicURL,
	})
	if err != nil {
		log.Fatalf("failed to init minio client: %v", err)
	}

	r := router.New(cfg, db, uploader)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
