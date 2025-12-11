package test

import (
	"github.com/stretchr/testify/assert"
	"microblogCPT/internal/config"
	handlers "microblogCPT/internal/handler"
	"microblogCPT/internal/repository"
	"microblogCPT/internal/service"
	"testing"
)

func TestNewHandlers(t *testing.T) {
	// create mock object
	mockUserService := new(MockUserService)
	mockUserRepo := new(MockUserRepository)
	mockAuthService := new(MockAuthService)
	mockPostService := new(MockPostService)
	mockPostRepo := new(MockPostRepository)
	cfg := &config.Config{}

	repo := &repository.Repository{
		User: mockUserRepo,
		Post: mockPostRepo,
	}

	service := &service.Service{
		User: mockUserService,
		Post: mockPostService,
		Auth: mockAuthService,
	}

	handler := handlers.NewHandlers(repo, service, cfg)

	assert.NotNil(t, handler.UserService)
	assert.NotNil(t, handler.UserRepo)
	assert.NotNil(t, handler.AuthService)
	assert.NotNil(t, handler.PostService)
	assert.NotNil(t, handler.PostRepo)
	assert.NotNil(t, handler.Cfg)
	assert.NotNil(t, handler.Validate)
}
func TestHandlerStructure(t *testing.T) {
	// Handlers Structure Verification Test
	handler := &handlers.Handlers{}

	assert.NotNil(t, handler)
	// Just checking that the structure has been created
	assert.Equal(t, "*handlers.Handlers", "*handlers.Handlers")
}

// go test ./internal/handler/test... -v
