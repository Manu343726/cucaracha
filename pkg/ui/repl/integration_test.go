package repl

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/debugger"
	"github.com/Manu343726/cucaracha/pkg/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/scripts/*.repl
var scriptsFS embed.FS

// ScriptExpectation represents an expected output string for a command
type ScriptExpectation struct {
	Command         string
	ExpectedOuput   string
	ExpectedSuccess bool
}

// TestFixture holds test setup and teardown
type IntegrationTestFixture struct {
	Debugger      ui.Debugger       // UI interface for REPL
	UnderDebugger debugger.Debugger // Underlying debugger for setup
	REPL          *REPL
	OutputBuf     *bytes.Buffer
}

// NewIntegrationTestFixture creates a new test fixture with a real debugger
func NewIntegrationTestFixture() *IntegrationTestFixture {
	underlying := debugger.NewDebugger()
	uiDebugger := debugger.NewDebuggerForUI(underlying)

	return &IntegrationTestFixture{
		Debugger:      uiDebugger,
		UnderDebugger: underlying,
	}
}

// RunScript executes a script with the fixture, filtering out comment lines
func (f *IntegrationTestFixture) RunScript(commands string) error {
	// Filter out comment lines (lines starting with #)
	var filteredLines []string
	for _, line := range strings.Split(commands, "\n") {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comment lines
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			filteredLines = append(filteredLines, line)
		}
	}

	filteredCommands := strings.Join(filteredLines, "\n")
	input := strings.NewReader(filteredCommands + "\nexit\n")
	f.OutputBuf = &bytes.Buffer{}
	f.REPL = NewREPLWithOutputFormat(f.Debugger, input, f.OutputBuf, MachineReadable)
	return f.REPL.Run()
}

// GetOutput returns the REPL output
func (f *IntegrationTestFixture) GetOutput() string {
	if f.OutputBuf == nil {
		return ""
	}
	return f.OutputBuf.String()
}

// ParseScriptExpectations parses a script file and extracts expected output markers
func ParseScriptExpectations(scriptContent string) []ScriptExpectation {
	var expectations []ScriptExpectation
	scanner := bufio.NewScanner(strings.NewReader(scriptContent))

	var currentExpectation ScriptExpectation
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			// Check if this is an expected output marker
			//
			// Expected output markers are in the format:
			// # Expected:
			// #   Result: success/error
			// #   Output: single line of exact expected output
			//
			// Or for multiline output:
			// # Expected:
			// #   Result: success/error
			// #   Output:
			// #     line 1 of expected output
			// #     line 2 of expected output
			//

			if strings.HasPrefix(line, "# Expected:") {
				// Start of a new expectation
				if currentExpectation.Command != "" {
					expectations = append(expectations, currentExpectation)
				}
				currentExpectation = ScriptExpectation{}
			} else if strings.HasPrefix(line, "#  Result:") {
				result := strings.TrimSpace(strings.TrimPrefix(line, "#  Result:"))
				currentExpectation.ExpectedSuccess = (result == "success")
			} else if strings.HasPrefix(line, "#  Output:") {
				output := strings.TrimSpace(strings.TrimPrefix(line, "#  Output:"))
				currentExpectation.ExpectedOuput = output
			} else if strings.HasPrefix(line, "#    ") {
				// Multiline expected output - append to existing expected output
				multilineOutput := strings.TrimSpace(strings.TrimPrefix(line, "#    "))
				if currentExpectation.ExpectedOuput != "" {
					currentExpectation.ExpectedOuput += "\n" + multilineOutput
				} else {
					currentExpectation.ExpectedOuput = multilineOutput
				}
			}

			continue
		}

		// This is a command
		currentExpectation.Command = line
		expectations = append(expectations, currentExpectation)
		currentExpectation = ScriptExpectation{}
	}

	return expectations
}

// TestIntegration runs comprehensive integration tests for REPL scripts
// Each script is self-contained with setup and test commands
func TestIntegration(t *testing.T) {
	// Glob all test scripts in the embedded filesystem
	entries, err := scriptsFS.ReadDir("testdata/scripts")
	require.NoError(t, err, "should read embedded scripts directory")

	var foundScripts []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".repl") {
			foundScripts = append(foundScripts, entry.Name())
		}
	}

	for _, scriptName := range foundScripts {
		t.Run(scriptName, func(t *testing.T) {
			// Load script file
			data, err := scriptsFS.ReadFile("testdata/scripts/" + scriptName)
			require.NoError(t, err, "should read script file: %s", scriptName)

			scriptContent := string(data)
			assert.NotEmpty(t, scriptContent, "script should not be empty")

			// Parse expected outputs from script
			expectations := ParseScriptExpectations(scriptContent)
			require.Greater(t, len(expectations), 0, "script should have expected output markers")

			// Verify each expectation is well-formed
			for i, exp := range expectations {
				require.NotEmpty(t, exp.Command, "expectation %d should have a command", i)
				require.NotEmpty(t, exp.ExpectedOuput, "expectation %d should have expected output", i)
			}

			// Setup and run the script
			fixture := NewIntegrationTestFixture()
			err = fixture.RunScript(scriptContent)
			require.NoError(t, err, "script should execute successfully")

			output := fixture.GetOutput()
			require.NotEmpty(t, output, "script should produce output")

			// Parse JSONL output
			commandOutputs := []CommandOutput{}
			scanner := bufio.NewScanner(strings.NewReader(output))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var cmdOutput CommandOutput
				err := json.Unmarshal([]byte(line), &cmdOutput)
				require.NoError(t, err, "should parse JSONL output line: %s", line)
				commandOutputs = append(commandOutputs, cmdOutput)
			}

			require.NoError(t, scanner.Err(), "should read all output lines")
			require.Greater(t, len(commandOutputs), 0, "should have at least one command output")

			t.Logf("Parsed %d command outputs from %s", len(commandOutputs), scriptName)

			// Build a map of commands to their outputs for easier lookup
			commandMap := make(map[string]CommandOutput)
			for _, cmdOut := range commandOutputs {
				commandMap[cmdOut.Command] = cmdOut
			}

			// Verify each expected output appears in the actual command output
			for _, exp := range expectations {
				cmdOut, found := commandMap[exp.Command]
				require.True(t, found, "command '%s' was not executed", exp.Command)

				assert.Equal(t, exp.ExpectedSuccess, cmdOut.Success, "%s:%v '%s' unexpected result", cmdOut.File, cmdOut.Line, exp.Command)
				assert.Equal(t, exp.ExpectedOuput, cmdOut.Output, "%s:%v '%s' unexpected output", cmdOut.File, cmdOut.Line, exp.Command)
			}
		})
	}
}
