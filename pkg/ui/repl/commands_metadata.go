package repl

import (
	"sort"
	"strings"

	debuggerUI "github.com/Manu343726/cucaracha/pkg/ui/debugger"
)

// CommandGroup groups related commands for help display
type CommandGroup struct {
	Name     string
	Commands []Command
}

// Command represents a debugger or REPL command with its metadata
type Command struct {
	Name        string   // Primary command name
	Aliases     []string // Alternative names for this command
	Description string   // Short description
	Usage       string   // Usage pattern (e.g., "step [count]")
	Details     string   // Longer explanation (optional)
	IsDebugger  bool     // True if this is a debugger command, false if REPL-specific
}

// GetAllNames returns all names (primary name + aliases) for this command
func (c *Command) GetAllNames() []string {
	return append([]string{c.Name}, c.Aliases...)
}

// buildDebuggerCommandsFromMap dynamically discovers all debugger commands from DebuggerCommandIdValues
// and pulls their documentation from the documentation index without any hardcoding
func buildDebuggerCommandsFromMap() []Command {
	var commands []Command
	syntaxFormatter := REPLSyntax{}

	// Iterate through all available debugger commands
	for cmdID := range debuggerUI.DebuggerCommandIdValues {
		cmd := buildCommandFromDebuggerID(syntaxFormatter, cmdID)
		if cmd != nil {
			commands = append(commands, *cmd)
		}
	}

	// Sort by name for consistent ordering
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	return commands
}

// replSpecificCommands defines all REPL-specific commands (not part of the debugger API)
// These are maintained manually as they don't come from the debugger API
var replSpecificCommands = []CommandGroup{
	{
		Name: "Settings",
		Commands: []Command{
			{
				Name:        "set",
				Aliases:     []string{},
				Description: "Set a REPL setting",
				Usage:       "set [name] [value]",
				Details:     "Set a REPL setting value. Show all with descriptions if no arguments provided.",
				IsDebugger:  false,
			},
			{
				Name:        "get",
				Aliases:     []string{},
				Description: "Get a setting value",
				Usage:       "get [name]",
				Details:     "Get a setting value. Show all current values if no argument provided.",
				IsDebugger:  false,
			},
			{
				Name:        "save-settings",
				Aliases:     []string{},
				Description: "Save current settings to file",
				Usage:       "save-settings [file]",
				Details:     "Save current settings to YAML file. Uses loaded settings file if not specified.",
				IsDebugger:  false,
			},
		},
	},
	{
		Name: "Command Aliases",
		Commands: []Command{
			{
				Name:        "define",
				Aliases:     []string{},
				Description: "Define a new multi-command alias",
				Usage:       "define <name>",
				Details:     "Define a new alias that can run multiple commands. Supports nesting of aliases.",
				IsDebugger:  false,
			},
			{
				Name:        "undefine",
				Aliases:     []string{"unalias"},
				Description: "Remove an alias definition",
				Usage:       "undefine <name>",
				Details:     "Remove a previously defined alias.",
				IsDebugger:  false,
			},
			{
				Name:        "save-aliases",
				Aliases:     []string{},
				Description: "Save all aliases to file",
				Usage:       "save-aliases [file]",
				Details:     "Save all aliases to settings file. Loads from settings file if not specified.",
				IsDebugger:  false,
			},
		},
	},
	{
		Name: "Utility",
		Commands: []Command{
			{
				Name:        "loggers",
				Aliases:     []string{},
				Description: "List all registered loggers and their sinks",
				Usage:       "loggers",
				Details:     "Display all registered loggers and their associated output sinks.",
				IsDebugger:  false,
			},
			{
				Name:        "help",
				Aliases:     []string{"h"},
				Description: "Show this help message",
				Usage:       "help [command]",
				Details:     "Display help for all commands or a specific command.",
				IsDebugger:  false,
			},
			{
				Name:        "exit",
				Aliases:     []string{"quit", "q"},
				Description: "Exit the debugger",
				Usage:       "exit",
				Details:     "Exit the Cucaracha debugger REPL.",
				IsDebugger:  false,
			},
		},
	},
}

