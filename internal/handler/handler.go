package handlers

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"microblogCPT/internal/config"
	"microblogCPT/internal/service"
	"net/http"
	"time"
)

type Handlers struct {
	UserService   service.UserService
	AuthService   service.AuthService
	PostService   service.PostService
	TablesService service.TablesService
	Cfg           *config.Config
	Validate      *validator.Validate
}

func NewHandlers(service *service.Service, config *config.Config) *Handlers {
	return &Handlers{
		UserService:   service.User,
		AuthService:   service.Auth,
		PostService:   service.Post,
		TablesService: service.Tables,
		Cfg:           config,
		Validate:      validator.New(),
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "internal/handler/html/userDocumentation.html")
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "ok", "service": "microblog", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
}
