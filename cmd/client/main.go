// Package main содержит точку входа клиентского приложения GophKeeper.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/victor/gophkeeper/cmd/client/commands"
	"github.com/victor/gophkeeper/internal/client/api"
	"github.com/victor/gophkeeper/internal/client/config"
	"github.com/victor/gophkeeper/internal/client/storage"
)

func main() {
	// Параметры командной строки
	serverURL := flag.String("server", "", "server URL (overrides GOPHKEEPER_CLIENT_SERVER_URL)")
	flag.Parse()

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Переопределяем из флагов, если указаны
	if *serverURL != "" {
		cfg.Server.URL = *serverURL
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println(getUsage())
		os.Exit(1)
	}

	commandName := args[0]
	commandArgs := args[1:]

	// Инициализируем зависимости
	client, localStorage, err := initializeDependencies(cfg.Server.URL)
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}
	defer localStorage.Close()

	// Создаем реестр команд и регистрируем все команды
	registry := createCommandRegistry(client, localStorage, cfg.MasterPassword)

	// Получаем команду
	cmd, err := registry.Get(commandName)
	if err != nil {
		fmt.Printf("%s\n", err)
		fmt.Println(getUsage())
		os.Exit(1)
	}

	// Выполняем команду
	ctx := context.Background()
	if err := cmd.Execute(ctx, commandArgs); err != nil {
		log.Fatalf("command failed: %v", err)
	}
}

// initializeDependencies инициализирует клиент API и локальное хранилище.
func initializeDependencies(serverURL string) (*api.Client, *storage.LocalStorage, error) {
	// Инициализируем локальное хранилище
	homeDir, err := getHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, ".gophkeeper", "client.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	localStorage, err := storage.NewLocalStorage(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize local storage: %w", err)
	}

	// Создаем API клиент
	client := api.NewClient(serverURL)

	// Загружаем сохраненные токены
	token, refreshToken, err := localStorage.GetTokens(context.Background())
	if err == nil {
		client.SetToken(token)
		client.SetRefreshToken(refreshToken)
	}

	return client, localStorage, nil
}

// createCommandRegistry создает реестр команд и регистрирует все команды.
func createCommandRegistry(client *api.Client, localStorage *storage.LocalStorage, masterPassword string) *commands.Registry {
	registry := commands.NewRegistry()
	baseCmd := commands.NewBaseCommand(client, localStorage)

	// Регистрируем все команды
	registry.Register(commands.NewRegisterCommand(baseCmd))
	registry.Register(commands.NewLoginCommand(baseCmd))
	registry.Register(commands.NewLogoutCommand(baseCmd))
	registry.Register(commands.NewListCommand(baseCmd))
	registry.Register(commands.NewGetCommand(baseCmd, masterPassword))
	registry.Register(commands.NewAddCommand(baseCmd))
	registry.Register(commands.NewUpdateCommand(baseCmd))
	registry.Register(commands.NewDeleteCommand(baseCmd))
	registry.Register(commands.NewSyncCommand(baseCmd))
	registry.Register(commands.NewVersionCommand(baseCmd))

	return registry
}

// getUsage возвращает строку использования всех команд.
func getUsage() string {
	return `GophKeeper CLI Client

Commands:
  register <login> <password>  - Register new user
  login <login> <password>     - Login user
  logout                       - Logout user
  list                         - List all data records
  get <id>                     - Get data record by ID
  add <type> <name> <data> [metadata] - Add new data record
  update <id> <data>           - Update data record
  delete <id>                  - Delete data record
  sync                         - Sync with server
  version                      - Show version and build date

Data types: CREDENTIALS, TEXT, BINARY, CARD`
}

// getHomeDir возвращает домашнюю директорию пользователя.
func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}
