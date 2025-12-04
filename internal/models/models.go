package models

import (
	"time"
)

type User struct {
	UserID                 string    `json:"userId" db:"user_id"`
	Email                  string    `json:"email" db:"email"`
	PasswordHash           string    `json:"passwordHash" db:"password_hash"`
	Role                   string    `json:"role" db:"role"`
	RefreshToken           string    `json:"refreshToken" db:"refresh_token"`
	RefreshTokenExpiryTime time.Time `json:"refreshTokenExpiryTime" db:"refresh_token_expiry_time"`
}

type CreateUserRequest struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
	Role     string `json:"role" db:"role"`
}

type Post struct {
	PostID         string    `json:"postId" db:"post_id"`
	AuthorID       string    `json:"authorId" db:"author_id"`
	IdempotencyKey string    `json:"idempotencyKey" db:"idempotency_key"`
	Title          string    `json:"title" db:"title"`
	Content        string    `json:"content" db:"content"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
	Status         string    `json:"status" db:"status"`
	Images         []Image   `json:"images" db:"-"`
}

type Image struct {
	ImageID   string    `json:"imageId" db:"image_id"`
	PostID    string    `json:"postId" db:"post_id"`
	ImageURL  string    `json:"imageUrl" db:"image_url"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
