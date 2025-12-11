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

func NewPostService(postRepo repository.PostRepository, imageRepo repository.ImageRepository, storage storage.Storage, cfg *config.Config) PostService {
	return &postService{
		postRepo:  postRepo,
		imageRepo: imageRepo,
		storage:   storage,
		cfg:       cfg,
	}
}

func (p *postService) CreatePost(ctx context.Context, req repository.CreatePostRequest) (*models.Post, error) {
	post := &models.Post{
		AuthorID:       req.AuthorID,
		IdempotencyKey: req.IdempotencyKey,
		Title:          req.Title,
		Content:        req.Content,
		Status:         "Draft",
	}

	err := p.postRepo.Create(ctx, post, []string{})
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
	// uploading an image to MinIO
	objectName, imageURL, err := p.storage.UploadImage(ctx, postID, fileName, file, size)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки изображения в MinIO: %w", err)
	}

	// create image in db
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
	// get image by id
	_, err := p.imageRepo.GetByImageID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("изображение не найдено")
	}

	// delete image in MinIO
	if err := p.storage.DeleteImage(ctx, imageID); err != nil {
		fmt.Printf("Предупреждение: не удалось удалить из MinIO: %v\n", err)
	}

	// delete image in db
	if err := p.imageRepo.Delete(ctx, imageID); err != nil {
		return fmt.Errorf("ошибка удаления из БД: %w", err)
	}

	return nil
}
