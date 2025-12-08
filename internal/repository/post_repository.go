package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"microblogCPT/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostRepositoryImpl struct {
	db *sqlx.DB
}

type CreatePostRequest struct {
	AuthorID       string  `json:"author_id"`
	IdempotencyKey *string `json:"idempotency_key"`
	Title          string  `json:"title"`
	Content        string  `json:"content"`
}

type UpdatePostRequest struct {
	PostID  string `json:"post_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func NewPostRepository(db *sqlx.DB) *PostRepositoryImpl {
	return &PostRepositoryImpl{db: db}
}

func (r *PostRepositoryImpl) Create(ctx context.Context, post *models.Post, imagesURL []string) error {
	query := `
        INSERT INTO posts 
        (post_id, author_id, idempotency_key, title, content, status, created_at, updated_at)
        VALUES 
        (:post_id, :author_id, :idempotency_key, :title, :content, :status, :created_at, :updated_at)
    `

	imageRepositoryImpl := ImageRepositoryImpl{db: r.db}
	for _, imageURL := range imagesURL {

		image := models.Image{
			PostID:   post.PostID,
			ImageURL: imageURL,
		}

		imageRepositoryImpl.Create(ctx, &image)
	}

	if post.PostID == "" {
		post.PostID = uuid.New().String()
	}

	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	_, err := r.db.NamedExecContext(ctx, query, post)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") &&
			strings.Contains(err.Error(), "idempotency_key") {
			return fmt.Errorf("idempotency key уже использован: %w", err)
		}
		return fmt.Errorf("ошибка при создании поста: %w", err)
	}

	return nil
}

func (r *PostRepositoryImpl) GetByID(ctx context.Context, postID string) (*models.Post, error) {
	query := `
        SELECT * FROM posts 
        WHERE post_id = $1
    `

	var post models.Post
	err := r.db.GetContext(ctx, &post, query, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пост с ID %s не найден", postID)
		}
		return nil, fmt.Errorf("ошибка при получении поста: %w", err)
	}

	return &post, nil
}

func (r *PostRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]models.Post, error) {
	query := `
        SELECT * FROM posts 
        WHERE author_id = $1
    `

	var posts []models.Post
	err := r.db.GetContext(ctx, &posts, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пост пользователя %s не найден", userID)
		}
		return nil, fmt.Errorf("ошибка при получении поста: %w", err)
	}

	return posts, nil
}

func (r *PostRepositoryImpl) GetPublishPosts(ctx context.Context) ([]models.Post, error) {
	query := `
        SELECT * FROM posts 
        WHERE status = 'Published'
    `

	var posts []models.Post
	err := r.db.GetContext(ctx, &posts, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("пост пользователя %s не найден")
		}
		return nil, fmt.Errorf("ошибка при получении поста: %w", err)
	}

	return posts, nil
}

func (r *PostRepositoryImpl) Update(ctx context.Context, post *models.Post) error {
	existingPost, err := r.GetByID(ctx, post.PostID)
	if err != nil {
		return err
	}

	if existingPost.AuthorID != post.AuthorID {
		return errors.New("нельзя изменить автора поста")
	}

	query := `
		UPDATE posts SET
			title = :title,
			content = :content,
			status = :status,
			updated_at = :updated_at
		WHERE post_id = :post_id AND author_id = :author_id
	`

	post.UpdatedAt = time.Now()

	result, err := r.db.NamedExecContext(ctx, query, post)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении поста: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("пост не найден или у вас нет прав на его изменение")
	}

	return nil
}

func (r *PostRepositoryImpl) Delete(ctx context.Context, postID string) error {
	query := `DELETE FROM posts WHERE post_id = $1`

	result, err := r.db.ExecContext(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении поста: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке удаленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("пост не найден")
	}

	imageRepositoryImpl := ImageRepositoryImpl{r.db}
	err = imageRepositoryImpl.DeleteByPostID(ctx, postID)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostRepositoryImpl) Publish(ctx context.Context, postID string) error {
	query := `
		UPDATE posts SET
			status = 'Published',
			updated_at = CURRENT_TIMESTAMP
		WHERE post_id = $1 AND status = 'Draft'
	`

	result, err := r.db.ExecContext(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("ошибка при публикации поста: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("пост не найден или уже опубликован")
	}

	return nil
}

func (r *PostRepositoryImpl) CheckIdempotencyKey(ctx context.Context, authorID, idempotencyKey string) (bool, error) {
	if idempotencyKey == "" {
		return true, nil
	}

	query := `
		SELECT COUNT(*) FROM posts 
		WHERE author_id = $1 AND idempotency_key = $2
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, authorID, idempotencyKey)
	if err != nil {
		return false, fmt.Errorf("ошибка при проверке idempotency key: %w", err)
	}

	return count == 0, nil
}
