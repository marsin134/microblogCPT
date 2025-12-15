package service

import (
	"microblogCPT/internal/config"
	"microblogCPT/internal/repository"
	"microblogCPT/internal/storage"
	"net/http"
)

type Service struct {
	User   UserService
	Post   PostService
	Auth   AuthService
	Tables TablesService
}

func NewService(rep *repository.Repository, cfg *config.Config, storage storage.Storage) *Service {
	return &Service{
		User:   NewUserService(rep.User, cfg),
		Post:   NewPostService(rep.Post, rep.Image, storage, cfg),
		Auth:   NewAuthService(rep.User, cfg),
		Tables: NewTablesService(rep.Tables),
	}
}

type Mux struct {
	Mux *http.ServeMux
}

func CreateMux() *Mux {
	return &Mux{
		Mux: http.NewServeMux(),
	}
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Mux.ServeHTTP(w, r)
}
