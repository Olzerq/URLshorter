package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerPort string
	BaseURL    string

	// Переключатель между in memory и бд
	StorageType string

	// конфигураци для бд
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
}

// Load загружает конфиги из env/берет дефолтные
func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		BaseURL:          getEnv("BASE_URL", "http://localhost:8080"),
		StorageType:      getEnv("STORAGE_TYPE", "memory"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "shorturl"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.StorageType != "memory" && c.StorageType != "postgres" {
		return fmt.Errorf("invalid storage type: %s ", c.StorageType)
	}

	if _, err := strconv.Atoi(c.ServerPort); err != nil {
		return fmt.Errorf("invalid server port: %s", c.ServerPort)
	}

	if c.StorageType == "postgres" {
		if c.PostgresHost == "" {
			return fmt.Errorf("postgres host is required")
		}
		if c.PostgresUser == "" {
			return fmt.Errorf("postgres user is required")
		}
		if c.PostgresDB == "" {
			return fmt.Errorf("postgres db is required")
		}
	}

	return nil
}

// PostgresConnectionString возращает строчку подключения
func (c *Config) PostgresConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresDB,
	)
}

// getEnv смотрит есть ли значение в env
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
