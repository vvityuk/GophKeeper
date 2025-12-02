package commands

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/victor/gophkeeper/internal/common/models"
)

// AddCommand реализует команду добавления новой записи.
type AddCommand struct {
	*BaseCommand
}

// NewAddCommand создает новую команду добавления.
func NewAddCommand(base *BaseCommand) *AddCommand {
	return &AddCommand{BaseCommand: base}
}

// Name возвращает имя команды.
func (c *AddCommand) Name() string {
	return "add"
}

// Usage возвращает строку использования.
func (c *AddCommand) Usage() string {
	return "add <type> <name> <data> [metadata]"
}

// Description возвращает описание команды.
func (c *AddCommand) Description() string {
	return "Add new data record"
}

// Execute выполняет команду добавления записи.
func (c *AddCommand) Execute(ctx context.Context, args []string) error {
	if err := c.ValidateArgs(args, 3, c.Usage()); err != nil {
		return err
	}

	dataType := strings.ToUpper(args[0])
	dataTypeEnum := models.DataType(dataType)
	if !dataTypeEnum.IsValid() {
		return fmt.Errorf("invalid data type: %s", dataType)
	}

	name := args[1]
	data := args[2]
	metadata := ""
	if len(args) >= 4 {
		metadata = args[3]
	}

	req := &models.CreateDataRequest{
		Type:     dataTypeEnum,
		Name:     name,
		Data:     data,
		Metadata: metadata,
	}

	record, err := c.Client.CreateData(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	// Сохраняем локально
	if err := c.LocalStorage.SaveDataRecord(ctx, record); err != nil {
		log.Printf("warning: failed to save locally: %v", err)
	}

	fmt.Printf("Record created: %s\n", record.ID)
	return nil
}
