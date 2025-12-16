package handlers

import (
	"encoding/json"
	"fmt"
	"microblogCPT/internal/models"
	"microblogCPT/internal/repository"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type PostsGetResponse struct {
	Posts      []models.Post
	Pagination PaginationResponse
}

type PostResponse struct {
	PostId         string    `json:"postId"`
	IdempotencyKey *string   `json:"idempotencyKey"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ImageResponse struct {
	ImageID   string `json:"image_id"`
	PostID    string `json:"post_id"`
	ImageUrl  string `json:"imageUrl"`
	FileName  string `json:"fileName"`
	FileSize  int64  `json:"fileSize"`
	MimeType  string `json:"mimeType"`
	CreatedAt string `json:"createdAt"`
}

func (h *Handlers) GetPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Getting information about the user from the context
	userRole, _ := r.Context().Value("role").(string)
	userID, _ := r.Context().Value("userID").(string)

	// Pagination parameters
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

	if userRole == "Author" { // Returning the user posts
		posts, err = h.PostRepo.GetByUserID(r.Context(), userID)
	} else { // Returning the all posts
		posts, err = h.PostRepo.GetPublishPosts(r.Context())
	}

	if err != nil {
		WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// forming the response
	response := PostsGetResponse{
		Posts: posts,
		Pagination: PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: (total + limit - 1) / limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) GetPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extracting the post ID from the URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[3]

	userID, _ := r.Context().Value("userID").(string)

	// we receive a post on id
	post, err := h.PostRepo.GetByID(r.Context(), postID)
	if post.Content != "Published" && post.AuthorID != userID {
		WriteError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			WriteError(w, "Пост не найден", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "доступ запрещен") {
			WriteError(w, "Доступ запрещен", http.StatusForbidden)
		} else {
			WriteError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func (h *Handlers) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut { // if put, then we update the user
		h.UpdatePost(w, r)
		return
	}

	if r.Method != http.MethodPost {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// checking the user's role
	userRole, ok := r.Context().Value("role").(string)
	if !ok || userRole != "Author" {
		WriteError(w, "Пользователь не является Автором", http.StatusForbidden)
		return
	}

	var req struct {
		IdempotencyKey *string `json:"idempotencyKey"`
		Title          string  `json:"title" Validate:"required"`
		Content        string  `json:"content" Validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// checking the title of the post
	if req.Title == "" {
		WriteError(w, "Отсутствует заголовок", http.StatusBadRequest)
		return
	}

	authorID := r.Context().Value("userID").(string)

	serviceReq := repository.CreatePostRequest{
		AuthorID:       authorID,
		IdempotencyKey: req.IdempotencyKey,
		Title:          req.Title,
		Content:        req.Content,
	}

	// creating a post
	post, err := h.PostService.CreatePost(r.Context(), serviceReq)
	if err != nil {
		if strings.Contains(err.Error(), "ключ идемпотентности уже использован") {
			WriteError(w, "Ключ идемпотентности уже использован", http.StatusConflict)
		} else {
			WriteError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// forming the response
	response := PostResponse{
		PostId:         post.PostID,
		IdempotencyKey: post.IdempotencyKey,
		Title:          post.Title,
		Content:        post.Content,
		Status:         post.Status,
		CreatedAt:      post.CreatedAt,
		UpdatedAt:      post.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// checking the user's role
	userRole, ok := r.Context().Value("role").(string)
	if !ok || userRole != "Author" {
		WriteError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	// check url
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[3]

	var req struct {
		Title   string `json:"title" Validate:"required"`
		Content string `json:"content" Validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		WriteError(w, "Отсутствует заголовок", http.StatusBadRequest)
		return
	}

	serviceReq := repository.UpdatePostRequest{
		PostID:  postID,
		Title:   req.Title,
		Content: req.Content,
	}

	// updating the post
	if err := h.PostService.UpdatePost(r.Context(), serviceReq); err != nil {
		if strings.Contains(err.Error(), "пост не найден") {
			WriteError(w, "Пост не найден", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "доступ запрещен") {
			WriteError(w, "Доступ запрещен", http.StatusForbidden)
		} else {
			WriteError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{Message: "Пост успешно обновлен"})
}

func (h *Handlers) AddedImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// checking the user's role
	userRole, ok := r.Context().Value("role").(string)
	if !ok || userRole != "Author" {
		WriteError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	// check url
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 || pathParts[4] != "images" {
		WriteError(w, "Неверный URL", http.StatusBadRequest)
		return
	}

	// we receive a post on id
	postID := pathParts[3]
	post, err := h.PostRepo.GetByID(r.Context(), postID)
	if err != nil {
		WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// we check that only the author can add
	authorID := r.Context().Value("userID").(string)
	if authorID != post.AuthorID {
		WriteError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	// setting the size limit from the config
	if err := r.ParseMultipartForm(h.Cfg.MaxUploadSize); err != nil {
		if err.Error() == "http: request body too large" {
			WriteError(w, fmt.Sprintf("Файл слишком большой (макс. %d MB)",
				h.Cfg.MaxUploadSize/(1024*1024)), http.StatusBadRequest)
		} else {
			WriteError(w, "Ошибка при обработке файла", http.StatusBadRequest)
		}
		return
	}

	// getting the file
	file, handler, err := r.FormFile("image")
	if err != nil {
		WriteError(w, "Не удалось получить файл", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// formats image
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	// check formats
	contentType := handler.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		WriteError(w, "Неподдерживаемый тип файла. Разрешены: JPEG, PNG, GIF, WebP", http.StatusBadRequest)
		return
	}

	// added image
	image, err := h.PostService.AddedImage(r.Context(), postID, handler.Filename, file, handler.Size)
	if err != nil {
		if strings.Contains(err.Error(), "доступ запрещен") {
			WriteError(w, "Доступ запрещен", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "пост не найден") {
			WriteError(w, "Пост не найден", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "размер файла превышает") {
			WriteError(w, err.Error(), http.StatusBadRequest)
		} else {
			WriteError(w, "Ошибка загрузки изображения", http.StatusInternalServerError)
		}
		return
	}

	// forming the response
	response := ImageResponse{
		ImageID:   image.ImageID,
		PostID:    image.PostID,
		ImageUrl:  image.ImageURL,
		FileName:  handler.Filename,
		FileSize:  handler.Size,
		MimeType:  contentType,
		CreatedAt: image.CreatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) DeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// checking the user's role
	userRole, ok := r.Context().Value("userRole").(string)
	if !ok || userRole != "Author" {
		WriteError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	// check url
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 || pathParts[4] != "images" {
		WriteError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[3]
	imageID := pathParts[5]

	// get post by id
	post, err := h.PostRepo.GetByID(r.Context(), postID)
	if err != nil {
		WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// we check that only the author can delete
	authorID := r.Context().Value("userID").(string)
	if authorID != post.AuthorID {
		WriteError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	// delete image
	err = h.PostService.DeleteImage(r.Context(), imageID)
	if err != nil {
		if strings.Contains(err.Error(), "доступ запрещен") {
			WriteError(w, "Доступ запрещен", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "не найдено") {
			WriteError(w, "Пост или картинка не найдены", http.StatusNotFound)
		} else {
			WriteError(w, "Ошибка удаления изображения", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{Message: "Картинка успешно удалена"})
}

func (h *Handlers) PublishPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// checking the user's role
	userRole, ok := r.Context().Value("role").(string)
	if !ok || userRole != "Author" {
		WriteError(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	// extracting the post id from the url
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 || pathParts[4] != "status" {
		WriteError(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	postID := pathParts[3]

	var req struct {
		Status string `json:"status" Validate:"required,oneof=Published"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.Validate.Struct(req); err != nil {
		WriteError(w, "Неверное значение статуса", http.StatusBadRequest)
		return
	}

	if err := h.PostService.PublishPost(r.Context(), postID); err != nil {
		if strings.Contains(err.Error(), "пост не найден") {
			WriteError(w, "Пост не найден", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "доступ запрещен") {
			WriteError(w, "Доступ запрещен", http.StatusForbidden)
		} else {
			WriteError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MessageResponse{Message: "Пост успешно опубликован"})
}
