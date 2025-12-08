package service

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"microblogCPT/internal/config"
	"microblogCPT/internal/models"
	"microblogCPT/internal/repository"
	"time"
)

type AuthService interface {
	Register(ctx context.Context, req repository.CreateUserRequest) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.User, string, string, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*models.User, string, string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	GetUserFromToken(tokenString string) (*models.User, error)
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *authService) Register(ctx context.Context, req repository.CreateUserRequest) (*models.User, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("пользователь с email %s уже существует", req.Email)
	}

	refreshToken, refreshTokenExpiry, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации refresh token: %w", err)
	}

	user := &models.User{
		Email:                  req.Email,
		Role:                   req.Role,
		RefreshToken:           refreshToken,
		RefreshTokenExpiryTime: refreshTokenExpiry,
	}

	err = s.userRepo.CreateUser(ctx, user, req.Password)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*models.User, string, string, error) {
	user, err := s.userRepo.VerifyPassword(ctx, email, password)
	if err != nil {
		return nil, "", "", fmt.Errorf("ошибка аутентификации: %w", err)
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("ошибка генерации access token: %w", err)
	}

	refreshToken, refreshTokenExpiry, err := s.generateRefreshToken()
	if err != nil {
		return nil, "", "", fmt.Errorf("ошибка генерации refresh token: %w", err)
	}

	err = s.userRepo.UpdateRefreshToken(ctx, user.UserID, refreshToken, refreshTokenExpiry)
	if err != nil {
		return nil, "", "", fmt.Errorf("ошибка сохранения refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) RefreshTokens(ctx context.Context, refreshToken string) (*models.User, string, string, error) {
	user, err := s.userRepo.GetUserByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, "", "", fmt.Errorf("недействительный refresh token: %w", err)
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("ошибка генерации access token: %w", err)
	}

	newRefreshToken, refreshTokenExpiry, err := s.generateRefreshToken()
	if err != nil {
		return nil, "", "", fmt.Errorf("ошибка генерации refresh token: %w", err)
	}

	err = s.userRepo.UpdateRefreshToken(ctx, user.UserID, newRefreshToken, refreshTokenExpiry)
	if err != nil {
		return nil, "", "", fmt.Errorf("ошибка обновления refresh token: %w", err)
	}

	return user, accessToken, newRefreshToken, nil
}

func (s *authService) generateAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"userId": user.UserID,
		"email":  user.Email,
		"role":   user.Role,
		"exp":    time.Now().Add(s.cfg.AccessTokenDuration).Unix(),
		"iat":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecretKey))
	if err != nil {
		return "", fmt.Errorf("ошибка подписи токена: %w", err)
	}

	return tokenString, nil
}

func (s *authService) generateRefreshToken() (string, time.Time, error) {
	refreshToken := uuid.New().String()

	expiryTime := time.Now().Add(s.cfg.RefreshTokenDuration)

	return refreshToken, expiryTime, nil
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга токена: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("недействительный токен")
	}

	return token, nil
}

func (s *authService) GetUserFromToken(tokenString string) (*models.User, error) {
	token, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("неверный формат claims")
	}

	user := &models.User{
		UserID: claims["userId"].(string),
		Email:  claims["email"].(string),
		Role:   claims["role"].(string),
	}

	return user, nil
}
