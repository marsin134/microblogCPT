package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"microblogCPT/internal/service"
)

// AuthMiddleware - middleware для проверки JWT токена
func AuthMiddleware(authService service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Исключаем публичные эндпоинты
			if r.URL.Path == "/api/auth/register" ||
				r.URL.Path == "/api/auth/login" ||
				r.URL.Path == "/api/auth/refresh-token" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, "Требуется аутентификация", http.StatusUnauthorized)
				return
			}

			// Проверяем формат "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, "Неверный формат токена", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Валидируем токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte("your-secret-key"), nil // TODO: взять из конфига
			})

			if err != nil || !token.Valid {
				writeError(w, "Недействительный токен", http.StatusUnauthorized)
				return
			}

			// Извлекаем claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeError(w, "Неверный формат токена", http.StatusUnauthorized)
				return
			}

			// Проверяем срок действия
			if exp, ok := claims["exp"].(float64); ok {
				if int64(exp) < jwt.TimeFunc().Unix() {
					writeError(w, "Токен истек", http.StatusUnauthorized)
					return
				}
			}

			// Добавляем информацию о пользователе в контекст
			ctx := context.WithValue(r.Context(), "userID", claims["userId"].(string))
			ctx = context.WithValue(ctx, "userEmail", claims["email"].(string))
			ctx = context.WithValue(ctx, "userRole", claims["role"].(string))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthorOnlyMiddleware - middleware для проверки роли Автор
func AuthorOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := r.Context().Value("userRole").(string)
		if !ok || userRole != "Author" {
			writeError(w, "Доступ запрещен. Требуется роль Автор", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RoleMiddleware - middleware для проверки роли
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value("userRole").(string)
			if !ok {
				http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
				return
			}

			// Проверяем, есть ли роль пользователя в списке разрешенных
			allowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Доступ запрещен", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware - middleware для CORS
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

// LoggingMiddleware - middleware для логирования запросов
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Логируем запрос
		// log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
