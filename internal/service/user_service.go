package service

import (
	"context"
	"microblogCPT/internal/config"
	"microblogCPT/internal/models"
	"microblogCPT/internal/repository"
)

type UserService interface {
	GetUserByIDService(ctx context.Context, userID string) (*models.User, error)
	UpdateUser(ctx context.Context, req repository.UpdateUserRequest) error
	DeleteUser(ctx context.Context, userID string) error
}

type userService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewUserService(userRepo repository.UserRepository, cfg *config.Config) UserService {
	return &userService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *userService) GetUserByIDService(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, req repository.UpdateUserRequest) error {
	// get user by id
	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	user.Email = req.Email
	user.Role = req.Role

	// update user
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	err := s.userRepo.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}
