package repl

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/debugger"
)

// =============================================================================
// REPLSyntaxDescriptor - Interactive REPL syntax
// =============================================================================

// REPLSyntaxDescriptor formats commands for interactive REPL
// Example: step count:5
type REPLSyntaxDescriptor struct{}

// Ensure REPLSyntaxDescriptor implements CommandSyntaxDescriptor
var _ debugger.CommandSyntaxDescriptor = (*REPLSyntaxDescriptor)(nil)

func (r REPLSyntaxDescriptor) Name() string {
	return "repl"
}

func (r REPLSyntaxDescriptor) FormatCommandName(name string) string {
	return name
}

func (r REPLSyntaxDescriptor) FormatArgumentName(name string, isRequired bool) string {
	// REPL uses simple name:value syntax
	return name + ":"
}

func (r REPLSyntaxDescriptor) FormatArgumentValue(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

// GetREPLSyntax returns the REPL syntax descriptor instance
func GetREPLSyntax() debugger.CommandSyntaxDescriptor {
	return REPLSyntaxDescriptor{}
}
