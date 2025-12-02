package main

import (
	"fmt"
	"log"
	"microblogCPT/internal/config"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Микроблог запущен!\n")
		fmt.Fprintf(w, "Порт: %d\n", cfg.ServerPort)
		fmt.Fprintf(w, "База данных: %s\n", cfg.DB.DbNAME)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "port": %d}`, cfg.ServerPort)
	})

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	fmt.Printf("Сервер запущен на %s\n", addr)

	log.Fatal(http.ListenAndServe(addr, mux))
}
