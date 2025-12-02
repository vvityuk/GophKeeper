package commands

import (
	"fmt"
)

// Registry управляет регистрацией и получением команд.
type Registry struct {
	commands map[string]Command
}

// NewRegistry создает новый реестр команд.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register регистрирует команду в реестре.
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

// Get возвращает команду по имени.
func (r *Registry) Get(name string) (Command, error) {
	cmd, ok := r.commands[name]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", name)
	}
	return cmd, nil
}

// GetAll возвращает все зарегистрированные команды.
func (r *Registry) GetAll() []Command {
	commands := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// GetUsage возвращает строку использования для всех команд.
func (r *Registry) GetUsage() string {
	usage := "GophKeeper CLI Client\n\nCommands:\n"
	for _, cmd := range r.GetAll() {
		usage += fmt.Sprintf("  %s - %s\n", cmd.Usage(), cmd.Description())
	}
	usage += "\nData types: CREDENTIALS, TEXT, BINARY, CARD"
	return usage
}

