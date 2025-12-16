package main

import (
	"fmt"
	"log"
	"microblogCPT/cmd/app"
	"microblogCPT/internal/config"
	"microblogCPT/internal/database"
	handlers "microblogCPT/internal/handler"
	"microblogCPT/internal/middleware"
	"microblogCPT/internal/service"
	"net/http"
)

func main() {
	// setting up config
	cfg := config.LoadConfig()

	if cfg.JWTSecretKey == "" {
		log.Fatal("JWT_SECRET_KEY не установлен в .env файле")
	}

	db, repo, services := app.App(cfg)
	defer database.MethodsDB.CloseDB(db)

	handler := handlers.NewHandlers(repo, services, cfg)

	mux := service.CreateMux()

	// setting up routes
	mux.Mux.HandleFunc("/", handlers.HomeHandler)
	mux.Mux.HandleFunc("/health", handlers.HealthHandler)
	mux.Mux.HandleFunc("/tables", handler.TablesHandler)

	mux.Mux.HandleFunc("/api/auth/register", handler.Register)
	mux.Mux.HandleFunc("/api/auth/login", handler.Login)
	mux.Mux.HandleFunc("/api/auth/refresh-token", handler.RefreshToken)

	mux.Mux.HandleFunc("/api/me", handler.GetCurrentUser)
	mux.Mux.HandleFunc("/api/user/", handler.GetUser)

	mux.Mux.HandleFunc("/api/posts", handler.GetPosts)
	mux.Mux.HandleFunc("/api/posts/", handler.CreatePost)
	mux.Mux.HandleFunc("/api/posts//status", handler.PublishPost)

	mux.Mux.HandleFunc("/api/posts//images", handler.AddedImage)
	mux.Mux.HandleFunc("/api/posts//images/", handler.DeleteImage)

	handlerChain := middleware.Chain(
		mux.Mux,
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
