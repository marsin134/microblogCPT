package testRepository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"microblogCPT/internal/models"
	"microblogCPT/internal/repository"
	"testing"
	"time"
)

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	t.Cleanup(func() { sqlxDB.Close() })

	return sqlxDB, mock
}

func TestNewPostRepository(t *testing.T) {
	db, _ := setupMockDB(t)

	repo := repository.NewPostRepository(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB)
}

func TestPostRepositoryImpl_Create(t *testing.T) {
	tests := []struct {
		name        string
		post        *models.Post
		imagesURL   []string
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		errorMsg    string
	}{
		{
			name: "Успешное создание поста без изображений",
			post: &models.Post{
				PostID:         "test-post-id",
				AuthorID:       "test-author-id",
				IdempotencyKey: stringPtr("test-key"),
				Title:          "Test Title",
				Content:        "Test Content",
				Status:         "Draft",
			},
			imagesURL: []string{},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO posts`).
					WithArgs(
						"test-post-id",
						"test-author-id",
						"test-key",
						"Test Title",
						"Test Content",
						"Draft",
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectError: false,
		},
		{
			name: "Успешное создание поста с изображениями",
			post: &models.Post{
				PostID:         "test-post-id",
				AuthorID:       "test-author-id",
				IdempotencyKey: stringPtr("test-key"),
				Title:          "Test Title",
				Content:        "Test Content",
				Status:         "Draft",
			},
			imagesURL: []string{"http://example.com/image1.jpg", "http://example.com/image2.jpg"},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO posts`).
					WithArgs(
						"test-post-id",
						"test-author-id",
						"test-key",
						"Test Title",
						"Test Content",
						"Draft",
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectError: false,
		},
		{
			name: "Ошибка при дублировании idempotency key",
			post: &models.Post{
				PostID:         "test-post-id",
				AuthorID:       "test-author-id",
				IdempotencyKey: stringPtr("duplicate-key"),
				Title:          "Test Title",
				Content:        "Test Content",
				Status:         "Draft",
			},
			imagesURL: []string{},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO posts`).
					WillReturnError(fmt.Errorf("duplicate key value violates unique constraint \"posts_idempotency_key_author_id_key\""))
			},
			expectError: true,
			errorMsg:    "idempotency key уже использован",
		},
		{
			name: "Ошибка базы данных",
			post: &models.Post{
				PostID:         "test-post-id",
				AuthorID:       "test-author-id",
				IdempotencyKey: stringPtr("test-key"),
				Title:          "Test Title",
				Content:        "Test Content",
				Status:         "Draft",
			},
			imagesURL: []string{},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO posts`).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectError: true,
			errorMsg:    "ошибка при создании поста",
		},
		{
			name: "Генерация PostID если пустой",
			post: &models.Post{
				PostID:         "",
				AuthorID:       "test-author-id",
				IdempotencyKey: stringPtr("test-key"),
				Title:          "Test Title",
				Content:        "Test Content",
				Status:         "Draft",
			},
			imagesURL: []string{},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO posts`).
					WithArgs(
						sqlmock.AnyArg(), // waiting for any UUID
						"test-author-id",
						"test-key",
						"Test Title",
						"Test Content",
						"Draft",
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewPostRepository(db)

			ctx := context.Background()
			err := repo.Create(ctx, tc.post, tc.imagesURL)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tc.post.CreatedAt)
				assert.NotEmpty(t, tc.post.UpdatedAt)
				if tc.post.PostID == "" {
					_, uuidErr := uuid.Parse(tc.post.PostID)
					assert.NoError(t, uuidErr)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostRepositoryImpl_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		postID      string
		setupMock   func(mock sqlmock.Sqlmock)
		expectPost  *models.Post
		expectError bool
		errorMsg    string
	}{
		{
			name:   "Успешное получение поста",
			postID: "existing-post-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"post_id", "author_id", "idempotency_key", "title",
					"content", "status", "created_at", "updated_at",
				}).
					AddRow(
						"existing-post-id",
						"test-author-id",
						"test-key",
						"Test Title",
						"Test Content",
						"Draft",
						time.Now(),
						time.Now(),
					)
				mock.ExpectQuery(`SELECT \* FROM posts WHERE post_id = \$1`).
					WithArgs("existing-post-id").
					WillReturnRows(rows)
			},
			expectPost: &models.Post{
				PostID:         "existing-post-id",
				AuthorID:       "test-author-id",
				IdempotencyKey: stringPtr("test-key"),
				Title:          "Test Title",
				Content:        "Test Content",
				Status:         "Draft",
			},
			expectError: false,
		},
		{
			name:   "Пост не найден",
			postID: "non-existing-post-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM posts WHERE post_id = \$1`).
					WithArgs("non-existing-post-id").
					WillReturnError(sql.ErrNoRows)
			},
			expectPost:  nil,
			expectError: true,
			errorMsg:    "не найден",
		},
		{
			name:   "Ошибка базы данных",
			postID: "test-post-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM posts WHERE post_id = \$1`).
					WithArgs("test-post-id").
					WillReturnError(fmt.Errorf("database error"))
			},
			expectPost:  nil,
			expectError: true,
			errorMsg:    "ошибка при получении поста",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewPostRepository(db)

			ctx := context.Background()
			post, err := repo.GetByID(ctx, tc.postID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, post)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, post)
				assert.Equal(t, tc.expectPost.PostID, post.PostID)
				assert.Equal(t, tc.expectPost.AuthorID, post.AuthorID)
				assert.Equal(t, tc.expectPost.Title, post.Title)
				assert.Equal(t, tc.expectPost.Content, post.Content)
				assert.Equal(t, tc.expectPost.Status, post.Status)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostRepositoryImpl_Update(t *testing.T) {
	tests := []struct {
		name        string
		post        *models.Post
		setupMock   func(mock sqlmock.Sqlmock, post *models.Post)
		expectError bool
		errorMsg    string
	}{
		{
			name: "Успешное обновление поста",
			post: &models.Post{
				PostID:   "test-post-id",
				AuthorID: "test-author-id",
				Title:    "Updated Title",
				Content:  "Updated Content",
				Status:   "Published",
			},
			setupMock: func(mock sqlmock.Sqlmock, post *models.Post) {
				// Mock for GetByID
				rows := sqlmock.NewRows([]string{
					"post_id", "author_id", "idempotency_key", "title",
					"content", "status", "created_at", "updated_at",
				}).
					AddRow(
						post.PostID,
						post.AuthorID,
						"test-key",
						"Old Title",
						"Old Content",
						"Draft",
						time.Now(),
						time.Now(),
					)
				mock.ExpectQuery(`SELECT \* FROM posts WHERE post_id = \$1`).
					WithArgs(post.PostID).
					WillReturnRows(rows)

				// Mock for UPDATE
				mock.ExpectExec(`UPDATE posts SET`).
					WithArgs(
						post.Title,
						post.Content,
						post.Status,
						sqlmock.AnyArg(), // updated_at
						post.PostID,
						post.AuthorID,
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "Ошибка - пост не найден",
			post: &models.Post{
				PostID:   "non-existing-post-id",
				AuthorID: "test-author-id",
				Title:    "Updated Title",
				Content:  "Updated Content",
				Status:   "Published",
			},
			setupMock: func(mock sqlmock.Sqlmock, post *models.Post) {
				mock.ExpectQuery(`SELECT \* FROM posts WHERE post_id = \$1`).
					WithArgs(post.PostID).
					WillReturnError(fmt.Errorf("пост с ID %s не найден", post.PostID))
			},
			expectError: true,
			errorMsg:    "не найден",
		},
		{
			name: "Ошибка - изменение автора поста",
			post: &models.Post{
				PostID:   "test-post-id",
				AuthorID: "different-author-id",
				Title:    "Updated Title",
				Content:  "Updated Content",
				Status:   "Published",
			},
			setupMock: func(mock sqlmock.Sqlmock, post *models.Post) {
				rows := sqlmock.NewRows([]string{
					"post_id", "author_id", "idempotency_key", "title",
					"content", "status", "created_at", "updated_at",
				}).
					AddRow(
						post.PostID,
						"original-author-id", // Different author
						"test-key",
						"Old Title",
						"Old Content",
						"Draft",
						time.Now(),
						time.Now(),
					)
				mock.ExpectQuery(`SELECT \* FROM posts WHERE post_id = \$1`).
					WithArgs(post.PostID).
					WillReturnRows(rows)
			},
			expectError: true,
			errorMsg:    "нельзя изменить автора поста",
		},
		{
			name: "Ошибка - пост не найден при обновлении",
			post: &models.Post{
				PostID:   "test-post-id",
				AuthorID: "test-author-id",
				Title:    "Updated Title",
				Content:  "Updated Content",
				Status:   "Published",
			},
			setupMock: func(mock sqlmock.Sqlmock, post *models.Post) {
				rows := sqlmock.NewRows([]string{
					"post_id", "author_id", "idempotency_key", "title",
					"content", "status", "created_at", "updated_at",
				}).
					AddRow(
						post.PostID,
						post.AuthorID,
						"test-key",
						"Old Title",
						"Old Content",
						"Draft",
						time.Now(),
						time.Now(),
					)
				mock.ExpectQuery(`SELECT \* FROM posts WHERE post_id = \$1`).
					WithArgs(post.PostID).
					WillReturnRows(rows)

				mock.ExpectExec(`UPDATE posts SET`).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectError: true,
			errorMsg:    "пост не найден или у вас нет прав на его изменение",
		},
		{
			name: "Ошибка базы данных при обновлении",
			post: &models.Post{
				PostID:   "test-post-id",
				AuthorID: "test-author-id",
				Title:    "Updated Title",
				Content:  "Updated Content",
				Status:   "Published",
			},
			setupMock: func(mock sqlmock.Sqlmock, post *models.Post) {
				rows := sqlmock.NewRows([]string{
					"post_id", "author_id", "idempotency_key", "title",
					"content", "status", "created_at", "updated_at",
				}).
					AddRow(
						post.PostID,
						post.AuthorID,
						"test-key",
						"Old Title",
						"Old Content",
						"Draft",
						time.Now(),
						time.Now(),
					)
				mock.ExpectQuery(`SELECT \* FROM posts WHERE post_id = \$1`).
					WithArgs(post.PostID).
					WillReturnRows(rows)

				mock.ExpectExec(`UPDATE posts SET`).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectError: true,
			errorMsg:    "ошибка при обновлении поста",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			tc.setupMock(mock, tc.post)

			repo := repository.NewPostRepository(db)

			ctx := context.Background()
			err := repo.Update(ctx, tc.post)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tc.post.UpdatedAt)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostRepositoryImpl_Delete(t *testing.T) {
	tests := []struct {
		name        string
		postID      string
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "Успешное удаление поста",
			postID: "test-post-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM posts WHERE post_id = \$1`).
					WithArgs("test-post-id").
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectExec(`DELETE FROM images WHERE post_id = \$1`).
					WithArgs("test-post-id").
					WillReturnResult(sqlmock.NewResult(0, 2))
			},
			expectError: false,
		},
		{
			name:   "Ошибка при удалении изображений",
			postID: "test-post-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM posts WHERE post_id = \$1`).
					WithArgs("test-post-id").
					WillReturnError(fmt.Errorf("image deletion error"))
			},
			expectError: true,
			errorMsg:    "image deletion error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewPostRepository(db)

			ctx := context.Background()
			err := repo.Delete(ctx, tc.postID)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostRepositoryImpl_Publish(t *testing.T) {
	tests := []struct {
		name        string
		postID      string
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "Успешная публикация поста",
			postID: "test-post-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE posts SET`).
					WithArgs("test-post-id").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name:   "Пост не найден или уже опубликован",
			postID: "test-post-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE posts SET`).
					WithArgs("test-post-id").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectError: true,
			errorMsg:    "пост не найден или уже опубликован",
		},
		{
			name:   "Ошибка базы данных",
			postID: "test-post-id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE posts SET`).
					WithArgs("test-post-id").
					WillReturnError(fmt.Errorf("database error"))
			},
			expectError: true,
			errorMsg:    "ошибка при публикации поста",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewPostRepository(db)

			ctx := context.Background()
			err := repo.Publish(ctx, tc.postID)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostRepositoryImpl_CheckIdempotencyKey(t *testing.T) {
	tests := []struct {
		name            string
		authorID        string
		idempotencyKey  string
		setupMock       func(mock sqlmock.Sqlmock)
		expectedAllowed bool
		expectError     bool
		errorMsg        string
	}{
		{
			name:            "Пустой idempotency key разрешен",
			authorID:        "test-author-id",
			idempotencyKey:  "",
			setupMock:       func(mock sqlmock.Sqlmock) {},
			expectedAllowed: true,
			expectError:     false,
		},
		{
			name:           "Idempotency key не используется",
			authorID:       "test-author-id",
			idempotencyKey: "unique-key",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts WHERE author_id = \$1 AND idempotency_key = \$2`).
					WithArgs("test-author-id", "unique-key").
					WillReturnRows(rows)
			},
			expectedAllowed: true,
			expectError:     false,
		},
		{
			name:           "Idempotency key уже использован",
			authorID:       "test-author-id",
			idempotencyKey: "duplicate-key",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts WHERE author_id = \$1 AND idempotency_key = \$2`).
					WithArgs("test-author-id", "duplicate-key").
					WillReturnRows(rows)
			},
			expectedAllowed: false,
			expectError:     false,
		},
		{
			name:           "Ошибка базы данных",
			authorID:       "test-author-id",
			idempotencyKey: "test-key",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts WHERE author_id = \$1 AND idempotency_key = \$2`).
					WithArgs("test-author-id", "test-key").
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedAllowed: false,
			expectError:     true,
			errorMsg:        "ошибка при проверке idempotency key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			tc.setupMock(mock)

			repo := repository.NewPostRepository(db)

			ctx := context.Background()
			allowed, err := repo.CheckIdempotencyKey(ctx, tc.authorID, tc.idempotencyKey)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAllowed, allowed)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Auxiliary function for creating a pointer to a string
func stringPtr(s string) *string {
	return &s
}
