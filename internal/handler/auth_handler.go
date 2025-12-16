package handlers

import (
	"encoding/json"
	"microblogCPT/internal/repository"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type RegisterRequest struct {
	Email    string `json:"email" Validate:"required,email"`
	Password string `json:"password" Validate:"required,min=6"`
	Role     string `json:"role" Validate:"required,oneof=Author Reader"`
}

type AuthResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodPost {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Handler is not nil
	if h == nil {
		WriteError(w, "Handler is nil", http.StatusInternalServerError)
		return
	}

	// present validate
	if h.Validate == nil {
		WriteError(w, "Validator is not initialized", http.StatusInternalServerError)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// email verification
	patternEmail := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(patternEmail, req.Email)
	if err != nil || !matched {
		WriteError(w, "Неверный формат email", http.StatusBadRequest)
		return
	}

	// password verification
	if utf8.RuneCountInString(req.Password) < 6 {
		WriteError(w, "Пароль должен быть не менее 6 символов", http.StatusBadRequest)
		return
	}

	// role verification
	roleSlice := []string{"Author", "Reader"}
	if !slices.Contains(roleSlice, req.Role) {
		WriteError(w, "Роль должна быть Author или Reader", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		WriteError(w, "Неверные данные", http.StatusBadRequest)
		return
	}

	// creating a form to create user
	serviceReq := repository.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}

	// registering a user in the service
	user, err := h.AuthService.Register(r.Context(), serviceReq)
	if err != nil {
		if strings.Contains(err.Error(), "уже существует") {
			WriteError(w, "Email уже существует", http.StatusForbidden)
		} else {
			WriteError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// logging
	user, accessToken, refreshToken, err := h.AuthService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// forming the response
	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			UserId: user.UserID,
			Email:  user.Email,
			Role:   user.Role,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodPost {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		if strings.Contains(err.Error(), "Email") {
			WriteError(w, "Неверный формат email", http.StatusBadRequest)
		} else {
			WriteError(w, "Неверные данные", http.StatusBadRequest)
		}
		return
	}

	// logging
	user, accessToken, refreshToken, err := h.AuthService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		WriteError(w, "Неверный email или пароль", http.StatusForbidden)
		return
	}

	// forming the response
	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			UserId: user.UserID,
			Email:  user.Email,
			Role:   user.Role,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodPost {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refreshToken" Validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// token missing
	if req.RefreshToken == "" {
		WriteError(w, "Отсуствует refreshToken", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// update accessToken and refreshToken
	user, accessToken, refreshToken, err := h.AuthService.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		WriteError(w, "Refresh Token истек или недействителен", http.StatusBadRequest)
		return
	}

	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			UserId: user.UserID,
			Email:  user.Email,
			Role:   user.Role,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
