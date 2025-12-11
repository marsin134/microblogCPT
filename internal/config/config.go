package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

type DB struct {
	DbHOST     string
	DbPORT     string
	DbUSER     string
	DbPASSWORD string
	DbNAME     string
	DbSSLMODE  string
}

type MinIO struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
	Region     string
	URLExpiry  time.Duration
}

type Config struct {
	ServerPort           int
	DB                   DB
	MinIO                MinIO
	JWTSecretKey         string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	MaxUploadSize        int64
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 2 * time.Hour
	}
	return duration
}

func LoadDB() DB {
	return DB{
		DbHOST:     getEnv("DB_HOST", "localhost"),
		DbPORT:     getEnv("DB_PORT", "5432"),
		DbUSER:     getEnv("DB_USER", "postgres"),
		DbPASSWORD: getEnv("DB_PASSWORD", "password"),
		DbNAME:     getEnv("DB_NAME", "microblog"),
		DbSSLMODE:  getEnv("DB_SSLMODE", "disable"),
	}
}

func LoadMinIO() MinIO {
	return MinIO{
		Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:  getEnv("MINI	O_SECRET_KEY", "minioadmin"),
		BucketName: getEnv("MINIO_BUCKET_NAME", "images"),
		UseSSL:     getEnvBool("MINIO_USE_SSL", false),
		Region:     getEnv("MINIO_REGION", "us-east-1"),
		URLExpiry:  parseDuration(getEnv("MINIO_URL_EXPIRY", "7d")),
	}
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		ServerPort:           getEnvAsInt("SERVER_PORT", 8080),
		DB:                   LoadDB(),
		MinIO:                LoadMinIO(),
		JWTSecretKey:         getEnv("JWT_SECRET_KEY", ""),
		AccessTokenDuration:  parseDuration(getEnv("ACCESS_TOKEN_DURATION", "2h")),
		RefreshTokenDuration: parseDuration(getEnv("REFRESH_TOKEN_DURATION", "168h")),
		MaxUploadSize:        parseMaxUploadSize(getEnv("MAX_UPLOAD_SIZE", "10485760")),
	}
}

func parseMaxUploadSize(value string) int64 {
	size, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 10 * 1024 * 1024
	}
	return size
}
