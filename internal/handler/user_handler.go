package handlers

import (
	"encoding/json"
	"microblogCPT/internal/repository"
	"net/http"
	"strings"

	"microblogCPT/internal/service"

	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService service.UserService
	userRepo    repository.UserRepository
	validate    *validator.Validate
}

func NewUserHandler(userService service.UserService, userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		userService: userService,
		userRepo:    userRepo,
		validate:    validator.New(),
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		writeError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	userID := pathParts[3]

	currentUserID, ok := r.Context().Value("userID").(string)
	if !ok {
		writeError(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	currentUserRole, _ := r.Context().Value("role").(string)
	if userID != currentUserID && currentUserRole != "Author" {
		writeError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"userId": user.UserID,
		"email":  user.Email,
		"role":   user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID пользователя из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		writeError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	userID := pathParts[3]

	currentUserID, ok := r.Context().Value("userID").(string)
	if !ok {
		writeError(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	if userID != currentUserID {
		writeError(w, "Нет прав для обновления этого пользователя", http.StatusForbidden)
		return
	}

	var req struct {
		Email string `json:"email" validate:"required,email"`
		Role  string `json:"role" validate:"required,oneof=Author Reader"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	serviceReq := repository.UpdateUserRequest{
		UserID: userID,
		Email:  req.Email,
		Role:   req.Role,
	}

	if err := h.userService.UpdateUser(r.Context(), serviceReq); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь обновлен"})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		writeError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	userID := pathParts[3]

	currentUserID, ok := r.Context().Value("userID").(string)
	if !ok {
		writeError(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	if userID != currentUserID {
		writeError(w, "Нет прав для удаления этого пользователя", http.StatusForbidden)
		return
	}

	if err := h.userService.DeleteUser(r.Context(), userID); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь удален"})
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		writeError(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"userId": user.UserID,
		"email":  user.Email,
		"role":   user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
