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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCurrentUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		contextValues  map[string]interface{}
		mockSetup      func(service *MockUserService)
		expectedStatus int
	}{
		{
			name: "Успешное получение текущего пользователя",
			contextValues: map[string]interface{}{
				"userID": "123",
			},
			mockSetup: func(service *MockUserService) {
				service.On("GetUserByIDService", mock.Anything, "123").
					Return(&models.User{
						UserID: "123",
						Email:  "test@example.com",
						Role:   "Author",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Пользователь не аутентифицирован",
			contextValues:  map[string]interface{}{},
			mockSetup:      func(service *MockUserService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Пользователь не найден",
			contextValues: map[string]interface{}{
				"userID": "999",
			},
			mockSetup: func(service *MockUserService) {
				service.On("GetUserByID", mock.Anything, "999").
					Return((*models.User)(nil), assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthService)
			mockUserService := new(MockUserService)
			mockPostService := new(MockPostService)

			tt.mockSetup(mockUserService)

			cfg := &config.Config{}
			handler := &handlers.Handlers{
				UserService: mockUserService,
				AuthService: mockAuthService,
				PostService: mockPostService,
				Cfg:         cfg,
				Validate:    validator.New(),
			}

			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)

			ctx := req.Context()
			for key, value := range tt.contextValues {
				ctx = context.WithValue(ctx, key, value)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetCurrentUser(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		urlPath        string
		contextValues  map[string]interface{}
		mockSetup      func(service *MockUserService)
		expectedStatus int
	}{
		{
			name:    "Автор получает другого пользователя",
			urlPath: "/api/user/456",
			contextValues: map[string]interface{}{
				"userID": "123",
				"role":   "Author",
			},
			mockSetup: func(service *MockUserService) {
				service.On("GetUserByIDService", mock.Anything, "456").
					Return(&models.User{
						UserID: "456",
						Email:  "other@example.com",
						Role:   "Reader",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Пользователь получает свой профиль",
			urlPath: "/api/user/123",
			contextValues: map[string]interface{}{
				"userID": "123",
				"role":   "Reader",
			},
			mockSetup: func(service *MockUserService) {
				service.On("GetUserByID", mock.Anything, "123").
					Return(&models.User{
						UserID: "123",
						Email:  "test@example.com",
						Role:   "Reader",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Reader пытается получить другого пользователя",
			urlPath: "/api/user/456",
			contextValues: map[string]interface{}{
				"userID": "123",
				"role":   "Reader",
			},
			mockSetup:      func(service *MockUserService) {},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthService)
			mockUserService := new(MockUserService)
			mockPostService := new(MockPostService)

			tt.mockSetup(mockUserService)

			cfg := &config.Config{}
			handler := &handlers.Handlers{
				UserService: mockUserService,
				AuthService: mockAuthService,
				PostService: mockPostService,
				Cfg:         cfg,
				Validate:    validator.New(),
			}

			req := httptest.NewRequest(http.MethodGet, tt.urlPath, nil)

			ctx := req.Context()
			for key, value := range tt.contextValues {
				ctx = context.WithValue(ctx, key, value)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetUser(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestUpdateUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		urlPath        string
		requestBody    map[string]interface{}
		contextValues  map[string]interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:    "Успешное обновление пользователя",
			urlPath: "/api/user/123",
			requestBody: map[string]interface{}{
				"email": "newemail@example.com",
				"role":  "Author",
			},
			contextValues: map[string]interface{}{
				"userID": "123",
			},
			mockSetup: func(service *MockUserService) {
				service.On("UpdateUser", mock.Anything, repository.UpdateUserRequest{
					UserID: "123",
					Email:  "newemail@example.com",
					Role:   "Author",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Обновление чужого профиля",
			urlPath: "/api/user/456",
			requestBody: map[string]interface{}{
				"email": "newemail@example.com",
				"role":  "Author",
			},
			contextValues: map[string]interface{}{
				"userID": "123",
			},
			mockSetup:      func(service *MockUserService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:    "Невалидный email",
			urlPath: "/api/user/123",
			requestBody: map[string]interface{}{
				"email": "invalid-email",
				"role":  "Author",
			},
			contextValues: map[string]interface{}{
				"userID": "123",
			},
			mockSetup:      func(service *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthService)
			mockUserService := new(MockUserService)
			mockPostService := new(MockPostService)

			tt.mockSetup(mockUserService)

			cfg := &config.Config{}
			handler := &handlers.Handlers{
				UserService: mockUserService,
				AuthService: mockAuthService,
				PostService: mockPostService,
				Cfg:         cfg,
				Validate:    validator.New(),
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, tt.urlPath, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			ctx := req.Context()
			for key, value := range tt.contextValues {
				ctx = context.WithValue(ctx, key, value)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.UpdateUser(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockUserService.AssertExpectations(t)
		})
	}
}
