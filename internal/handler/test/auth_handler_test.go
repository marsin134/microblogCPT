package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"microblogCPT/internal/config"
	handlers "microblogCPT/internal/handler"
	"microblogCPT/internal/models"
	"microblogCPT/internal/repository"
)

func createTestHandler(authService *MockAuthService) *handlers.Handlers {
	cfg := &config.Config{
		JWTSecretKey:  "test-secret-key",
		ServerPort:    8080,
		MaxUploadSize: 10 * 1024 * 1024,
	}

	return &handlers.Handlers{
		UserService: &MockUserService{},
		UserRepo:    &MockUserRepository{},
		AuthService: authService,
		PostService: nil,
		PostRepo:    nil,
		Cfg:         cfg,
		Validate:    validator.New(),
	}
}

// assertJSONError checks the JSON response with an error
func assertJSONError(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
	assert.Equal(t, expectedStatus, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], expectedError)
}

// assertJSONSuccess checks the successful JSON response
func assertJSONSuccess(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(t, expectedStatus, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestRegisterHandler_Success(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "Author",
	}

	// Setting up mock
	mockAuthService.On("Register", mock.Anything, repository.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "Author",
	}).Return(&models.User{
		UserID: "user-123",
		Email:  "test@example.com",
		Role:   "Author",
	}, nil)

	mockAuthService.On("Login", mock.Anything, "test@example.com", "password123").
		Return(&models.User{
			UserID: "user-123",
			Email:  "test@example.com",
			Role:   "Author",
		}, "access-token-123", "refresh-token-123", nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Register(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "access-token-123", response["accessToken"])
	assert.Equal(t, "refresh-token-123", response["refreshToken"])

	userData, ok := response["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "user-123", userData["userId"])
	assert.Equal(t, "test@example.com", userData["email"])
	assert.Equal(t, "Author", userData["role"])

	mockAuthService.AssertExpectations(t)
}

func TestRegisterHandler_InvalidEmail(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"email":    "invalid-email",
		"password": "password123",
		"role":     "Author",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Register(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["error"], "email")

	// Making sure that the service was not called
	mockAuthService.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
}

func TestRegisterHandler_ShortPassword(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "123",
		"role":     "Author",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Register(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusBadRequest, "Пароль должен быть не менее 6 символов")
	mockAuthService.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
}

func TestRegisterHandler_InvalidRole(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "InvalidRole",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Register(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusBadRequest, "Роль должна быть Author или Reader")
	mockAuthService.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
}

func TestRegisterHandler_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"email":    "existing@example.com",
		"password": "password123",
		"role":     "Author",
	}

	// Setting up mock
	mockAuthService.On("Register", mock.Anything, repository.CreateUserRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Role:     "Author",
	}).Return((*models.User)(nil), fmt.Errorf("пользователь с email existing@example.com уже существует"))

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Register(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusForbidden, "Email уже существует")
	mockAuthService.AssertExpectations(t)
}

func TestRegisterHandler_WrongMethod(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/register", nil)
	rr := httptest.NewRecorder()

	// Act
	handler.Register(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusMethodNotAllowed, "Method not allowed")
	mockAuthService.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
}

// Test login

func TestLoginHandler_Success(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"email":    "user@example.com",
		"password": "password123",
	}

	// Setting up mock
	mockAuthService.On("Login", mock.Anything, "user@example.com", "password123").
		Return(&models.User{
			UserID: "user-456",
			Email:  "user@example.com",
			Role:   "Author",
		}, "access-token-456", "refresh-token-456", nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Login(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "access-token-456", response["accessToken"])
	assert.Equal(t, "refresh-token-456", response["refreshToken"])

	userData, ok := response["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "user-456", userData["userId"])
	assert.Equal(t, "user@example.com", userData["email"])
	assert.Equal(t, "Author", userData["role"])

	mockAuthService.AssertExpectations(t)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"email":    "wrong@example.com",
		"password": "wrongpass",
	}

	// Setting up mock
	mockAuthService.On("Login", mock.Anything, "wrong@example.com", "wrongpass").
		Return((*models.User)(nil), "", "", fmt.Errorf("неверные учетные данные"))

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Login(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusForbidden, "Неверный email или пароль")
	mockAuthService.AssertExpectations(t)
}

func TestLoginHandler_InvalidEmail(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"email":    "invalid-email",
		"password": "password123",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Login(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusBadRequest, "Неверный формат email")
	mockAuthService.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything)
}

func TestLoginHandler_MissingPassword(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"email": "test@example.com",
		// password absent
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Login(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusBadRequest, "Неверные данные")
	mockAuthService.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything)
}

