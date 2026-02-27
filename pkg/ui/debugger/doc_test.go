package debugger

import (
	"bytes"
	"testing"
)

// mockFormatter implements CommandsFormatter for testing
type mockFormatter struct{}

func (m *mockFormatter) FormatCommandName(command DebuggerCommandId) string {
	return "test-command"
}

func (m *mockFormatter) FormatArgumentName(name string, isRequired bool) string {
	if isRequired {
		return "<" + name + ">"
	}
	return "[" + name + "]"
}

func (m *mockFormatter) FormatArgumentValue(value interface{}) string {
	return ""
}

// Helper function to create test documentation
func createTestDocumentation() Documentation {
	return Documentation{
		0: &CommandDocumentation{
			CommandID:   0,
			CommandName: "step",
			MethodName:  "Step",
			Summary:     "Executes a single step with {StepMode} mode",
			Description: "This command supports {CountMode} counting",
			Placeholders: []*DocumentationPlaceholder{
				{
					RawText:    "{StepMode}",
					Type:       "argument",
					EntityName: "StepMode",
					StartPos:   29,
					EndPos:     39,
				},
				{
					RawText:    "{CountMode}",
					Type:       "argument",
					EntityName: "CountMode",
					StartPos:   52,
					EndPos:     63,
				},
			},
			Arguments: map[string]*ArgumentDocumentation{
				"StepMode": {
					FieldName:  "StepMode",
					TypeName:   "StepMode",
					IsRequired: true,
					Summary:    "The stepping mode to use",
				},
				"CountMode": {
					FieldName:  "CountMode",
					TypeName:   "StepCountMode",
					IsRequired: false,
					Summary:    "How to count steps",
				},
			},
			Results: map[string]*ArgumentDocumentation{
				"State": {
					FieldName: "State",
					TypeName:  "ExecutionState",
					Summary:   "Current execution state",
				},
			},
			Examples: []CommandExample{
				{
					Description: "Step one instruction",
					Command:     "step",
					Output:      "stepped to next instruction",
				},
			},
		},
	}
}

func TestNewDocsRenderer(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}

	renderer := NewDocsRenderer(buf, formatter)

	if renderer == nil {
		t.Fatal("NewDocsRenderer returned nil")
	}

	if renderer.writer != buf {
		t.Error("writer not set correctly")
	}

	if renderer.formatter != formatter {
		t.Error("formatter not set correctly")
	}
}

func TestRenderName(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := createTestDocumentation()

	err := renderer.RenderName(0, docs)
	if err != nil {
		t.Fatalf("RenderName returned error: %v", err)
	}

	output := buf.String()
	if output != "test-command" {
		t.Errorf("expected 'test-command', got '%s'", output)
	}
}

func TestRenderName_NoDocumentation(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	err := renderer.RenderName(0, nil)
	if err == nil {
		t.Error("expected error for nil documentation")
	}
}

func TestRenderName_CommandNotFound(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := createTestDocumentation()

	err := renderer.RenderName(999, docs)
	if err == nil {
		t.Error("expected error for missing command")
	}
}

func TestRenderSummary(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := createTestDocumentation()

	err := renderer.RenderSummary(0, docs)
	if err != nil {
		t.Fatalf("RenderSummary returned error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("RenderSummary produced no output")
	}
}

func TestRenderDescription(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := createTestDocumentation()

	err := renderer.RenderDescription(0, docs)
	if err != nil {
		t.Fatalf("RenderDescription returned error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("RenderDescription produced no output")
	}
}

func TestRenderDescription_NoDescription(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := Documentation{
		0: &CommandDocumentation{
			CommandID:   0,
			CommandName: "test",
			MethodName:  "Test",
			Summary:     "Test command",
			Description: "", // No description
		},
	}

	err := renderer.RenderDescription(0, docs)
	if err != nil {
		t.Fatalf("RenderDescription returned error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output for no description, got %q", output)
	}
}

func TestRenderArguments(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := createTestDocumentation()

	err := renderer.RenderArguments(0, docs)
	if err != nil {
		t.Fatalf("RenderArguments returned error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("RenderArguments produced no output")
	}

	if !bytes.Contains([]byte(output), []byte("StepMode")) {
		t.Error("output does not contain StepMode")
	}
}

func TestRenderArguments_NoArguments(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := Documentation{
		0: &CommandDocumentation{
			CommandID:   0,
			CommandName: "test",
			MethodName:  "Test",
			Summary:     "Test command",
			Arguments:   make(map[string]*ArgumentDocumentation),
		},
	}

	err := renderer.RenderArguments(0, docs)
	if err != nil {
		t.Fatalf("RenderArguments returned error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output for no arguments, got %q", output)
	}
}

func TestRenderResults(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := createTestDocumentation()

	err := renderer.RenderResults(0, docs)
	if err != nil {
		t.Fatalf("RenderResults returned error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("RenderResults produced no output")
	}

	if !bytes.Contains([]byte(output), []byte("State")) {
		t.Error("output does not contain State")
	}
}

func TestRenderResults_NoResults(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := Documentation{
		0: &CommandDocumentation{
			CommandID:   0,
			CommandName: "test",
			MethodName:  "Test",
			Summary:     "Test command",
			Results:     make(map[string]*ArgumentDocumentation),
		},
	}

	err := renderer.RenderResults(0, docs)
	if err != nil {
		t.Fatalf("RenderResults returned error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output for no results, got %q", output)
	}
}

func TestRenderExamples(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := createTestDocumentation()

	err := renderer.RenderExamples(0, docs)
	if err != nil {
		t.Fatalf("RenderExamples returned error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("RenderExamples produced no output")
	}

	if !bytes.Contains([]byte(output), []byte("Step one instruction")) {
		t.Error("output does not contain example description")
	}
}

func TestRenderExamples_NoExamples(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	docs := Documentation{
		0: &CommandDocumentation{
			CommandID:   0,
			CommandName: "test",
			MethodName:  "Test",
			Summary:     "Test command",
			Examples:    []CommandExample{},
		},
	}

	err := renderer.RenderExamples(0, docs)
	if err != nil {
		t.Fatalf("RenderExamples returned error: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("expected empty output for no examples, got %q", output)
	}
}

func TestResolvePlaceholders(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	text := "Test with {ArgumentName} placeholder"
	placeholders := []*DocumentationPlaceholder{
		{
			RawText:    "{ArgumentName}",
			Type:       "argument",
			EntityName: "ArgumentName",
		},
	}

	result, err := renderer.resolvePlaceholders(text, placeholders)
	if err != nil {
		t.Fatalf("resolvePlaceholders returned error: %v", err)
	}

	if result == "" {
		t.Error("resolvePlaceholders produced empty result")
	}

	// Should contain the formatted argument name
	if !bytes.Contains([]byte(result), []byte("<")) {
		t.Error("result does not contain formatted argument")
	}
}

func TestResolvePlaceholders_NoPlaceholders(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &mockFormatter{}
	renderer := NewDocsRenderer(buf, formatter)

	text := "Test without placeholders"
	result, err := renderer.resolvePlaceholders(text, []*DocumentationPlaceholder{})
	if err != nil {
		t.Fatalf("resolvePlaceholders returned error: %v", err)
	}

	if result != text {
		t.Errorf("expected %q, got %q", text, result)
	}
}
