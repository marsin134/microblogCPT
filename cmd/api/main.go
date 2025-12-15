package main

import (
	"fmt"
	"log"
	"microblogCPT/internal/config"
	"microblogCPT/internal/database"
	handlers "microblogCPT/internal/handler"
	"microblogCPT/internal/middleware"
	"microblogCPT/internal/repository"
	"microblogCPT/internal/service"
	"microblogCPT/internal/storage"
	"net/http"
)

func main() {
	// setting up config
	cfg := config.LoadConfig()

	if cfg.JWTSecretKey == "" {
		log.Fatal("JWT_SECRET_KEY не установлен в .env файле")
	}

	// connection DB
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer database.MethodsDB.CloseDB(db)

	// connection MinIO
	minioClient, err := storage.NewMinIOClient(cfg)
	if err != nil {
		log.Fatalf("Не удалось инициализировать MinIO: %v", err)
	}

	// enabling dependencies
	repo := repository.NewRepository(db.DB)

	services := service.NewService(repo, cfg, minioClient)

	handler := handlers.NewHandlers(repo, services, cfg)

	mux := http.NewServeMux()

	// setting up routes
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/tables", handler.TablesHandler)

	mux.HandleFunc("/api/auth/register", handler.Register)
	mux.HandleFunc("/api/auth/login", handler.Login)
	mux.HandleFunc("/api/auth/refresh-token", handler.RefreshToken)

	mux.HandleFunc("/api/me", handler.GetCurrentUser)
	mux.HandleFunc("/api/user/", handler.GetUser)

	mux.HandleFunc("/api/posts", handler.GetPosts)
	mux.HandleFunc("/api/posts/", handler.CreatePost)
	mux.HandleFunc("/api/posts//status", handler.PublishPost)

	mux.HandleFunc("/api/posts//images", handler.AddedImage)
	mux.HandleFunc("/api/posts//images/", handler.DeleteImage)

	handlerChain := middleware.Chain(
		mux,
		middleware.LoggingMiddleware,
		middleware.CORSMiddleware,
		middleware.AuthMiddleware(cfg),
	)

	// Starting the server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	fmt.Printf("Сервер запущен на %s\n", addr)
	fmt.Printf("База данных: %s\n", cfg.DB.DbNAME)
	fmt.Printf("Адресс: http://localhost:8080/\n")

	if err := http.ListenAndServe(addr, handlerChain); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
