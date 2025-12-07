package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"microblogCPT/internal/models"
	"time"
)

type ImageRepositoryImpl struct {
	db *sqlx.DB
}

func NewImageRepository(db *sqlx.DB) *ImageRepositoryImpl {
	return &ImageRepositoryImpl{db: db}
}

func (r *ImageRepositoryImpl) Create(ctx context.Context, image *models.Image) error {
	query := `
		INSERT INTO images (image_id, post_id, image_url, created_at)
		VALUES (:image_id, :post_id, :image_url, :created_at)
	`

	if image.ImageID == "" {
		image.ImageID = uuid.New().String()
	}

	if image.CreatedAt.IsZero() {
		image.CreatedAt = time.Now()
	}

	_, err := r.db.NamedExecContext(ctx, query, image)
	if err != nil {
		return fmt.Errorf("ошибка при создании изображения: %w", err)
	}

	return nil
}

func (r *ImageRepositoryImpl) GetByImageID(ctx context.Context, imageID string) (*models.Image, error) {
	query := `SELECT * FROM images WHERE image_id = $1 ORDER BY created_at`

	var image *models.Image
	err := r.db.SelectContext(ctx, &image, query, imageID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения изображения: %w", err)
	}

	return image, nil
}

func (r *ImageRepositoryImpl) GetByPostID(ctx context.Context, postID string) ([]*models.Image, error) {
	query := `SELECT * FROM images WHERE post_id = $1 ORDER BY created_at`

	var images []*models.Image
	err := r.db.SelectContext(ctx, &images, query, postID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении изображений: %w", err)
	}

	return images, nil
}

func (r *ImageRepositoryImpl) Delete(ctx context.Context, imageID string) error {
	query := `DELETE FROM images WHERE image_id = $1`

	result, err := r.db.ExecContext(ctx, query, imageID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении изображения: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке удаленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("изображение не найдено")
	}

	return nil
}

func (r *ImageRepositoryImpl) DeleteByPostID(ctx context.Context, postID string) error {
	query := `DELETE FROM images WHERE post_id = $1`

	_, err := r.db.ExecContext(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении изображений поста: %w", err)
	}

	return nil
}
