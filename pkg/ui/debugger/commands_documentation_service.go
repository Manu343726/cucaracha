package debugger

import (
	"fmt"
	"sync"
)

// CommandDocumentation represents the parsed command documentation schema
type CommandDocumentation struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	Arguments     []CommandArgumentInfo `json:"arguments"`
	Result        string                `json:"result"`
	ResultFields  []CommandResultField  `json:"resultFields"`
	Aliases       []string              `json:"aliases,omitempty"`
	AvailableWhen map[string]bool       `json:"availableWhen,omitempty"`
}

// CommandArgumentInfo represents documentation for a command argument
type CommandArgumentInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Required     bool   `json:"required"`
	Description  string `json:"description"`
	Example      string `json:"example,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
}

// CommandResultField represents documentation for a result field
type CommandResultField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
	OmitEmpty   bool   `json:"omitEmpty,omitempty"`
}

// EnumDocumentation represents documentation for an enum type
type EnumDocumentation map[string]string

// CommandsDocumentationSchema is the parsed documentation schema
type CommandsDocumentationSchema struct {
	Version  string                                     `json:"version"`
	Commands map[DebuggerCommandId]CommandDocumentation `json:"commands"`
	Enums    map[string]EnumDocumentation               `json:"enums,omitempty"`
}

// DocumentationService provides access to command documentation
type DocumentationService struct {
	schema *CommandsDocumentationSchema
	mu     sync.RWMutex
	// Index for quick lookups
	commandsByName map[string]*CommandDocumentation
	commandsByID   map[string]*CommandDocumentation
}

var (
	// globalDocService is the global documentation service instance
	globalDocService *DocumentationService
	docServiceOnce   sync.Once
)

// GetDocumentationService returns the global documentation service instance
func GetDocumentationService() (*DocumentationService, error) {
	var err error
	docServiceOnce.Do(func() {
		globalDocService, err = NewDocumentationService()
	})
	return globalDocService, err
}

// NewDocumentationService creates a new documentation service from the embedded schema
func NewDocumentationService() (*DocumentationService, error) {
	service := &DocumentationService{
		schema:         CommandsDocsSchema,
		commandsByName: make(map[string]*CommandDocumentation),
		commandsByID:   make(map[string]*CommandDocumentation),
	}

	// Build indexes for quick lookups
	for _, cmd := range service.schema.Commands {
		cmdCopy := cmd // Create a copy to get a stable pointer
		service.commandsByName[cmd.Name] = &cmdCopy
		service.commandsByID[cmd.ID] = &cmdCopy
	}

	return service, nil
}

// GetCommand retrieves documentation for a command by name
func (ds *DocumentationService) GetCommand(name string) (*CommandDocumentation, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if cmd, exists := ds.commandsByName[name]; exists {
		return cmd, nil
	}

	return nil, fmt.Errorf("command not found: %s", name)
}

// GetCommandByID retrieves documentation for a command by ID
func (ds *DocumentationService) GetCommandByID(id string) (*CommandDocumentation, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if cmd, exists := ds.commandsByID[id]; exists {
		return cmd, nil
	}

	return nil, fmt.Errorf("command not found by ID: %s", id)
}

// GetAllCommands returns all command documentation
func (ds *DocumentationService) GetAllCommands() []CommandDocumentation {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	result := make([]CommandDocumentation, 0, len(ds.schema.Commands))
	for _, cmd := range ds.schema.Commands {
		result = append(result, cmd)
	}
	return result
}

// GetEnum retrieves documentation for an enum type
func (ds *DocumentationService) GetEnum(enumName string) (EnumDocumentation, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if enum, exists := ds.schema.Enums[enumName]; exists {
		return enum, nil
	}

	return nil, fmt.Errorf("enum not found: %s", enumName)
}

// GetAllEnums returns all enum documentation
func (ds *DocumentationService) GetAllEnums() map[string]EnumDocumentation {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	result := make(map[string]EnumDocumentation)
	for name, enum := range ds.schema.Enums {
		result[name] = enum
	}
	return result
}

// GetCommandUsage returns formatted usage documentation for a command
func (ds *DocumentationService) GetCommandUsage(name string) (string, error) {
	cmd, err := ds.GetCommand(name)
	if err != nil {
		return "", err
	}

	usage := fmt.Sprintf("Command: %s\n", cmd.Name)
	usage += fmt.Sprintf("Description: %s\n", cmd.Description)

	if len(cmd.Arguments) > 0 {
		usage += "\nArguments:\n"
		for _, arg := range cmd.Arguments {
			required := ""
			if arg.Required {
				required = " (required)"
			} else {
				required = " (optional)"
			}
			usage += fmt.Sprintf("  - %s (%s)%s: %s\n", arg.Name, arg.Type, required, arg.Description)
			if arg.Example != "" {
				usage += fmt.Sprintf("    Example: %s\n", arg.Example)
			}
		}
	}

	if cmd.Result != "" {
		usage += fmt.Sprintf("\nResult: %s\n", cmd.Result)
		if len(cmd.ResultFields) > 0 {
			usage += "Result Fields:\n"
			for _, field := range cmd.ResultFields {
				usage += fmt.Sprintf("  - %s (%s): %s\n", field.Name, field.Type, field.Description)
			}
		}
	}

	return usage, nil
}

// GetFormattedDocumentation returns formatted documentation for all commands
func (ds *DocumentationService) GetFormattedDocumentation() string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	doc := "# Debugger Commands Documentation\n\n"

	for _, cmd := range ds.schema.Commands {
		doc += fmt.Sprintf("## %s\n", cmd.Name)
		doc += fmt.Sprintf("%s\n\n", cmd.Description)

		if len(cmd.Arguments) > 0 {
			doc += "### Arguments\n"
			for _, arg := range cmd.Arguments {
				required := "optional"
				if arg.Required {
					required = "required"
				}
				doc += fmt.Sprintf("- **%s** (`%s`, %s): %s\n", arg.Name, arg.Type, required, arg.Description)
			}
			doc += "\n"
		}

		if cmd.Result != "" {
			doc += fmt.Sprintf("### Result\n")
			doc += fmt.Sprintf("Type: `%s`\n\n", cmd.Result)

			if len(cmd.ResultFields) > 0 {
				doc += "#### Result Fields\n"
				for _, field := range cmd.ResultFields {
					doc += fmt.Sprintf("- **%s** (`%s`): %s\n", field.Name, field.Type, field.Description)
				}
				doc += "\n"
			}
		}

		doc += "\n"
	}

	return doc
}
