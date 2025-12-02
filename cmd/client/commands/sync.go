package commands

import (
	"context"
	"fmt"
	"log"
)

// SyncCommand реализует команду синхронизации с сервером.
type SyncCommand struct {
	*BaseCommand
}

// NewSyncCommand создает новую команду синхронизации.
func NewSyncCommand(base *BaseCommand) *SyncCommand {
	return &SyncCommand{BaseCommand: base}
}

// Name возвращает имя команды.
func (c *SyncCommand) Name() string {
	return "sync"
}

// Usage возвращает строку использования.
func (c *SyncCommand) Usage() string {
	return "sync"
}

// Description возвращает описание команды.
func (c *SyncCommand) Description() string {
	return "Sync with server"
}

// Execute выполняет команду синхронизации.
func (c *SyncCommand) Execute(ctx context.Context, args []string) error {
	// Получаем данные с сервера
	serverRecords, err := c.Client.GetAllData(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	// Сохраняем все записи локально
	for _, record := range serverRecords {
		if err := c.LocalStorage.SaveDataRecord(ctx, record); err != nil {
			log.Printf("warning: failed to save record %s: %v", record.ID, err)
		}
	}

	fmt.Printf("Synced %d records\n", len(serverRecords))
	return nil
}

