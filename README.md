# Microblog for CPT API

## О проекте:

Микроблог-платформа с ролевой системой (Author/Reader), JWT-аутентификацией и загрузкой изображений в MinIO, **созданная
в
процессе отбора в ЦПТ ПГУ.**

# Установка и запуск

### Переменные окружения *(.env)*

```
SERVER_PORT=8080
JWT_SECRET_KEY=your_super_secret_key_here
JWT_ACCESS_EXPIRATION=2h
JWT_REFRESH_EXPIRATION=168h

# База данных (PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=microblog

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET_NAME=images
MINIO_USE_SSL=false

# Загрузка файлов
MAX_UPLOAD_SIZE=10485760  # 10 MB
```

### Запуск

```
# Установка зависимостей
go mod download

# Запуск приложения
go run main.go
```

## API Эндпоинты

| Method | URL                              | Description          | Requires a token | Roles         |
|--------|----------------------------------|----------------------|------------------|---------------|
| POST   | /api/auth/register               | Регистрация          | No               | All           |
| POST   | /api/auth/login                  | Вход                 | No               | All           |
| POST   | /api/auth/refresh-token          | Обновление токена    | No               | All           |
| GET    | /api/me                          | Текущий пользователь | Yes              | Author/Reader |
| GET    | /api/user/{id}                   | Пользователь по ID   | Yes              | Author/Reader |
| GET    | /api/posts                       | Все посты            | Yes              | All           |
| POST   | /api/posts                       | Создать пост         | Yes              | Author        |
| PUT    | /api/posts/{id}                  | Обновить пост        | Yes              | Author        |
| PATCH  | /api/posts/{id}/status           | Публикация поста     | Yes              | Author        |
| POST   | /api/posts/{id}/images           | Добавить изображение | Yes              | Author        |
| DELETE | /api/posts/{id}/images/{imageId} | Удалить изображение  | Yes              | Author        |
| GET    | /health                          | Статус сервера       | No               | All           |
| GET    | /tables                          | Таблицы БД           | No               | All           |
| Get    | /                                | Документация API     | No               | All           |

# Примеры запросов

### Регистрация пользователя

```
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "role": "Author"
  }'
  ```

### Вход в систему

```
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
  ```

### Создание поста (с токеном)

```
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "title": "Мой первый пост",
    "content": "Содержание поста..."
  }'
  ```

### Добавление изображения

```
curl -X POST http://localhost:8080/api/posts/123/images \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "image=@/path/to/image.jpg"
  ```

# Авторизация

### Формат заголовка

```
Authorization: Bearer <ваш_jwt_токен>
```

### Роли пользователей

- Author — может создавать, редактировать и публиковать посты

- Reader — может только просматривать опубликованные посты

# Особенности реализации

При создании постов поддерживается параметр idempotencyKey для предотвращения дублирования запросов.

### Валидация

- Email: стандартный формат email

- Пароль: минимум 6 символов

- Роли: только "Author" или "Reader"

- Изображения: JPEG, PNG, GIF, WebP, максимум 10 MB

# Мониторинг

- Приложение: http://localhost:8080

- MinIO Console: http://localhost:9001

- База данных: localhost:5432

# Примечания

1. Все POST/PUT/PATCH запросы ожидают JSON в теле запроса

2. Для загрузки файлов используйте multipart/form-data

3. Access Token истекает через 2 часа, Refresh Token — через 7 дней

4. После запуска сервера документация доступна по корневому пути /

## Тесты 

```
# handlers
go test ./internal/handler/test... -v

# repository
go test ./internal/repository/testRepository/... -v
```
