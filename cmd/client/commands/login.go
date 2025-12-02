package commands

import (
	"context"
	"fmt"
	"log"
)

// LoginCommand реализует команду входа пользователя.
type LoginCommand struct {
	*BaseCommand
}

// NewLoginCommand создает новую команду входа.
func NewLoginCommand(base *BaseCommand) *LoginCommand {
	return &LoginCommand{BaseCommand: base}
}

// Name возвращает имя команды.
func (c *LoginCommand) Name() string {
	return "login"
}

// Usage возвращает строку использования.
func (c *LoginCommand) Usage() string {
	return "login <login> <password>"
}

// Description возвращает описание команды.
func (c *LoginCommand) Description() string {
	return "Login user"
}

// Execute выполняет команду входа.
func (c *LoginCommand) Execute(ctx context.Context, args []string) error {
	if err := c.ValidateArgs(args, 2, c.Usage()); err != nil {
		return err
	}

	login := args[0]
	password := args[1]

	resp, err := c.Client.Login(ctx, login, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if err := c.LocalStorage.SaveTokens(ctx, resp.Token, resp.RefreshToken); err != nil {
		log.Fatalf("failed to save tokens: %v", err)
	}

	fmt.Println("Login successful!")
	return nil
}

