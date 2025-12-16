package app

import (
	"log"
	"microblogCPT/internal/config"
	"microblogCPT/internal/database"
	"microblogCPT/internal/repository"
	"microblogCPT/internal/service"
	"microblogCPT/internal/storage"
)

func App(cfg *config.Config) (*database.DB, *repository.Repository, *service.Service) {
	// connection DB
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}

	// connection MinIO
	minioClient, err := storage.NewMinIOClient(cfg)
	if err != nil {
		log.Fatalf("Не удалось инициализировать MinIO: %v", err)
	}

	// enabling dependencies
	repo := repository.NewRepository(db.DB)

	services := service.NewService(repo, cfg, minioClient)

	return db, repo, services
}