// buildDebuggerCommandGroups builds command groups from all available debugger commands
// discovered dynamically from DebuggerCommandIdValues, with documentation pulled from the index
func buildDebuggerCommandGroups() []CommandGroup {
	commands := buildDebuggerCommandsFromMap()

	// Group all debugger commands under a single "Debugger Commands" group
	// This ensures new commands are automatically included as they're added to the enum
	if len(commands) > 0 {
		return []CommandGroup{
			{
				Name:     "Debugger Commands",
				Commands: commands,
			},
		}
	}

	return []CommandGroup{}
}

// buildCommandFromDebuggerID builds a Command struct from a DebuggerCommandId using the documentation system.
// It pulls documentation from:
// 1. The method itself (DebuggerCommands.MethodName)
// 2. The args type if it exists
// 3. The result type if it exists
func buildCommandFromDebuggerID(syntaxFormatter REPLSyntax, cmdID debuggerUI.DebuggerCommandId) *Command {
	// Get the command name using the syntax formatter (converts to kebab-case)
	cmdNameFormatted := syntaxFormatter.FormatCommandName(cmdID)

	// Extract the method name from the command ID string
	// e.g., "DebuggerCommandStep" -> "Step"
	methodName := strings.TrimPrefix(cmdID.String(), "DebuggerCommand")

	var description, details string

	// Look up documentation from the debugger documentation system if available
	if debuggerUI.Documentation != nil && debuggerUI.Documentation.Entries != nil {
		// Look up method documentation
		packagePath := "github.com/Manu343726/cucaracha/pkg/ui/debugger"
		interfaceName := "DebuggerCommands"
		methodQualName := packagePath + "." + interfaceName + "." + methodName

		if entry, exists := debuggerUI.Documentation.Entries[methodQualName]; exists {
			description = entry.Summary
			if entry.Details != "" {
				details = entry.Details
			}
		}

		// Look up args type documentation
		argsTypeName := methodName + "Args"
		argsQualName := packagePath + "." + argsTypeName
		if entry, exists := debuggerUI.Documentation.Entries[argsQualName]; exists {
			argsDoc := entry.Details
			if argsDoc == "" {
				argsDoc = entry.Summary
			}
			if argsDoc != "" {
				if details != "" {
					details += " "
				}
				details += "Arguments: " + argsDoc
			}
		}

		// Look up result type documentation
		resultTypeName := methodName + "Result"
		resultQualName := packagePath + "." + resultTypeName
		if entry, exists := debuggerUI.Documentation.Entries[resultQualName]; exists {
			resultDoc := entry.Details
			if resultDoc == "" {
				resultDoc = entry.Summary
			}
			if resultDoc != "" {
				if details != "" {
					details += " "
				}
				details += "Result: " + resultDoc
			}
		}
	}

	// Fallback description if not found in documentation
	if description == "" {
		description = "Debugger command"
	}

	return &Command{
		Name:        cmdNameFormatted,
		Aliases:     []string{},
		Description: description,
		Usage:       cmdNameFormatted,
		Details:     details,
		IsDebugger:  true,
	}
}

// buildAllCommandGroups builds command groups combining debugger commands (from documentation system)
// and REPL-specific commands
func buildAllCommandGroups() []CommandGroup {
	groups := []CommandGroup{}

	// Add debugger command groups from the documentation system
	debuggerGroups := buildDebuggerCommandGroups()
	groups = append(groups, debuggerGroups...)

	// Add REPL-specific commands
	groups = append(groups, replSpecificCommands...)

	return groups
}

// GetREPLCommandsMetadata returns the complete command metadata for display
func GetREPLCommandsMetadata() []CommandGroup {
	return buildAllCommandGroups()
}

// GetDebuggerCommands returns only debugger commands
func GetDebuggerCommands() []Command {
	var commands []Command
	groups := GetREPLCommandsMetadata()
	for _, group := range groups {
		for _, cmd := range group.Commands {
			if cmd.IsDebugger {
				commands = append(commands, cmd)
			}
		}
	}
	return commands
}

