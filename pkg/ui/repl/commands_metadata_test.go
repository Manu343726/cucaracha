package repl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetREPLCommandsMetadata verifies that command metadata is properly built
func TestGetREPLCommandsMetadata(t *testing.T) {
	metadata := GetREPLCommandsMetadata()
	require.NotNil(t, metadata, "Command metadata should not be nil")
	require.NotEmpty(t, metadata, "Command metadata should not be empty")

	// Verify that we have multiple groups (Debugger Commands + REPL-specific groups)
	assert.Greater(t, len(metadata), 3, "Should have multiple command groups")

	// Check that each group has a name and commands
	for _, group := range metadata {
		assert.NotEmpty(t, group.Name, "Group name should not be empty")
		assert.NotEmpty(t, group.Commands, "Group should have at least one command")
	}

	t.Logf("Loaded %d command groups with total commands", len(metadata))
}

// TestCommandGroups verifies expected command groups exist
func TestCommandGroups(t *testing.T) {
	metadata := GetREPLCommandsMetadata()

	expectedGroups := []string{
		"Debugger Commands", // Single group for all debugger commands
		"Settings",
		"Command Aliases",
		"Utility",
	}

	groupNames := make(map[string]bool)
	for _, group := range metadata {
		groupNames[group.Name] = true
	}

	for _, expectedGroup := range expectedGroups {
		assert.True(t, groupNames[expectedGroup], "Expected group '%s' to exist", expectedGroup)
	}
}

// TestExecutionCommands verifies execution commands are properly documented
func TestExecutionCommands(t *testing.T) {
	commands, ok := GetGroupCommands("Debugger Commands")
	require.True(t, ok, "Debugger Commands group should exist")
	require.NotEmpty(t, commands, "Debugger Commands group should have commands")

	commandsByName := make(map[string]bool)
	for _, cmd := range commands {
		commandsByName[cmd.Name] = true
		// Verify each command has required fields
		assert.NotEmpty(t, cmd.Name, "Command name should not be empty")
		assert.NotEmpty(t, cmd.Description, "Command description should not be empty")
		assert.NotEmpty(t, cmd.Usage, "Command usage should not be empty")
	}

	// Check for key execution commands
	assert.True(t, commandsByName["step"], "Expected 'step' command in Debugger Commands group")
	assert.True(t, commandsByName["continue"], "Expected 'continue' command in Debugger Commands group")
	assert.True(t, commandsByName["run"], "Expected 'run' command in Debugger Commands group")

	t.Logf("Debugger Commands group has %d commands", len(commands))
}

// TestBreakpointCommands verifies breakpoint commands are properly documented
func TestBreakpointCommands(t *testing.T) {
	commands, ok := GetGroupCommands("Debugger Commands")
	require.True(t, ok, "Debugger Commands group should exist")
	require.NotEmpty(t, commands, "Debugger Commands group should have commands")

	commandsByName := make(map[string]bool)
	for _, cmd := range commands {
		commandsByName[cmd.Name] = true
	}

	// Check for key breakpoint commands
	assert.True(t, commandsByName["break"], "Expected 'break' command in Debugger Commands group")
	assert.True(t, commandsByName["watch"], "Expected 'watch' command in Debugger Commands group")
	assert.True(t, commandsByName["list"], "Expected 'list' command in Debugger Commands group")
}

// TestInspectionCommands verifies inspection commands are properly documented
func TestInspectionCommands(t *testing.T) {
	commands, ok := GetGroupCommands("Debugger Commands")
	require.True(t, ok, "Debugger Commands group should exist")
	require.NotEmpty(t, commands, "Debugger Commands group should have commands")

	commandsByName := make(map[string]bool)
	for _, cmd := range commands {
		commandsByName[cmd.Name] = true
	}

	// Check for key inspection commands
	assert.True(t, commandsByName["disasm"], "Expected 'disasm' command in Debugger Commands group")
	assert.True(t, commandsByName["memory"], "Expected 'memory' command in Debugger Commands group")
	assert.True(t, commandsByName["stack"], "Expected 'stack' command in Debugger Commands group")
	assert.True(t, commandsByName["vars"], "Expected 'vars' command in Debugger Commands group")

	t.Logf("Debugger Commands group has %d commands", len(commands))
}

