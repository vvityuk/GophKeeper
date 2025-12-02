// Package config предоставляет конфигурацию для клиента.
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config представляет конфигурацию клиента.
type Config struct {
	Server         ServerConfig
	MasterPassword string
}

// ServerConfig содержит настройки сервера.
type ServerConfig struct {
	URL string
}

// Load загружает конфигурацию из переменных окружения и флагов командной строки.
func Load() (*Config, error) {
	viper.SetDefault("server.url", "http://localhost:8080")

	// Читаем из переменных окружения
	viper.SetEnvPrefix("GOPHKEEPER_CLIENT")
	viper.AutomaticEnv()

	// Маппинг переменных окружения
	viper.BindEnv("server.url", "GOPHKEEPER_CLIENT_SERVER_URL")
	viper.BindEnv("master_password", "GOPHKEEPER_CLIENT_MASTER_PASSWORD")

	config := &Config{
		Server: ServerConfig{
			URL: viper.GetString("server.url"),
		},
		MasterPassword: viper.GetString("master_password"),
	}

	// Валидация
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate проверяет корректность конфигурации.
func (c *Config) Validate() error {
	if c.MasterPassword == "" {
		// В development режиме разрешаем дефолтный пароль, но предупреждаем
		if os.Getenv("GOPHKEEPER_ENV") != "production" {
			c.MasterPassword = "dev-master-password"
			fmt.Fprintf(os.Stderr, "WARNING: Using default master password. Set GOPHKEEPER_CLIENT_MASTER_PASSWORD in production!\n")
		} else {
			return fmt.Errorf("master password is required in production. Set GOPHKEEPER_CLIENT_MASTER_PASSWORD environment variable")
		}
	}

	return nil
}