// GetREPLSpecificCommands returns only REPL-specific commands
func GetREPLSpecificCommands() []Command {
	var commands []Command
	groups := GetREPLCommandsMetadata()
	for _, group := range groups {
		for _, cmd := range group.Commands {
			if !cmd.IsDebugger {
				commands = append(commands, cmd)
			}
		}
	}
	return commands
}

// GetCommandByName finds a command by its primary name or alias
func GetCommandByName(name string) *Command {
	lowerName := strings.ToLower(name)
	groups := GetREPLCommandsMetadata()
	for _, group := range groups {
		for i, cmd := range group.Commands {
			if strings.EqualFold(cmd.Name, lowerName) {
				return &group.Commands[i]
			}
			for _, alias := range cmd.Aliases {
				if strings.EqualFold(alias, lowerName) {
					return &group.Commands[i]
				}
			}
		}
	}
	return nil
}

// GetAllCommands returns all REPL commands as a flat list
func GetAllCommands() []Command {
	var commands []Command
	groups := GetREPLCommandsMetadata()
	for _, group := range groups {
		commands = append(commands, group.Commands...)
	}
	return commands
}

// GetCommandGroups returns all command groups
func GetCommandGroups() []CommandGroup {
	return GetREPLCommandsMetadata()
}

// FormatCommandHelp formats help text for display
func FormatCommandHelp(cmd *Command, maxNameLen int) string {
	// Format: name, aliases - description
	names := []string{cmd.Name}
	names = append(names, cmd.Aliases...)
	nameStr := strings.Join(names, ", ")

	// Pad the name string to align descriptions
	if maxNameLen > 0 {
		nameStr = strings.ToLower(nameStr)
		for len(nameStr) < maxNameLen {
			nameStr += " "
		}
		return nameStr + " - " + cmd.Description
	}

	return strings.ToLower(nameStr) + " - " + cmd.Description
}

// GetMaxCommandNameLength returns the maximum length of command names for formatting
func GetMaxCommandNameLength() int {
	maxLen := 0
	groups := GetREPLCommandsMetadata()
	for _, group := range groups {
		for _, cmd := range group.Commands {
			names := append([]string{cmd.Name}, cmd.Aliases...)
			nameStr := strings.Join(names, ", ")
			if len(nameStr) > maxLen {
				maxLen = len(nameStr)
			}
		}
	}
	return maxLen
}

// GetGroupCommands returns commands for a specific group by name
func GetGroupCommands(groupName string) ([]Command, bool) {
	groups := GetREPLCommandsMetadata()
	for _, group := range groups {
		if strings.EqualFold(group.Name, groupName) {
			return group.Commands, true
		}
	}
	return nil, false
}

// FindCommandsByKeyword searches for commands matching a keyword
func FindCommandsByKeyword(keyword string) []Command {
	var results []Command
	lowerKeyword := strings.ToLower(keyword)
	groups := GetREPLCommandsMetadata()

	for _, group := range groups {
		for _, cmd := range group.Commands {
			if strings.Contains(strings.ToLower(cmd.Name), lowerKeyword) ||
				strings.Contains(strings.ToLower(cmd.Description), lowerKeyword) ||
				strings.Contains(strings.ToLower(cmd.Details), lowerKeyword) {

				// Check if already in results (could be added from an alias match)
				found := false
				for _, existing := range results {
					if existing.Name == cmd.Name {
						found = true
						break
					}
				}
				if !found {
					results = append(results, cmd)
				}
			}

			// Also check aliases
			for _, alias := range cmd.Aliases {
				if strings.Contains(strings.ToLower(alias), lowerKeyword) {
					found := false
					for _, existing := range results {
						if existing.Name == cmd.Name {
							found = true
							break
						}
					}
					if !found {
						results = append(results, cmd)
					}
				}
			}
		}
	}

	// Sort results by name for consistency
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}
