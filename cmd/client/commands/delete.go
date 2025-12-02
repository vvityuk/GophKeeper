package commands

import (
	"context"
	"fmt"
	"log"
)

// DeleteCommand реализует команду удаления записи.
type DeleteCommand struct {
	*BaseCommand
}

// NewDeleteCommand создает новую команду удаления.
func NewDeleteCommand(base *BaseCommand) *DeleteCommand {
	return &DeleteCommand{BaseCommand: base}
}

// Name возвращает имя команды.
func (c *DeleteCommand) Name() string {
	return "delete"
}

// Usage возвращает строку использования.
func (c *DeleteCommand) Usage() string {
	return "delete <id>"
}

// Description возвращает описание команды.
func (c *DeleteCommand) Description() string {
	return "Delete data record"
}

// Execute выполняет команду удаления записи.
func (c *DeleteCommand) Execute(ctx context.Context, args []string) error {
	if err := c.ValidateArgs(args, 1, c.Usage()); err != nil {
		return err
	}

	recordID := args[0]

	if err := c.Client.DeleteData(ctx, recordID); err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	// Удаляем локально
	if err := c.LocalStorage.DeleteDataRecord(ctx, recordID); err != nil {
		log.Printf("warning: failed to delete locally: %v", err)
	}

	fmt.Printf("Record deleted: %s\n", recordID)
	return nil
}

