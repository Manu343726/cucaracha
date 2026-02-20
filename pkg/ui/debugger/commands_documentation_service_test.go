package debugger

import (
	"testing"
)

// TestGetDocumentationService tests that the documentation service can be created and used
func TestGetDocumentationService(t *testing.T) {
	service, err := GetDocumentationService()
	if err != nil {
		t.Fatalf("Failed to create documentation service: %v", err)
	}

	// Test getting a command
	cmd, err := service.GetCommand("step")
	if err != nil {
		t.Fatalf("Failed to get 'step' command: %v", err)
	}

	if cmd.Name != "step" {
		t.Errorf("Expected command name 'step', got '%s'", cmd.Name)
	}

	if cmd.Description == "" {
		t.Error("Expected command description, got empty string")
	}

	t.Logf("Successfully retrieved command: %s - %s", cmd.Name, cmd.Description)
}

// TestGetAllCommands tests that all commands can be retrieved
func TestGetAllCommands(t *testing.T) {
	service, err := GetDocumentationService()
	if err != nil {
		t.Fatalf("Failed to create documentation service: %v", err)
	}

	commands := service.GetAllCommands()
	if len(commands) == 0 {
		t.Fatal("Expected at least one command, got none")
	}

	t.Logf("Successfully retrieved %d commands", len(commands))

	// Print some example commands for verification
	for i := 0; i < len(commands) && i < 3; i++ {
		t.Logf("  - %s: %s", commands[i].Name, commands[i].Description)
	}
}

// TestGetEnums tests that enums can be retrieved
func TestGetEnums(t *testing.T) {
	service, err := GetDocumentationService()
	if err != nil {
		t.Fatalf("Failed to create documentation service: %v", err)
	}

	enums := service.GetAllEnums()
	if len(enums) == 0 {
		t.Fatal("Expected at least one enum, got none")
	}

	t.Logf("Successfully retrieved %d enums", len(enums))
}

// TestGetCommandUsage tests that command usage documentation can be retrieved
func TestGetCommandUsage(t *testing.T) {
	service, err := GetDocumentationService()
	if err != nil {
		t.Fatalf("Failed to create documentation service: %v", err)
	}

	usage, err := service.GetCommandUsage("continue")
	if err != nil {
		t.Fatalf("Failed to get usage for 'continue' command: %v", err)
	}

	if usage == "" {
		t.Error("Expected usage string, got empty")
	}

	t.Logf("Usage documentation successfully retrieved:\n%s", usage)
}

// TestGetFormattedDocumentation tests that full documentation can be formatted
func TestGetFormattedDocumentation(t *testing.T) {
	service, err := GetDocumentationService()
	if err != nil {
		t.Fatalf("Failed to create documentation service: %v", err)
	}

	docs := service.GetFormattedDocumentation()
	if docs == "" {
		t.Error("Expected documentation string, got empty")
	}

	if len(docs) < 100 {
		t.Errorf("Expected significant documentation, got %d characters", len(docs))
	}

	t.Logf("Generated %d characters of formatted documentation", len(docs))
}
