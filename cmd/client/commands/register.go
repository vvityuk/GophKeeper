package commands

import (
	"context"
	"fmt"
	"log"
)

// RegisterCommand реализует команду регистрации пользователя.
type RegisterCommand struct {
	*BaseCommand
}

// NewRegisterCommand создает новую команду регистрации.
func NewRegisterCommand(base *BaseCommand) *RegisterCommand {
	return &RegisterCommand{BaseCommand: base}
}

// Name возвращает имя команды.
func (c *RegisterCommand) Name() string {
	return "register"
}

// Usage возвращает строку использования.
func (c *RegisterCommand) Usage() string {
	return "register <login> <password>"
}

// Description возвращает описание команды.
func (c *RegisterCommand) Description() string {
	return "Register new user"
}

// Execute выполняет команду регистрации.
func (c *RegisterCommand) Execute(ctx context.Context, args []string) error {
	if err := c.ValidateArgs(args, 2, c.Usage()); err != nil {
		return err
	}

	login := args[0]
	password := args[1]

	resp, err := c.Client.Register(ctx, login, password)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	if err := c.LocalStorage.SaveTokens(ctx, resp.Token, resp.RefreshToken); err != nil {
		log.Fatalf("failed to save tokens: %v", err)
	}

	fmt.Println("Registration successful!")
	return nil
}

