package test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"microblogCPT/internal/config"
	handlers "microblogCPT/internal/handler"
	"microblogCPT/internal/models"
	"microblogCPT/internal/repository"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"
)

func TestGetPostsHandler(t *testing.T) {
	tests := []struct {
		name           string
		contextValues  map[string]interface{}
		mockSetup      func(*MockPostRepository)
		expectedStatus int
	}{
		{
			name: "Author получает свои посты",
			contextValues: map[string]interface{}{
				"userID": "123",
				"role":   "Author",
			},
			mockSetup: func(repo *MockPostRepository) {
				repo.On("GetByUserID", mock.Anything, "123").
					Return([]models.Post{
						{
							PostID:    "post1",
							Title:     "Test Post",
							Content:   "Test Content",
							AuthorID:  "123",
							Status:    "Draft",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Reader получает опубликованные посты",
			contextValues: map[string]interface{}{
				"userID": "456",
				"role":   "Reader",
			},
			mockSetup: func(repo *MockPostRepository) {
				repo.On("GetPublishPosts", mock.Anything).
					Return([]models.Post{
						{
							PostID:    "post2",
							Title:     "Published Post",
							Content:   "Published Content",
							AuthorID:  "123",
							Status:    "Published",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthService)
			mockUserRepo := new(MockUserRepository)
			mockUserService := new(MockUserService)
			mockPostService := new(MockPostService)
			mockPostRepo := new(MockPostRepository)

			tt.mockSetup(mockPostRepo)

			cfg := &config.Config{}
			handler := &handlers.Handlers{
				UserService: mockUserService,
				UserRepo:    mockUserRepo,
				AuthService: mockAuthService,
				PostService: mockPostService,
				PostRepo:    mockPostRepo,
				Cfg:         cfg,
				Validate:    validator.New(),
			}

			req := httptest.NewRequest(http.MethodGet, "/api/posts?page=1&limit=20", nil)

			ctx := req.Context()
			for key, value := range tt.contextValues {
				ctx = context.WithValue(ctx, key, value)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetPosts(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(rr.Body.Bytes(), &response)
				assert.Contains(t, response, "posts")
				assert.Contains(t, response, "pagination")
			}

			mockPostRepo.AssertExpectations(t)
		})
	}
}

func TestCreatePostHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		contextValues  map[string]interface{}
		mockSetup      func(*MockPostService)
		expectedStatus int
		shouldCallMock bool
	}{
		{
			name: "Успешное создание поста",
			requestBody: map[string]interface{}{
				"title":          "Test Post",
				"content":        "Test Content",
				"idempotencyKey": "key123",
			},
			contextValues: map[string]interface{}{
				"userID": "123",
				"role":   "Author",
			},
			mockSetup: func(service *MockPostService) {
				key := "key123"
				service.On("CreatePost", mock.Anything, repository.CreatePostRequest{
					AuthorID:       "123",
					Title:          "Test Post",
					Content:        "Test Content",
					IdempotencyKey: &key,
				}).Return(&models.Post{
					PostID:         "post123",
					Title:          "Test Post",
					Content:        "Test Content",
					AuthorID:       "123",
					Status:         "Draft",
					IdempotencyKey: &key,
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			shouldCallMock: true,
		},
		{
			name: "Reader пытается создать пост",
			requestBody: map[string]interface{}{
				"title":   "Test Post",
				"content": "Test Content",
			},
			contextValues: map[string]interface{}{
				"userID": "456",
				"role":   "Reader",
			},
			mockSetup:      func(service *MockPostService) {},
			expectedStatus: http.StatusForbidden,
			shouldCallMock: false,
		},
		{
			name: "Отсутствует заголовок",
			requestBody: map[string]interface{}{
				"content": "Test Content",
			},
			contextValues: map[string]interface{}{
				"userID": "123",
				"role":   "Author",
			},
			mockSetup:      func(service *MockPostService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthService)
			mockUserRepo := new(MockUserRepository)
			mockUserService := new(MockUserService)
			mockPostService := new(MockPostService)
			mockPostRepo := new(MockPostRepository)

			tt.mockSetup(mockPostService)

			cfg := &config.Config{}
			handler := &handlers.Handlers{
				UserService: mockUserService,
				UserRepo:    mockUserRepo,
				AuthService: mockAuthService,
				PostService: mockPostService,
				PostRepo:    mockPostRepo,
				Cfg:         cfg,
				Validate:    validator.New(),
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/posts", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			ctx := req.Context()
			for key, value := range tt.contextValues {
				ctx = context.WithValue(ctx, key, value)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.CreatePost(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.shouldCallMock {
				mockPostService.AssertExpectations(t)
			} else {
				mockPostService.AssertNotCalled(t, "CreatePost", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestPublishPostHandler(t *testing.T) {
	tests := []struct {
		name           string
		urlPath        string
		requestBody    map[string]interface{}
		contextValues  map[string]interface{}
		mockSetup      func(*MockPostService)
		expectedStatus int
	}{
		{
			name:    "Успешная публикация поста",
			urlPath: "/api/posts/post123/status",
			requestBody: map[string]interface{}{
				"status": "Published",
			},
			contextValues: map[string]interface{}{
				"userID": "123",
				"role":   "Author",
			},
			mockSetup: func(service *MockPostService) {
				service.On("PublishPost", mock.Anything, "post123").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Reader пытается опубликовать пост",
			urlPath: "/api/posts/post123/status",
			requestBody: map[string]interface{}{
				"status": "Published",
			},
			contextValues: map[string]interface{}{
				"userID": "456",
				"role":   "Reader",
			},
			mockSetup:      func(service *MockPostService) {},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthService)
			mockUserRepo := new(MockUserRepository)
			mockUserService := new(MockUserService)
			mockPostService := new(MockPostService)
			mockPostRepo := new(MockPostRepository)

			tt.mockSetup(mockPostService)

			cfg := &config.Config{}
			handler := &handlers.Handlers{
				UserService: mockUserService,
				UserRepo:    mockUserRepo,
				AuthService: mockAuthService,
				PostService: mockPostService,
				PostRepo:    mockPostRepo,
				Cfg:         cfg,
				Validate:    validator.New(),
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, tt.urlPath, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			ctx := req.Context()
			for key, value := range tt.contextValues {
				ctx = context.WithValue(ctx, key, value)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.PublishPost(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockPostService.AssertExpectations(t)
		})
	}
}

func TestAddedImageHandler(t *testing.T) {
	tests := []struct {
		name           string
		urlPath        string
		contextValues  map[string]interface{}
		mockSetup      func(*MockPostService, *MockPostRepository)
		expectedStatus int
	}{
		{
			name:    "Успешная загрузка изображения",
			urlPath: "/api/posts/post123/images",
			contextValues: map[string]interface{}{
				"userID": "123",
				"role":   "Author",
			},
			mockSetup: func(service *MockPostService, repo *MockPostRepository) {
				repo.On("GetByID", mock.Anything, "post123").
					Return(&models.Post{
						PostID:   "post123",
						AuthorID: "123",
					}, nil)

				service.On("AddedImage",
					mock.Anything,
					"post123",
					"test.jpg",
					mock.Anything,
					mock.AnythingOfType("int64"),
				).
					Return(&models.Image{
						ImageID:   "img123",
						PostID:    "post123",
						ImageURL:  "http://example.com/image.jpg",
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthService)
			mockUserRepo := new(MockUserRepository)
			mockUserService := new(MockUserService)
			mockPostService := new(MockPostService)
			mockPostRepo := new(MockPostRepository)

			tt.mockSetup(mockPostService, mockPostRepo)

			cfg := &config.Config{
				MaxUploadSize: 10 * 1024 * 1024, // 10MB
			}
			handler := &handlers.Handlers{
				UserService: mockUserService,
				UserRepo:    mockUserRepo,
				AuthService: mockAuthService,
				PostService: mockPostService,
				PostRepo:    mockPostRepo,
				Cfg:         cfg,
				Validate:    validator.New(),
			}

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", `form-data; name="image"; filename="test.jpg"`)
			h.Set("Content-Type", "image/jpeg")

			part, err := writer.CreatePart(h)
			assert.NoError(t, err)

			part.Write([]byte("fake image content"))
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, tt.urlPath, body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			ctx := req.Context()
			for key, value := range tt.contextValues {
				ctx = context.WithValue(ctx, key, value)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.AddedImage(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "imageId")
				assert.Contains(t, response, "imageUrl")
				assert.Equal(t, "img123", response["imageId"])
				assert.Equal(t, "post123", response["postId"])
			}

			mockPostRepo.AssertExpectations(t)
			mockPostService.AssertExpectations(t)
		})
	}
}
