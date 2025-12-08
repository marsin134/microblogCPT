package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse - стандартный ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeError - универсальная функция для отправки ошибок
func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// writeSuccess - функция для успешных ответов
func writeSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
