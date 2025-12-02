package commands

import (
	"context"
	"fmt"

	"github.com/victor/gophkeeper/pkg/version"
)

// VersionCommand реализует команду вывода версии.
type VersionCommand struct {
	*BaseCommand
}

// NewVersionCommand создает новую команду версии.
func NewVersionCommand(base *BaseCommand) *VersionCommand {
	return &VersionCommand{BaseCommand: base}
}

// Name возвращает имя команды.
func (c *VersionCommand) Name() string {
	return "version"
}

// Usage возвращает строку использования.
func (c *VersionCommand) Usage() string {
	return "version"
}

// Description возвращает описание команды.
func (c *VersionCommand) Description() string {
	return "Show version and build date"
}

// Execute выполняет команду вывода версии.
func (c *VersionCommand) Execute(ctx context.Context, args []string) error {
	fmt.Println(version.GetInfo())
	return nil
}
