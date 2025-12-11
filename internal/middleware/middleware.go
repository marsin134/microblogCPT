package middleware

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"microblogCPT/internal/config"
	handlers "microblogCPT/internal/handler"
	"microblogCPT/internal/service"
	"net/http"
	"strings"
)

type Middlewares interface {
	AuthMiddleware(authService service.AuthService, next http.Handler) http.Handler
	AuthorOnlyMiddleware(next http.Handler) http.Handler
	RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler
	CORSMiddleware(next http.Handler) http.Handler
	LoggingMiddleware(next http.Handler) http.Handler
}

type Middleware func(http.Handler) http.Handler

// AuthMiddleware verifies the JWT token and adds user data to the context
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skipping public endpoints
			publicPaths := []string{
				"/api/auth/register",
				"/api/auth/login",
				"/api/auth/refresh-token",
				"/health",
				"/tables",
				"/",
			}

			for _, path := range publicPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extracting the token from the header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handlers.WriteError(w, "Требуется авторизация", http.StatusUnauthorized)
				return
			}

			// Checking the "Bearer <token>" format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				handlers.WriteError(w, "Неверный формат токена", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Parse token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Checking the signature algorithm
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
				}
				return []byte(cfg.JWTSecretKey), nil
			})

			if err != nil {
				handlers.WriteError(w, "Недействительный токен: "+err.Error(), http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				handlers.WriteError(w, "Недействительный токен", http.StatusUnauthorized)
				return
			}

			// Extracting claims
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				userID, ok1 := claims["user_id"].(string)
				email, ok2 := claims["email"].(string)
				role, ok3 := claims["role"].(string)

				if !ok1 || !ok2 || !ok3 {
					handlers.WriteError(w, "Неверные данные в токене", http.StatusUnauthorized)
					return
				}

				// Adding user data to the context
				ctx := r.Context()
				ctx = context.WithValue(ctx, "userID", userID)
				ctx = context.WithValue(ctx, "email", email)
				ctx = context.WithValue(ctx, "role", role)

				// Passing the updated context on
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				handlers.WriteError(w, "Неверные claims токена", http.StatusUnauthorized)
			}
		})
	}
}

func AuthorOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := r.Context().Value("role").(string)
		if !ok || userRole != "Author" {
			handlers.WriteError(w, "Доступ запрещен. Требуется роль Автор", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// if strings.HasPrefix(r.URL.Path, "/") {
			//     next.ServeHTTP(w, r)
			// }

			userRole, ok := r.Context().Value("role").(string)
			if !ok {
				handlers.WriteError(w, "Требуется авторизация", http.StatusUnauthorized)
				return
			}

			// Checking if the user's role is in the allowed list
			allowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					allowed = true
					break
				}
			}

			if !allowed {
				handlers.WriteError(w, "Доступ запрещен", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Method: %s, URl: %s\nBody: %s\nContext: %s\n\n", r.Method, r.RequestURI, r.Body, r.Context())
		next.ServeHTTP(w, r)
	})
}

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
