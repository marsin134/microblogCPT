package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse - standard response with an error
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteError is a universal function for sending errors
func WriteError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// WriteSuccess - function for successful responses
func WriteSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
