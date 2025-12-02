// Package config предоставляет конфигурацию для сервера.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config представляет конфигурацию сервера.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig содержит настройки сервера.
type ServerConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig содержит настройки базы данных.
type DatabaseConfig struct {
	Path string
}

// JWTConfig содержит настройки JWT.
type JWTConfig struct {
	SecretKey             string
	ExpirationTime        time.Duration
	RefreshExpirationTime time.Duration
}

// Load загружает конфигурацию из переменных окружения и флагов командной строки.
func Load() (*Config, error) {
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.shutdown_timeout", "30s")
	viper.SetDefault("database.path", "gophkeeper.db")
	viper.SetDefault("jwt.expiration_time", "1h")
	viper.SetDefault("jwt.refresh_expiration_time", "168h") // 7 days

	// Читаем из переменных окружения
	viper.SetEnvPrefix("GOPHKEEPER")
	viper.AutomaticEnv()

	// Маппинг переменных окружения
	viper.BindEnv("server.address", "GOPHKEEPER_SERVER_ADDRESS")
	viper.BindEnv("server.read_timeout", "GOPHKEEPER_SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "GOPHKEEPER_SERVER_WRITE_TIMEOUT")
	viper.BindEnv("server.idle_timeout", "GOPHKEEPER_SERVER_IDLE_TIMEOUT")
	viper.BindEnv("server.shutdown_timeout", "GOPHKEEPER_SERVER_SHUTDOWN_TIMEOUT")
	viper.BindEnv("database.path", "GOPHKEEPER_DB_PATH")
	viper.BindEnv("jwt.secret_key", "GOPHKEEPER_JWT_SECRET_KEY")
	viper.BindEnv("jwt.expiration_time", "GOPHKEEPER_JWT_EXPIRATION_TIME")
	viper.BindEnv("jwt.refresh_expiration_time", "GOPHKEEPER_JWT_REFRESH_EXPIRATION_TIME")
	viper.BindEnv("master_password", "GOPHKEEPER_MASTER_PASSWORD")

	config := &Config{
		Server: ServerConfig{
			Address:         viper.GetString("server.address"),
			ReadTimeout:     parseDuration(viper.GetString("server.read_timeout")),
			WriteTimeout:    parseDuration(viper.GetString("server.write_timeout")),
			IdleTimeout:     parseDuration(viper.GetString("server.idle_timeout")),
			ShutdownTimeout: parseDuration(viper.GetString("server.shutdown_timeout")),
		},
		Database: DatabaseConfig{
			Path: viper.GetString("database.path"),
		},
		JWT: JWTConfig{
			SecretKey:             viper.GetString("jwt.secret_key"),
			ExpirationTime:        parseDuration(viper.GetString("jwt.expiration_time")),
			RefreshExpirationTime: parseDuration(viper.GetString("jwt.refresh_expiration_time")),
		},
	}

	// Валидация обязательных полей
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate проверяет корректность конфигурации.
func (c *Config) Validate() error {
	if c.JWT.SecretKey == "" {
		// В development режиме разрешаем дефолтный ключ, но предупреждаем
		if os.Getenv("GOPHKEEPER_ENV") != "production" {
			c.JWT.SecretKey = "dev-secret-key-change-in-production"
			fmt.Fprintf(os.Stderr, "WARNING: Using default JWT secret key. Set GOPHKEEPER_JWT_SECRET_KEY in production!\n")
		} else {
			return fmt.Errorf("JWT secret key is required in production. Set GOPHKEEPER_JWT_SECRET_KEY environment variable")
		}
	}

	if len(c.JWT.SecretKey) < 32 {
		return fmt.Errorf("JWT secret key must be at least 32 characters long")
	}

	return nil
}

// parseDuration парсит строку в time.Duration.
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		// Возвращаем дефолтное значение при ошибке
		return 1 * time.Hour
	}
	return d
}
