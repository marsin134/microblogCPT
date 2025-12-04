package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"microblogCPT/internal/models"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	ctx := context.Background()

	email := "test@example.com"
	password := "password123"
	role := "Author"

	// Создаем пользователя БЕЗ предустановленного ID
	user := &models.User{
		Email:                  email,
		Role:                   role,
		RefreshToken:           "refresh_token",
		RefreshTokenExpiryTime: time.Time{},
	}

	t.Run("Успешное создание пользователя", func(t *testing.T) {
		mock.ExpectExec(`
			INSERT INTO users (user_id, email, password_hash, role, refresh_token, refresh_token_expiry_time)
			VALUES (?, ?, ?, ?, ?, ?)
		`).
			WithArgs(
				sqlmock.AnyArg(), // user_id будет сгенерирован в репозитории
				email,
				sqlmock.AnyArg(), // password_hash
				role,
				"refresh_token",
				time.Time{},
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateUser(ctx, user, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, user.UserID) // Проверяем что ID был сгенерирован
		assert.NotEqual(t, password, user.PasswordHash)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Ошибка при дублировании email", func(t *testing.T) {
		// Для второго теста нужно создать нового пользователя
		user2 := &models.User{
			Email:                  email,
			Role:                   role,
			RefreshToken:           "refresh_token",
			RefreshTokenExpiryTime: time.Time{},
		}

		mock.ExpectExec(`
			INSERT INTO users (user_id, email, password_hash, role, refresh_token, refresh_token_expiry_time)
			VALUES (?, ?, ?, ?, ?, ?)
		`).
			WithArgs(
				sqlmock.AnyArg(),
				email,
				sqlmock.AnyArg(),
				role,
				"refresh_token",
				time.Time{},
			).
			WillReturnError(errors.New("duplicate key value violates unique constraint"))

		err := repo.CreateUser(ctx, user2, password)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при создании пользователя")
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	ctx := context.Background()
	userID := uuid.New().String()
	expectedUser := &models.User{
		UserID:                 userID,
		Email:                  "test@example.com",
		PasswordHash:           "hashed_password",
		Role:                   "Author",
		RefreshToken:           "refresh_token",
		RefreshTokenExpiryTime: time.Now().Add(24 * time.Hour),
	}

	t.Run("Успешное получение пользователя по ID", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"user_id", "email", "password_hash", "role",
			"refresh_token", "refresh_token_expiry_time",
		}).
			AddRow(
				expectedUser.UserID,
				expectedUser.Email,
				expectedUser.PasswordHash,
				expectedUser.Role,
				expectedUser.RefreshToken,
				expectedUser.RefreshTokenExpiryTime,
			)

		mock.ExpectQuery(`SELECT * FROM users WHERE user_id = $1`).
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetUserByID(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, expectedUser.UserID, user.UserID)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Role, user.Role)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Пользователь не найден", func(t *testing.T) {
		mock.ExpectQuery(`SELECT * FROM users WHERE user_id = $1`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetUserByID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "не найден")
	})

	t.Run("Ошибка базы данных", func(t *testing.T) {
		mock.ExpectQuery(`SELECT * FROM users WHERE user_id = $1`).
			WithArgs(userID).
			WillReturnError(errors.New("connection failed"))

		user, err := repo.GetUserByID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "ошибка при получении пользователя")
	})
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	ctx := context.Background()
	email := "test@example.com"

	t.Run("Успешное получение по email", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"user_id", "email", "password_hash", "role",
			"refresh_token", "refresh_token_expiry_time",
		}).
			AddRow(uuid.New().String(), email, "hashed_password", "Reader", "", time.Time{})

		mock.ExpectQuery(`SELECT * FROM users WHERE email = $1`).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.GetUserByEmail(ctx, email)

		require.NoError(t, err)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, "Reader", user.Role)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestUserRepository_VerifyPassword(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	ctx := context.Background()
	email := "test@example.com"
	password := "correct_password"
	wrongPassword := "wrong_password"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	t.Run("Успешная проверка пароля", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"user_id", "email", "password_hash", "role",
			"refresh_token", "refresh_token_expiry_time",
		}).
			AddRow(uuid.New().String(), email, string(hashedPassword), "Author", "", time.Time{})

		mock.ExpectQuery(`SELECT * FROM users WHERE email = $1`).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.VerifyPassword(ctx, email, password)

		require.NoError(t, err)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, "Author", user.Role)
	})

	t.Run("Неверный пароль", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"user_id", "email", "password_hash", "role",
			"refresh_token", "refresh_token_expiry_time",
		}).
			AddRow(uuid.New().String(), email, string(hashedPassword), "Author", "", time.Time{})

		mock.ExpectQuery(`SELECT * FROM users WHERE email = $1`).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.VerifyPassword(ctx, email, wrongPassword)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "неверный пароль")
	})

	t.Run("Пользователь не найден", func(t *testing.T) {
		mock.ExpectQuery(`SELECT * FROM users WHERE email = $1`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.VerifyPassword(ctx, email, password)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "не найден")
	})
}

