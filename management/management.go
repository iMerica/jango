package management

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

type Command struct {
	Name        string
	Description string
	Handler     func(args []string) error
}

var commands = make(map[string]*Command)

func RegisterCommand(cmd *Command) {
	commands[cmd.Name] = cmd
}

func GetCommand(name string) (*Command, bool) {
	cmd, ok := commands[name]
	return cmd, ok
}

func AllCommands() map[string]*Command {
	result := make(map[string]*Command, len(commands))
	for k, v := range commands {
		result[k] = v
	}
	return result
}

func DiscoverAppCommands(appsDir string) error {
	if _, err := os.Stat(appsDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return fmt.Errorf("management: cannot read apps directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		commandsDir := filepath.Join(appsDir, entry.Name(), "management", "commands")
		if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
			continue
		}

		commandFiles, err := os.ReadDir(commandsDir)
		if err != nil {
			continue
		}

		for _, f := range commandFiles {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".so") {
				continue
			}

			cmdName := strings.TrimSuffix(f.Name(), ".so")
			cmdPath := filepath.Join(commandsDir, f.Name())

			p, err := plugin.Open(cmdPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "management: warning: cannot load plugin %s: %v\n", cmdPath, err)
				continue
			}

			handlerSym, err := p.Lookup("Handle")
			if err != nil {
				continue
			}

			handler, ok := handlerSym.(func([]string) error)
			if !ok {
				continue
			}

			descSym, err := p.Lookup("Description")
			desc := ""
			if err == nil {
				if d, ok := descSym.(*string); ok {
					desc = *d
				}
			}

			RegisterCommand(&Command{
				Name:        cmdName,
				Description: desc,
				Handler:     handler,
			})
		}
	}

	return nil
}