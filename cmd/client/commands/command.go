// Package commands содержит реализации команд CLI клиента.
package commands

import (
	"context"
	"fmt"

	"github.com/victor/gophkeeper/internal/client/api"
	"github.com/victor/gophkeeper/internal/client/storage"
)

// Command представляет интерфейс команды CLI.
type Command interface {
	// Name возвращает имя команды.
	Name() string
	// Usage возвращает строку использования команды.
	Usage() string
	// Description возвращает описание команды.
	Description() string
	// Execute выполняет команду с переданными аргументами.
	Execute(ctx context.Context, args []string) error
}

// BaseCommand содержит общие зависимости для всех команд.
type BaseCommand struct {
	Client       *api.Client
	LocalStorage *storage.LocalStorage
}

// NewBaseCommand создает новую базовую команду.
func NewBaseCommand(client *api.Client, localStorage *storage.LocalStorage) *BaseCommand {
	return &BaseCommand{
		Client:       client,
		LocalStorage: localStorage,
	}
}

// ValidateArgs проверяет минимальное количество аргументов.
func (b *BaseCommand) ValidateArgs(args []string, minArgs int, usage string) error {
	if len(args) < minArgs {
		return fmt.Errorf("usage: %s", usage)
	}
	return nil
}
