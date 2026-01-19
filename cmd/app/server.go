package app

import (
	"microblogCPT/internal/config"
	handlers "microblogCPT/internal/handler"
	"microblogCPT/internal/middleware"
	"microblogCPT/internal/service"
	"net/http"
)

func InitializationHandlers(services *service.Service, cfg *config.Config) http.Handler {
	handler := handlers.NewHandlers(services, cfg)

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

	return handlerChain
}