func TestUserRepository_UpdateRefreshToken(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	ctx := context.Background()
	userID := uuid.New().String()
	refreshToken := "new_refresh_token"
	expiryTime := time.Now().Add(168 * time.Hour)

	t.Run("Успешное обновление refresh token", func(t *testing.T) {
		mock.ExpectExec(`UPDATE users SET refresh_token = $1, refresh_token_expiry_time = $2 WHERE user_id = $3`).
			WithArgs(refreshToken, expiryTime, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateRefreshToken(ctx, userID, refreshToken, expiryTime)

		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Ошибка при обновлении", func(t *testing.T) {
		mock.ExpectExec(`UPDATE users SET refresh_token = $1, refresh_token_expiry_time = $2 WHERE user_id = $3`).
			WithArgs(refreshToken, expiryTime, userID).
			WillReturnError(errors.New("update failed"))

		err := repo.UpdateRefreshToken(ctx, userID, refreshToken, expiryTime)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при обновлении refresh token")
	})
}

func TestUserRepository_GetUserByRefreshToken(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	ctx := context.Background()
	refreshToken := "valid_refresh_token"
	expiredToken := "expired_refresh_token"
	validExpiry := time.Now().Add(1 * time.Hour)

	t.Run("Успешное получение по валидному refresh token", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"user_id", "email", "password_hash", "role",
			"refresh_token", "refresh_token_expiry_time",
		}).
			AddRow(
				uuid.New().String(),
				"test@example.com",
				"hashed_password",
				"Author",
				refreshToken,
				validExpiry,
			)

		mock.ExpectQuery(`SELECT * FROM users WHERE refresh_token = $1 AND refresh_token_expiry_time > CURRENT_TIMESTAMP`).
			WithArgs(refreshToken).
			WillReturnRows(rows)

		user, err := repo.GetUserByRefreshToken(ctx, refreshToken)

		require.NoError(t, err)
		assert.Equal(t, refreshToken, user.RefreshToken)
	})

	t.Run("Просроченный refresh token", func(t *testing.T) {
		mock.ExpectQuery(`SELECT * FROM users WHERE refresh_token = $1 AND refresh_token_expiry_time > CURRENT_TIMESTAMP`).
			WithArgs(expiredToken).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetUserByRefreshToken(ctx, expiredToken)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "недействительный или просроченный")
	})

	t.Run("Refresh token не найден", func(t *testing.T) {
		mock.ExpectQuery(`SELECT * FROM users WHERE refresh_token = $1 AND refresh_token_expiry_time > CURRENT_TIMESTAMP`).
			WithArgs("non_existent_token").
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetUserByRefreshToken(ctx, "non_existent_token")

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_UpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	ctx := context.Background()
	user := &models.User{
		UserID: uuid.New().String(),
		Email:  "updated@example.com",
		Role:   "Reader",
	}

	t.Run("Успешное обновление пользователя", func(t *testing.T) {
		mock.ExpectExec(`UPDATE users SET email = ?, role = ? WHERE user_id = ?`).
			WithArgs(user.Email, user.Role, user.UserID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateUser(ctx, user)

		assert.NoError(t, err)
	})

	t.Run("Пользователь не найден при обновлении", func(t *testing.T) {
		mock.ExpectExec(`UPDATE users SET email = ?, role = ? WHERE user_id = ?`).
			WithArgs(user.Email, user.Role, user.UserID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateUser(ctx, user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не найден")
	})
}

func TestUserRepository_DeleteUser(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(sqlxDB)

	ctx := context.Background()
	userID := uuid.New().String()

	t.Run("Успешное удаление пользователя", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM users WHERE user_id = $1`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteUser(ctx, userID)

		assert.NoError(t, err)
	})

	t.Run("Пользователь не найден при удалении", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM users WHERE user_id = $1`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteUser(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не найден")
	})
}

//go test ./internal/repository/... -v
