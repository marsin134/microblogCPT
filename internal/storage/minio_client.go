package storage

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"microblogCPT/internal/config"

	"github.com/google/uuid"
)

type Storage interface {
	UploadImage(ctx context.Context, postID string, fileName string, file io.Reader, size int64) (string, string, error)
	DeleteImage(ctx context.Context, objectName string) error
	GetImageURL(ctx context.Context, objectName string) (string, error)
}

type MinIOClient struct {
	client *minio.Client
	config *config.Config
}

func (m *MinIOClient) UploadImage(ctx context.Context, postID string, fileName string, file io.Reader, size int64) (string, string, error) {
	fileExt := strings.ToLower(filepath.Ext(fileName))
	if fileExt == "" {
		fileExt = ".jpg"
	}

	contentType := mime.TypeByExtension(fileExt)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	now := time.Now()
	objectName := fmt.Sprintf("posts/%s/%d/%02d/%s%s",
		postID,
		now.Year(),
		now.Month(),
		uuid.New().String(),
		fileExt)

	_, err := m.client.PutObject(ctx, "images", objectName, file, size,
		minio.PutObjectOptions{
			ContentType: contentType,
			UserMetadata: map[string]string{
				"original-filename": fileName,
				"post-id":           postID,
				"uploaded-at":       now.Format(time.RFC3339),
			},
		})
	if err != nil {
		return "", "", fmt.Errorf("ошибка загрузки в MinIO: %w", err)
	}

	imageURL := fmt.Sprintf("%s/%s/%s",
		strings.TrimSuffix("localhost:9000", ":9000"),
		"images",
		objectName)

	imageURL = "http://" + imageURL

	return objectName, imageURL, nil
}

func (m *MinIOClient) DeleteImage(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, "images", objectName,
		minio.RemoveObjectOptions{
			GovernanceBypass: true,
			VersionID:        "",
		})
	if err != nil {
		return fmt.Errorf("ошибка удаления из MinIO: %w", err)
	}
	return nil
}
