package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"microblogCPT/internal/models"
	"time"
)

type userRepository struct {
	db *sqlx.DB
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User, password string) error {
	// create password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("ошибка при хешировании пароля: %w", err)
	}

	// create user id
	user.UserID = uuid.New().String()
	user.PasswordHash = string(hashedPassword)

	query := `
		INSERT INTO users (user_id, email, password_hash, role, refresh_token, refresh_token_expiry_time)
		VALUES (:user_id, :email, :password_hash, :role, :refresh_token, :refresh_token_expiry_time)
	`

	_, err = r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return nil
}

func (r *userRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User

	query := `SELECT * FROM users WHERE user_id = $1`

	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пользователь с ID %s не найден", userID)
		}
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	query := `SELECT * FROM users WHERE email = $1`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пользователь с email %s не найден", email)
		}
		return nil, fmt.Errorf("ошибка при получении пользователя по email: %w", err)
	}

	return &user, nil
}

func (r *userRepository) VerifyPassword(ctx context.Context, email, password string) (*models.User, error) {
	user, err := r.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// checking that the password hash is the same
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("неверный пароль")
	}

	return user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = :email, role = :role
		WHERE user_id = :user_id
	`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении пользователя: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("пользователь с ID %s не найден", user.UserID)
	}

	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении пользователя: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке удаленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("пользователь с ID %s не найден", userID)
	}

	return nil
}

func (r *userRepository) UpdateRefreshToken(ctx context.Context, userID, refreshToken string, expiryTime time.Time) error {
	query := `
		UPDATE users 
		SET refresh_token = $1, refresh_token_expiry_time = $2
		WHERE user_id = $3
	`

	_, err := r.db.ExecContext(ctx, query, refreshToken, expiryTime, userID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении refresh token: %w", err)
	}

	return nil
}

func (r *userRepository) GetUserByRefreshToken(ctx context.Context, refreshToken string) (*models.User, error) {
	var user models.User

	query := `
		SELECT * FROM users 
		WHERE refresh_token = $1 
		AND refresh_token_expiry_time > CURRENT_TIMESTAMP
	`

	err := r.db.GetContext(ctx, &user, query, refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("недействительный или просроченный refresh token")
		}
		return nil, fmt.Errorf("ошибка при получении пользователя по refresh token: %w", err)
	}

	return &user, nil
}
