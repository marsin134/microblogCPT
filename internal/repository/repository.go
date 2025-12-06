package repository

import (
	"context"
	"microblogCPT/internal/models"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User, password string) error
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, userID string) error
	VerifyPassword(ctx context.Context, email, password string) (*models.User, error)
	UpdateRefreshToken(ctx context.Context, userID, refreshToken string, expiryTime time.Time) error
	GetUserByRefreshToken(ctx context.Context, refreshToken string) (*models.User, error)
}

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, postID string) (*models.Post, error)
	GetByAuthorID(ctx context.Context, authorID string, status string) ([]*models.Post, error)
	GetAllPublished(ctx context.Context, limit, offset int) ([]*models.Post, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, postID string) error
	Publish(ctx context.Context, postID string) error
	CheckIdempotencyKey(ctx context.Context, authorID, idempotencyKey string) (bool, error)
}

type ImageRepository interface {
	Create(ctx context.Context, image *models.Image) error
	GetByPostID(ctx context.Context, postID string) ([]*models.Image, error)
	Delete(ctx context.Context, imageID string) error
	DeleteByPostID(ctx context.Context, postID string) error
}
