package test

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"io"
	"microblogCPT/internal/models"
	"microblogCPT/internal/repository"
	"time"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req repository.CreateUserRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*models.User, string, string, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(*models.User), args.String(1), args.String(2), args.Error(3)
}

func (m *MockAuthService) RefreshTokens(ctx context.Context, refreshToken string) (*models.User, string, string, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(*models.User), args.String(1), args.String(2), args.Error(3)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockAuthService) GetUserFromToken(token string) (*models.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User, password string) error {
	args := m.Called(ctx, user, password)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) VerifyPassword(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateRefreshToken(ctx context.Context, userID, refreshToken string, expiryTime time.Time) error {
	args := m.Called(ctx, userID, refreshToken, expiryTime)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByRefreshToken(ctx context.Context, refreshToken string) (*models.User, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) UpdateUser(ctx context.Context, req repository.UpdateUserRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockPostService struct {
	mock.Mock
}

func (m *MockPostService) CreatePost(ctx context.Context, req repository.CreatePostRequest) (*models.Post, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostService) UpdatePost(ctx context.Context, req repository.UpdatePostRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockPostService) DeletePost(ctx context.Context, postID string) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostService) PublishPost(ctx context.Context, postID string) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostService) AddedImage(ctx context.Context, postID, fileName string, file io.Reader, size int64) (*models.Image, error) {
	args := m.Called(ctx, postID, fileName, file, size)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Image), args.Error(1)
}

func (m *MockPostService) DeleteImage(ctx context.Context, imageID string) error {
	args := m.Called(ctx, imageID)
	return args.Error(0)
}

type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(ctx context.Context, post *models.Post, imagesURL []string) error {
	args := m.Called(ctx, post, imagesURL)
	return args.Error(0)
}

func (m *MockPostRepository) GetByID(ctx context.Context, postID string) (*models.Post, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) GetByUserID(ctx context.Context, userID string) ([]models.Post, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Post), args.Error(1)
}

func (m *MockPostRepository) GetPublishPosts(ctx context.Context) ([]models.Post, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Post), args.Error(1)
}

func (m *MockPostRepository) Update(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) Delete(ctx context.Context, postID string) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostRepository) CheckIdempotencyKey(ctx context.Context, authorID, idempotencyKey string) (bool, error) {
	args := m.Called(ctx, authorID, idempotencyKey)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostRepository) Publish(ctx context.Context, postID string) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}
