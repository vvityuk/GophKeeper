package commands

import (
	"context"
	"fmt"
)

// LogoutCommand реализует команду выхода пользователя.
type LogoutCommand struct {
	*BaseCommand
}

// NewLogoutCommand создает новую команду выхода.
func NewLogoutCommand(base *BaseCommand) *LogoutCommand {
	return &LogoutCommand{BaseCommand: base}
}

// Name возвращает имя команды.
func (c *LogoutCommand) Name() string {
	return "logout"
}

// Usage возвращает строку использования.
func (c *LogoutCommand) Usage() string {
	return "logout"
}

// Description возвращает описание команды.
func (c *LogoutCommand) Description() string {
	return "Logout user"
}

// Execute выполняет команду выхода.
func (c *LogoutCommand) Execute(ctx context.Context, args []string) error {
	if err := c.Client.Logout(ctx); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	// Удаляем токены из локального хранилища
	if _, _, err := c.LocalStorage.GetTokens(ctx); err == nil {
		c.LocalStorage.SaveTokens(ctx, "", "")
	}

	fmt.Println("Logout successful!")
	return nil
}

