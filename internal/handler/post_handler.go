package handlers

import (
	"encoding/json"
	"microblogCPT/internal/models"
	"microblogCPT/internal/repository"
	"net/http"
	"strconv"
	"strings"
	"time"

	"microblogCPT/internal/service"

	"github.com/go-playground/validator/v10"
)

type PostHandler struct {
	postService service.PostService
	postRepo    repository.PostRepository
	validate    *validator.Validate
}

func NewPostHandler(postService service.PostService, postRepo repository.PostRepository) *PostHandler {
	return &PostHandler{
		postService: postService,
		postRepo:    postRepo,
		validate:    validator.New(),
	}
}

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем информацию о пользователе из контекста
	userRole, _ := r.Context().Value("role").(string)
	userID, _ := r.Context().Value("userID").(string)

	// Параметры пагинации
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var posts []models.Post
	var total int
	var err error

	if userRole == "Author" {
		posts, err = h.postRepo.GetByUserID(r.Context(), userID)
	} else {
		posts, err = h.postRepo.GetPublishPosts(r.Context())
	}

	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"posts": posts,
		"pagination": map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	// Извлекаем ID поста из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		writeError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[3]

	userID, _ := r.Context().Value("userID").(string)

	post, err := h.postRepo.GetByID(r.Context(), postID)
	if post.Content != "Published" && post.AuthorID != userID {
		writeError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			writeError(w, "Пост не найден", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "доступ запрещен") {
			writeError(w, "Доступ запрещен", http.StatusForbidden)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userRole, ok := r.Context().Value("role").(string)
	if !ok || userRole != "Author" {
		writeError(w, "Пользователь не является Автором", http.StatusForbidden)
		return
	}

	var req struct {
		IdempotencyKey *string `json:"idempotencyKey"`
		Title          string  `json:"title" validate:"required"`
		Content        string  `json:"content" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	authorID := r.Context().Value("userID").(string)

	serviceReq := repository.CreatePostRequest{
		AuthorID:       authorID,
		IdempotencyKey: req.IdempotencyKey,
		Title:          req.Title,
		Content:        req.Content,
	}

	post, err := h.postService.CreatePost(r.Context(), serviceReq)
	if err != nil {
		if strings.Contains(err.Error(), "ключ идемпотентности уже использован") {
			writeError(w, "Ключ идемпотентности уже использован", http.StatusConflict)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"postId":         post.PostID,
		"idempotencyKey": post.IdempotencyKey,
		"title":          post.Title,
		"content":        post.Content,
		"status":         post.Status,
		"createdAt":      post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updatedAt":      post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userRole, ok := r.Context().Value("role").(string)
	if !ok || userRole != "Author" {
		writeError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		writeError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[3]

	var req struct {
		Title   string `json:"title" validate:"required"`
		Content string `json:"content" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	serviceReq := repository.UpdatePostRequest{
		PostID:  postID,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := h.postService.UpdatePost(r.Context(), serviceReq); err != nil {
		if strings.Contains(err.Error(), "пост не найден") {
			writeError(w, "Пост не найден", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "доступ запрещен") {
			writeError(w, "Доступ запрещен", http.StatusForbidden)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пост успешно обновлен"})
}

// AddedImage - POST /api/posts/{postId}/images
func (h *PostHandler) AddedImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userRole, ok := r.Context().Value("role").(string)
	if !ok || userRole != "Author" {
		writeError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 || pathParts[4] != "images" {
		writeError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[3]
	authorID := r.Context().Value("userID").(string)

	post, err := h.postRepo.GetByID(r.Context(), postID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if post.AuthorID != authorID {
		writeError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		writeError(w, "Не удалось получить файл", http.StatusBadRequest)
		return
	}
	defer file.Close()

	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	contentType := handler.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		writeError(w, "Неподдерживаемый тип файла. Разрешены: JPEG, PNG, GIF, WebP", http.StatusBadRequest)
		return
	}

	image, err := h.postService.AddedImage(r.Context(), postID, handler.Filename, file, handler.Size)
	if err != nil {
		if strings.Contains(err.Error(), "доступ запрещен") {
			writeError(w, "Доступ запрещен", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "пост не найден") {
			writeError(w, "Пост не найден", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "размер файла превышает") {
			writeError(w, err.Error(), http.StatusBadRequest)
		} else {
			writeError(w, "Ошибка загрузки изображения", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"imageId":   image.ImageID,
		"postId":    image.PostID,
		"imageUrl":  image.ImageURL,
		"createdAt": image.CreatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// DeleteImage - DELETE /api/posts/{postId}/images/{imageId}
func (h *PostHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userRole, ok := r.Context().Value("role").(string)
	if !ok || userRole != "Author" {
		writeError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 || pathParts[4] != "images" {
		writeError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[3]
	imageID := pathParts[5]
	authorID := r.Context().Value("userID").(string)
	post, err := h.postRepo.GetByID(r.Context(), postID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if post.AuthorID != authorID {
		writeError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	err = h.postService.DeleteImage(r.Context(), imageID)
	if err != nil {
		if strings.Contains(err.Error(), "доступ запрещен") {
			writeError(w, "Доступ запрещен", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "не найдено") {
			writeError(w, "Пост или картинка не найдены", http.StatusNotFound)
		} else {
			writeError(w, "Ошибка удаления изображения", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Картинка успешно удалена",
		"postId":  postID,
		"imageId": imageID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *PostHandler) PublishPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем роль пользователя
	userRole, ok := r.Context().Value("role").(string)
	if !ok || userRole != "Author" {
		writeError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	// Извлекаем ID поста из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 || pathParts[4] != "status" {
		writeError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[3]

	var req struct {
		Status string `json:"status" validate:"required,oneof=Published"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, "Неверное значение статуса", http.StatusBadRequest)
		return
	}

	if err := h.postService.PublishPost(r.Context(), postID); err != nil {
		if strings.Contains(err.Error(), "пост не найден") {
			writeError(w, "Пост не найден", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "доступ запрещен") {
			writeError(w, "Доступ запрещен", http.StatusForbidden)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пост успешно опубликован"})
}
