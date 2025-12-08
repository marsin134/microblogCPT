package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"io"
	"microblogCPT/internal/config"
	"microblogCPT/internal/models"
	"microblogCPT/internal/repository"
	"microblogCPT/internal/storage"
	"strings"
	"time"
)

type PostService interface {
	CreatePost(ctx context.Context, req repository.CreatePostRequest) (*models.Post, error)
	UpdatePost(ctx context.Context, req repository.UpdatePostRequest) error
	DeletePost(ctx context.Context, postID string) error
	PublishPost(ctx context.Context, postID string) error
	AddedImage(ctx context.Context, postID, fileName string, file io.Reader, size int64) (*models.Image, error)
	DeleteImage(ctx context.Context, imageID string) error
}

type postService struct {
	postRepo  repository.PostRepository
	imageRepo repository.ImageRepository
	storage   storage.Storage
	cfg       *config.Config
}

func (p *postService) CreatePost(ctx context.Context, req repository.CreatePostRequest) (*models.Post, error) {
	post := &models.Post{
		AuthorID:       req.AuthorID,
		IdempotencyKey: req.IdempotencyKey,
		Title:          req.Title,
		Content:        req.Content,
		Status:         "Draft",
	}

	err := p.postRepo.Create(ctx, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (p *postService) UpdatePost(ctx context.Context, req repository.UpdatePostRequest) error {
	post, err := p.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		return err
	}

	post.Title = req.Title
	post.Content = req.Content

	err = p.postRepo.Update(ctx, post)
	if err != nil {
		return err
	}

	return nil
}

func (p *postService) DeletePost(ctx context.Context, postID string) error {
	err := p.postRepo.Delete(ctx, postID)
	if err != nil {
		return err
	}

	return nil
}

func (p *postService) PublishPost(ctx context.Context, postID string) error {
	err := p.postRepo.Publish(ctx, postID)
	if err != nil {
		return err
	}
	return nil
}

func (p *postService) AddedImage(ctx context.Context, postID, fileName string, file io.Reader, size int64) (*models.Image, error) {
	objectName, imageURL, err := p.storage.UploadImage(ctx, postID, fileName, file, size)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки изображения в MinIO: %w", err)
	}

	image := &models.Image{
		ImageID:   uuid.New().String(),
		PostID:    postID,
		ImageURL:  imageURL,
		CreatedAt: time.Now(),
	}

	err = p.imageRepo.Create(ctx, image)
	if err != nil {
		p.storage.DeleteImage(ctx, objectName)
		return nil, fmt.Errorf("ошибка сохранения изображения в БД: %w", err)
	}

	return image, nil
}

func (p *postService) DeleteImage(ctx context.Context, imageID string) error {
	image, err := p.imageRepo.GetByImageID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("изображение не найдено")
	}

	urlParts := strings.Split(image.ImageURL, "/")
	if len(urlParts) < 2 {
		return fmt.Errorf("неверный формат URL изображения")
	}

	objectPath := ""

	for i, _ := range urlParts {
		if i+1 < len(urlParts) {
			objectPath = strings.Join(urlParts[i+1:], "/")
			break
		}
	}

	if objectPath == "" {
		objectPath = urlParts[len(urlParts)-1]
	}

	if err := p.storage.DeleteImage(ctx, objectPath); err != nil {
		fmt.Printf("Предупреждение: не удалось удалить из MinIO: %v\n", err)
	}

	if err := p.imageRepo.Delete(ctx, imageID); err != nil {
		return fmt.Errorf("ошибка удаления из БД: %w", err)
	}

	return nil
}
