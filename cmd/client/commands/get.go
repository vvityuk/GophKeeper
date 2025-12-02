package commands

import (
	"context"
	"fmt"

	"github.com/victor/gophkeeper/internal/client/crypto"
)

// GetCommand реализует команду получения записи по ID.
type GetCommand struct {
	*BaseCommand
	masterPassword string
}

// NewGetCommand создает новую команду получения записи.
func NewGetCommand(base *BaseCommand, masterPassword string) *GetCommand {
	return &GetCommand{
		BaseCommand:    base,
		masterPassword: masterPassword,
	}
}

// Name возвращает имя команды.
func (c *GetCommand) Name() string {
	return "get"
}

// Usage возвращает строку использования.
func (c *GetCommand) Usage() string {
	return "get <id>"
}

// Description возвращает описание команды.
func (c *GetCommand) Description() string {
	return "Get data record by ID"
}

// Execute выполняет команду получения записи.
func (c *GetCommand) Execute(ctx context.Context, args []string) error {
	if err := c.ValidateArgs(args, 1, c.Usage()); err != nil {
		return err
	}

	recordID := args[0]

	record, err := c.Client.GetData(ctx, recordID)
	if err != nil {
		return fmt.Errorf("failed to get record: %w", err)
	}

	// Расшифровываем данные
	decryptedData, err := crypto.DecryptData([]byte(record.Data), c.masterPassword)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	fmt.Printf("ID: %s\n", record.ID)
	fmt.Printf("Type: %s\n", record.Type)
	fmt.Printf("Name: %s\n", record.Name)
	fmt.Printf("Metadata: %s\n", record.Metadata)
	fmt.Printf("Data: %s\n", string(decryptedData))
	fmt.Printf("Version: %d\n", record.Version)

	return nil
}
