package handlers

import (
	"encoding/json"
	"microblogCPT/internal/repository"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	userID := pathParts[3]

	currentUserID, ok := r.Context().Value("userID").(string)
	if !ok {
		WriteError(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	currentUserRole, _ := r.Context().Value("role").(string)
	if userID != currentUserID && currentUserRole != "Author" {
		WriteError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	user, err := h.UserRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		WriteError(w, err.Error(), http.StatusUnauthorized)
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

func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// extracting the user id from the url
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	userID := pathParts[3]

	currentUserID, ok := r.Context().Value("userID").(string)
	if !ok {
		WriteError(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	if userID != currentUserID {
		WriteError(w, "Нет прав для обновления этого пользователя", http.StatusForbidden)
		return
	}

	var req struct {
		Email string `json:"email" Validate:"required,email"`
		Role  string `json:"role" Validate:"required,oneof=Author Reader"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// email verification
	patternEmail := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, errorEmail := regexp.MatchString(patternEmail, req.Email)
	if req.Email == "" || errorEmail != nil || !matched {
		WriteError(w, "Неверный формат email", http.StatusBadRequest)
		return
	}

	// role verification
	roleSlice := []string{"Author", "Reader"}
	if req.Role == "" || !slices.Contains(roleSlice, req.Role) {
		WriteError(w, "Роль должна быть Author или Reader", http.StatusBadRequest)
		return
	}

	serviceReq := repository.UpdateUserRequest{
		UserID: userID,
		Email:  req.Email,
		Role:   req.Role,
	}

	if err := h.UserService.UpdateUser(r.Context(), serviceReq); err != nil {
		WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь обновлен"})
}

func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	userID := pathParts[3]

	currentUserID, ok := r.Context().Value("userID").(string)
	if !ok {
		WriteError(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	if userID != currentUserID {
		WriteError(w, "Нет прав для удаления этого пользователя", http.StatusForbidden)
		return
	}

	if err := h.UserService.DeleteUser(r.Context(), userID); err != nil {
		WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь удален"})
}

func (h *Handlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		WriteError(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	// get user by id
	user, err := h.UserRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		WriteError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// forming the response
	response := map[string]interface{}{
		"userId": user.UserID,
		"email":  user.Email,
		"role":   user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
