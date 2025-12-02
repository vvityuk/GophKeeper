package commands

import (
	"context"
	"fmt"
)

// ListCommand реализует команду списка всех записей.
type ListCommand struct {
	*BaseCommand
}

// NewListCommand создает новую команду списка.
func NewListCommand(base *BaseCommand) *ListCommand {
	return &ListCommand{BaseCommand: base}
}

// Name возвращает имя команды.
func (c *ListCommand) Name() string {
	return "list"
}

// Usage возвращает строку использования.
func (c *ListCommand) Usage() string {
	return "list"
}

// Description возвращает описание команды.
func (c *ListCommand) Description() string {
	return "List all data records"
}

// Execute выполняет команду списка.
func (c *ListCommand) Execute(ctx context.Context, args []string) error {
	records, err := c.Client.GetAllData(ctx)
	if err != nil {
		return fmt.Errorf("failed to get data: %w", err)
	}

	if len(records) == 0 {
		fmt.Println("No records found")
		return nil
	}

	fmt.Println("Data records:")
	for _, record := range records {
		fmt.Printf("  ID: %s, Type: %s, Name: %s\n", record.ID, record.Type, record.Name)
	}
	return nil
}