func TestLoginHandler_WrongMethod(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	rr := httptest.NewRecorder()

	// Act
	handler.Login(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusMethodNotAllowed, "Method not allowed")
	mockAuthService.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything)
}

func TestRefreshTokenHandler_Success(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"refreshToken": "valid-refresh-token",
	}

	// Setting up mock
	mockAuthService.On("RefreshTokens", mock.Anything, "valid-refresh-token").
		Return(&models.User{
			UserID: "user-789",
			Email:  "user@example.com",
			Role:   "Reader",
		}, "new-access-token-789", "new-refresh-token-789", nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.RefreshToken(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "new-access-token-789", response["accessToken"])
	assert.Equal(t, "new-refresh-token-789", response["refreshToken"])

	userData, ok := response["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "user-789", userData["userId"])
	assert.Equal(t, "user@example.com", userData["email"])
	assert.Equal(t, "Reader", userData["role"])

	mockAuthService.AssertExpectations(t)
}

func TestRefreshTokenHandler_InvalidToken(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"refreshToken": "invalid-token",
	}

	// Настраиваем мок на возврат ошибки
	mockAuthService.On("RefreshTokens", mock.Anything, "invalid-token").
		Return((*models.User)(nil), "", "", fmt.Errorf("refresh token истек или недействителен"))

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.RefreshToken(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusBadRequest, "Refresh Token истек или недействителен")
	mockAuthService.AssertExpectations(t)
}

func TestRefreshTokenHandler_MissingToken(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"otherField": "value",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.RefreshToken(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusBadRequest, "")
	mockAuthService.AssertNotCalled(t, "RefreshTokens", mock.Anything, mock.Anything)
}

func TestRefreshTokenHandler_WrongMethod(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/refresh-token", nil)
	rr := httptest.NewRecorder()

	// Act
	handler.RefreshToken(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusMethodNotAllowed, "Method not allowed")
	mockAuthService.AssertNotCalled(t, "RefreshTokens", mock.Anything, mock.Anything)
}

// integration tests

func TestAuthFlow_Integration(t *testing.T) {
	// Full Cycle authentication test
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	// 1. Registration
	registerBody := map[string]interface{}{
		"email":    "newuser@example.com",
		"password": "securepass123",
		"role":     "Author",
	}

	mockAuthService.On("Register", mock.Anything, repository.CreateUserRequest{
		Email:    "newuser@example.com",
		Password: "securepass123",
		Role:     "Author",
	}).Return(&models.User{
		UserID: "new-user-123",
		Email:  "newuser@example.com",
		Role:   "Author",
	}, nil)

	mockAuthService.On("Login", mock.Anything, "newuser@example.com", "securepass123").
		Return(&models.User{
			UserID: "new-user-123",
			Email:  "newuser@example.com",
			Role:   "Author",
		}, "initial-access-token", "initial-refresh-token", nil)

	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// 3. update token
	refreshBody := map[string]interface{}{
		"refreshToken": "logged-in-refresh-token",
	}

	mockAuthService.On("RefreshTokens", mock.Anything, "logged-in-refresh-token").
		Return(&models.User{
			UserID: "new-user-123",
			Email:  "newuser@example.com",
			Role:   "Author",
		}, "refreshed-access-token", "refreshed-refresh-token", nil)

	body, _ = json.Marshal(refreshBody)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	handler.RefreshToken(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	mockAuthService.AssertExpectations(t)
}

func TestRegisterHandler_EmptyRequestBody(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Register(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusBadRequest, "Неверный формат запроса")
}

func TestLoginHandler_MalformedJSON(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.Login(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusBadRequest, "Неверный формат запроса")
}

func TestRefreshTokenHandler_EmptyBody(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Act
	handler.RefreshToken(rr, req)

	// Assert
	assertJSONError(t, rr, http.StatusBadRequest, "Неверный формат запроса")
}

func BenchmarkRegisterHandler(b *testing.B) {
	mockAuthService := new(MockAuthService)
	handler := createTestHandler(mockAuthService)

	requestBody := map[string]interface{}{
		"email":    "benchmark@example.com",
		"password": "password123",
		"role":     "Author",
	}

	body, _ := json.Marshal(requestBody)

	mockAuthService.On("Register", mock.Anything, mock.Anything).
		Return(&models.User{
			UserID: "bench-user",
			Email:  "benchmark@example.com",
			Role:   "Author",
		}, nil)

	mockAuthService.On("Login", mock.Anything, "benchmark@example.com", "password123").
		Return(&models.User{
			UserID: "bench-user",
			Email:  "benchmark@example.com",
			Role:   "Author",
		}, "bench-token", "bench-refresh", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.Register(rr, req)
	}
}
