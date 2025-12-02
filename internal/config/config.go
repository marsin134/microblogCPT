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

type Config struct {
	ServerPort           int
	DB                   DB
	JWTSecretKey         string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
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

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		ServerPort:           getEnvAsInt("SERVER_PORT", 8080),
		DB:                   LoadDB(),
		JWTSecretKey:         getEnv("JWT_SECRET_KEY", ""),
		AccessTokenDuration:  parseDuration(getEnv("ACCESS_TOKEN_DURATION", "2h")),
		RefreshTokenDuration: parseDuration(getEnv("REFRESH_TOKEN_DURATION", "168h")),
	}
}
