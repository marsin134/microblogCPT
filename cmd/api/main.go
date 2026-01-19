package main

import (
	"fmt"
	"log"
	"microblogCPT/cmd/app"
	"microblogCPT/internal/config"
	"microblogCPT/internal/database"
	"net/http"
)

func main() {
	// setting up config
	cfg := config.LoadConfig()

	if cfg.JWTSecretKey == "" {
		log.Fatal("JWT_SECRET_KEY не установлен в .env файле")
	}

	// connection DB and services
	db, services := app.App(cfg)
	defer database.MethodsDB.CloseDB(db)

	handlerChain := app.InitializationHandlers(services, cfg)

	// Starting the server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	fmt.Printf("Сервер запущен на %s\n", addr)
	fmt.Printf("База данных: %s\n", cfg.DB.DbNAME)
	fmt.Printf("Адресс: http://localhost:8080/\n")

	if err := http.ListenAndServe(addr, handlerChain); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