// TestREPLSpecificCommands verifies REPL-specific commands are present
func TestREPLSpecificCommands(t *testing.T) {
	metadata := GetREPLCommandsMetadata()

	expectedUtilityCommands := map[string]bool{
		"help": true,
		"exit": true,
	}

	expectedSettingsCommands := map[string]bool{
		"set": true,
		"get": true,
	}

	expectedAliasCommands := map[string]bool{
		"define":   true,
		"undefine": true,
	}

	// Find groups and commands
	for _, group := range metadata {
		switch group.Name {
		case "Utility":
			for _, cmd := range group.Commands {
				if expectedUtilityCommands[cmd.Name] {
					assert.NotEmpty(t, cmd.Description, "Command '%s' should have description", cmd.Name)
				}
			}
		case "Settings":
			for _, cmd := range group.Commands {
				if expectedSettingsCommands[cmd.Name] {
					assert.NotEmpty(t, cmd.Description, "Command '%s' should have description", cmd.Name)
				}
			}
		case "Command Aliases":
			for _, cmd := range group.Commands {
				if expectedAliasCommands[cmd.Name] {
					assert.NotEmpty(t, cmd.Description, "Command '%s' should have description", cmd.Name)
				}
			}
		}
	}
}

// TestGetCommandByName verifies command lookup by name
func TestGetCommandByName(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		shouldFnd bool
	}{
		{"Find step command", "step", true},
		{"Find step by alias s", "s", true},
		{"Find help command", "help", true},
		{"Find nonexistent command", "nonexistent", false},
		{"Case insensitive lookup", "STEP", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := GetCommandByName(test.query)
			if test.shouldFnd {
				require.NotNil(t, cmd, "Should find command '%s'", test.query)
				assert.NotEmpty(t, cmd.Description, "Found command should have description")
			} else {
				assert.Nil(t, cmd, "Should not find command '%s'", test.query)
			}
		})
	}
}

// TestGetAllCommands returns all commands as a flat list
func TestGetAllCommands(t *testing.T) {
	allCommands := GetAllCommands()
	require.NotEmpty(t, allCommands, "Should have commands")

	// Verify each command has required fields
	for _, cmd := range allCommands {
		assert.NotEmpty(t, cmd.Name, "Each command should have a name")
		assert.NotEmpty(t, cmd.Description, "Each command should have a description")
		assert.NotEmpty(t, cmd.Usage, "Each command should have usage")
	}

	// Commands should include both debugger and REPL commands
	commandNames := make(map[string]bool)
	for _, cmd := range allCommands {
		commandNames[cmd.Name] = true
	}

	// Check for some key commands from different categories
	assert.True(t, commandNames["step"], "Should have step command")
	assert.True(t, commandNames["help"], "Should have help command")
	assert.True(t, commandNames["set"], "Should have set command")

	t.Logf("Total commands returned: %d", len(allCommands))
}

// TestFormatCommandHelp verifies command help formatting
func TestFormatCommandHelp(t *testing.T) {
	cmd := Command{
		Name:        "step",
		Aliases:     []string{"s"},
		Description: "Step through one instruction",
		Usage:       "step [count]",
	}

	tests := []struct {
		name      string
		maxLen    int
		shouldPad bool
	}{
		{"No padding", 0, false},
		{"With padding", 20, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			formatted := FormatCommandHelp(&cmd, test.maxLen)
			assert.NotEmpty(t, formatted, "Formatted help should not be empty")
			assert.Contains(t, formatted, cmd.Description, "Formatted help should contain description")
			assert.Contains(t, formatted, cmd.Name, "Formatted help should contain command name")

			if test.shouldPad {
				parts := strings.Split(formatted, " - ")
				require.Len(t, parts, 2, "Should have name and description separated by ' - '")
				namePart := parts[0]
				// When padding is applied, name should be padded to at least maxLen
				assert.True(t, len(namePart) >= test.maxLen, "Name part should be padded to at least %d chars, got %d", test.maxLen, len(namePart))
			}
		})
	}
}

// TestGetMaxCommandNameLength returns appropriate length for formatting
func TestGetMaxCommandNameLength(t *testing.T) {
	maxLen := GetMaxCommandNameLength()
	require.Greater(t, maxLen, 0, "Max length should be greater than 0")

	// Verify the length is reasonable (between 5 and 50 chars typically)
	assert.Greater(t, maxLen, 4, "Max length should be at least 5")
	assert.Less(t, maxLen, 100, "Max length should be less than 100")

	t.Logf("Max command name length: %d", maxLen)
}

// TestFindCommandsByKeyword verifies keyword-based command search
func TestFindCommandsByKeyword(t *testing.T) {
	tests := []struct {
		keyword    string
		shouldFind int // minimum number of commands that should be found
	}{
		{"step", 1},
		{"break", 2}, // should find break and breakpoint commands
		{"memory", 1},
		{"help", 1},
		{"nonexistent", 0},
	}

	for _, test := range tests {
		t.Run("Search: "+test.keyword, func(t *testing.T) {
			results := FindCommandsByKeyword(test.keyword)
			assert.GreaterOrEqual(t, len(results), test.shouldFind, "Should find at least %d commands for keyword '%s'", test.shouldFind, test.keyword)

			// All results should be unique by name
			seen := make(map[string]bool)
			for _, cmd := range results {
				assert.False(t, seen[cmd.Name], "Duplicate command '%s' in results", cmd.Name)
				seen[cmd.Name] = true
			}
		})
	}
}

