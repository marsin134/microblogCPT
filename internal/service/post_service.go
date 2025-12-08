package service

import (
	"context"
	"microblogCPT/internal/config"
	"microblogCPT/internal/models"
	"microblogCPT/internal/repository"
)

type PostService interface {
	CreatePost(ctx context.Context, req repository.CreatePostRequest) (*models.Post, error)
	UpdatePost(ctx context.Context, req repository.UpdatePostRequest) error
	DeletePost(ctx context.Context, postID string) error
	PublishPost(ctx context.Context, postID string) error
	AddedImage(ctx context.Context, req repository.CreateImageRequest) error
	DeleteImage(ctx context.Context, imageURL string) error
}

type postService struct {
	postRepo  repository.PostRepository
	imageRepo repository.ImageRepository
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
	post, err := p.postRepo.GetByID(ctx, req.AuthorID)
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

func (p *postService) AddedImage(ctx context.Context, req repository.CreateImageRequest) error {
	// нужно будет добавить загрузку изображения или сделать это в хендлерах
	image := &models.Image{
		PostID:   req.PostID,
		ImageURL: req.ImageURL,
	}

	err := p.imageRepo.Create(ctx, image)
	if err != nil {
		return err
	}

	return nil
}

func (p *postService) DeleteImage(ctx context.Context, imageURL string) error {
	err := p.imageRepo.Delete(ctx, imageURL)
	if err != nil {
		return err
	}

	return nil
}
