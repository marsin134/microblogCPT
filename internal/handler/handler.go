package handlers

import (
	"github.com/go-playground/validator/v10"
	"microblogCPT/internal/config"
	"microblogCPT/internal/repository"
	"microblogCPT/internal/service"
)

type Handlers struct {
	UserService   service.UserService
	UserRepo      repository.UserRepository
	AuthService   service.AuthService
	PostService   service.PostService
	PostRepo      repository.PostRepository
	TablesRepo    repository.TablesRepository
	TablesService service.TablesService
	Cfg           *config.Config
	Validate      *validator.Validate
}

func NewHandlers(repo *repository.Repository, service *service.Service, config *config.Config) *Handlers {
	return &Handlers{
		UserService:   service.User,
		UserRepo:      repo.User,
		AuthService:   service.Auth,
		PostService:   service.Post,
		PostRepo:      repo.Post,
		TablesRepo:    repo.Tables,
		TablesService: service.Tables,
		Cfg:           config,
		Validate:      validator.New(),
	}
}