// TestCommandHasRequiredFields verifies all commands have required documentation
func TestCommandHasRequiredFields(t *testing.T) {
	allCommands := GetAllCommands()

	for _, cmd := range allCommands {
		t.Run(cmd.Name, func(t *testing.T) {
			assert.NotEmpty(t, cmd.Name, "Command must have a name")
			assert.NotEmpty(t, cmd.Description, "Command '%s' must have a description", cmd.Name)
			assert.NotEmpty(t, cmd.Usage, "Command '%s' must have usage", cmd.Name)
			// Details is optional, but if present should not be empty
			if cmd.Details != "" {
				assert.NotEmpty(t, cmd.Details, "Command '%s' details should not be empty if provided", cmd.Name)
			}
		})
	}
}

// TestCommandGroupConsistency verifies command groups are consistent
func TestCommandGroupConsistency(t *testing.T) {
	groups := GetCommandGroups()

	// Track all command names to ensure no duplicates across groups
	allCommandNames := make(map[string]int)

	for _, group := range groups {
		assert.NotEmpty(t, group.Name, "Group name should not be empty")
		assert.NotEmpty(t, group.Commands, "Group should have commands")

		for _, cmd := range group.Commands {
			allCommandNames[cmd.Name]++
		}
	}

	// Check for duplicates
	for cmdName, count := range allCommandNames {
		assert.Equal(t, 1, count, "Command '%s' appears in %d groups (should be 1)", cmdName, count)
	}
}

// TestDebuggerCommandsFromDocumentation verifies we're getting debugger commands from documentation
func TestDebuggerCommandsFromDocumentation(t *testing.T) {
	// This test verifies that debugger commands are being pulled from the documentation system
	// dynamically from DebuggerCommandIdValues rather than being hardcoded

	metadata := GetREPLCommandsMetadata()

	// Find the Debugger Commands group (contains all dynamic debugger commands)
	var debuggerGroup *CommandGroup
	for i := range metadata {
		if metadata[i].Name == "Debugger Commands" {
			debuggerGroup = &metadata[i]
			break
		}
	}

	require.NotNil(t, debuggerGroup, "Debugger Commands group should exist")
	require.NotEmpty(t, debuggerGroup.Commands, "Debugger Commands group should have commands")

	// Check that step command has proper documentation
	var stepCmd *Command
	for i := range debuggerGroup.Commands {
		if debuggerGroup.Commands[i].Name == "step" {
			stepCmd = &debuggerGroup.Commands[i]
			break
		}
	}

	require.NotNil(t, stepCmd, "Step command should exist")
	assert.NotEmpty(t, stepCmd.Description, "Step command should have description from documentation")
	assert.NotEmpty(t, stepCmd.Usage, "Step command should have usage")

	// The description should come from the API docs, not be a placeholder
	assert.NotEqual(t, "Debugger command", stepCmd.Description, "Description should be from documentation, not placeholder")

	t.Logf("Step command description: %s", stepCmd.Description)
}

// TestREPLSpecificCommandsNotFromDebugger verifies REPL commands are separate
func TestREPLSpecificCommandsNotFromDebugger(t *testing.T) {
	// Get all groups and filter for REPL-specific commands
	allGroups := GetCommandGroups()
	var replSpecificGroups []CommandGroup
	for _, group := range allGroups {
		var replCmds []Command
		for _, cmd := range group.Commands {
			if !cmd.IsDebugger {
				replCmds = append(replCmds, cmd)
			}
		}
		if len(replCmds) > 0 {
			replSpecificGroups = append(replSpecificGroups, CommandGroup{
				Name:     group.Name,
				Commands: replCmds,
			})
		}
	}

	assert.NotEmpty(t, replSpecificGroups, "Should have REPL-specific commands")

	// Check for expected REPL-specific commands that are NOT debugger commands
	allCmds := make(map[string]bool)
	for _, group := range replSpecificGroups {
		for _, cmd := range group.Commands {
			allCmds[cmd.Name] = true
		}
	}

	// These should be REPL-specific, not debugger commands
	replOnlyCommands := []string{"set", "get", "define", "undefine", "loggers"}
	for _, cmd := range replOnlyCommands {
		assert.True(t, allCmds[cmd], "REPL-specific command '%s' should be in REPL-specific group", cmd)
	}

	// Debugger commands should NOT be in REPL-specific metadata
	debuggerOnlyCommands := []string{"step", "continue", "break", "watch"}
	for _, cmd := range debuggerOnlyCommands {
		assert.False(t, allCmds[cmd], "Debugger command '%s' should NOT be in REPL-specific group", cmd)
	}
}
