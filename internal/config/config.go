package config

import (
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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
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
	return &Config{
		ServerPort:           getEnvAsInt("SERVER_PORT", 8080),
		DB:                   LoadDB(),
		JWTSecretKey:         getEnv("JWT_SECRET_KEY", ""),
		AccessTokenDuration:  parseDuration(getEnv("ACCESS_TOKEN_DURATION", "2h")),
		RefreshTokenDuration: parseDuration(getEnv("REFRESH_TOKEN_DURATION", "168h")),
	}
}
