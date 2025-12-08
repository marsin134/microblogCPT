package handlers

import (
	"encoding/json"
	"microblogCPT/internal/repository"
	"net/http"

	"microblogCPT/internal/service"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService service.AuthService
	validate    *validator.Validate
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=Author Reader"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Валидация
	if err := h.validate.Struct(req); err != nil {
		writeError(w, "Неверный формат email", http.StatusBadRequest)
		return
	}

	serviceReq := repository.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}

	user, err := h.authService.Register(r.Context(), serviceReq)
	if err != nil {
		if err.Error() == "пользователь с email "+req.Email+" уже существует" {
			writeError(w, "Email уже существует", http.StatusForbidden)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	user, accessToken, refreshToken, err := h.authService.Login(r.Context(), req.Email, req.Password)

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}

	response := map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user": map[string]interface{}{
			"userId": user.UserID,
			"email":  user.Email,
			"role":   user.Role,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, accessToken, refreshToken, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, "Неверный email или пароль", http.StatusForbidden)
		return
	}

	response := map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user": map[string]interface{}{
			"userId": user.UserID,
			"email":  user.Email,
			"role":   user.Role,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, accessToken, refreshToken, err := h.authService.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, "Refresh Token истек или недействителен", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user": map[string]interface{}{
			"userId": user.UserID,
			"email":  user.Email,
			"role":   user.Role,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
