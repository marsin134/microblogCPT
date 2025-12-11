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
	"time"
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
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/tables", tablesHandler(db))

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

// main page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Микроблог API</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 40px; }
			h1 { color: #333; }
			.endpoint { background: #f5f5f5; padding: 10px; margin: 5px 0; border-radius: 5px; }
			.method { font-weight: bold; color: #007bff; }
			.path { color: #28a745; }
		</style>
	</head>
	<body>
		<h1>Микроблог API</h1>
		<p>Добро пожаловать в микроблог платформу!</p>
		
		<h2>Аутентификация</h2>
		<div class="endpoint"><span class="method">POST</span> <span class="path">/api/auth/register</span> - Регистрация</div>
		<div class="endpoint"><span class="method">POST</span> <span class="path">/api/auth/login</span> - Вход</div>
		<div class="endpoint"><span class="method">POST</span> <span class="path">/api/auth/refresh-token</span> - Обновление токена</div>
		
		<h2>Пользователи</h2>
		<div class="endpoint"><span class="method">GET</span> <span class="path">/api/me</span> - Текущий пользователь</div>
		<div class="endpoint"><span class="method">GET</span> <span class="path">/api/user/{id}</span> - Пользователь по ID</div>
		
		<h2>Посты</h2>
		<div class="endpoint"><span class="method">GET</span> <span class="path">/api/posts</span> - Все посты</div>
		<div class="endpoint"><span class="method">POST</span> <span class="path">/api/posts</span> - Создать пост</div>
		<div class="endpoint"><span class="method">PUT</span> <span class="path">/api/posts/{id}</span> - Обновить пост</div>
		<div class="endpoint"><span class="method">PATCH</span> <span class="path">/api/posts/{id}/status</span> - Публикация поста</div>
		
		<h2>Изображения</h2>
		<div class="endpoint"><span class="method">POST</span> <span class="path">/api/posts/{id}/images</span> - Добавить изображение</div>
		<div class="endpoint"><span class="method">DELETE</span> <span class="path">/api/posts/{id}/images/{imageId}</span> - Удалить изображение</div>
		
		<h2>Система</h2>
		<div class="endpoint"><span class="method">GET</span> <span class="path">/health</span> - Статус сервера</div>
		<div class="endpoint"><span class="method">GET</span> <span class="path">/tables</span> - Таблицы БД</div>
		
		<hr>
		<p><strong>Для работы с API используйте Bearer токен:</strong> Authorization: Bearer YOUR_TOKEN</p>
		<p><strong>Роли:</strong> Author (может создавать посты), Reader (только чтение)</p>
	</body>
	</html>
	`)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "ok", "service": "microblog", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
}

func tablesHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var count int
		err := db.Get(&count, `
			SELECT COUNT(*) 
			FROM information_schema.tables 
			WHERE table_schema = 'public'
		`)

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"tables_count": %d}`, count)
	}
}
