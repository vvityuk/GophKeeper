package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/victor/gophkeeper/internal/common/models"
)

// UpdateCommand реализует команду обновления записи.
type UpdateCommand struct {
	*BaseCommand
}

// NewUpdateCommand создает новую команду обновления.
func NewUpdateCommand(base *BaseCommand) *UpdateCommand {
	return &UpdateCommand{BaseCommand: base}
}

// Name возвращает имя команды.
func (c *UpdateCommand) Name() string {
	return "update"
}

// Usage возвращает строку использования.
func (c *UpdateCommand) Usage() string {
	return "update <id> <data>"
}

// Description возвращает описание команды.
func (c *UpdateCommand) Description() string {
	return "Update data record"
}

// Execute выполняет команду обновления записи.
func (c *UpdateCommand) Execute(ctx context.Context, args []string) error {
	if err := c.ValidateArgs(args, 2, c.Usage()); err != nil {
		return err
	}

	recordID := args[0]
	data := args[1]

	// Получаем текущую запись для версии
	record, err := c.Client.GetData(ctx, recordID)
	if err != nil {
		return fmt.Errorf("failed to get record: %w", err)
	}

	req := &models.UpdateDataRequest{
		Data:    data,
		Version: record.Version,
	}

	updated, err := c.Client.UpdateData(ctx, recordID, req)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	// Сохраняем локально
	if err := c.LocalStorage.SaveDataRecord(ctx, updated); err != nil {
		log.Printf("warning: failed to save locally: %v", err)
	}

	fmt.Printf("Record updated: %s\n", updated.ID)
	return nil
}
