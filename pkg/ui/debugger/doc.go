package debugger

// Code Generation
//
// The debugger package uses code generation to maintain command infrastructure derived from
// the DebuggerCommands interface. Run the following command to regenerate:
//   go generate ./pkg/ui/debugger
//
// Generated Files (all start with zz_ prefix):
//   - zz_commands_enum_generated.go: DebuggerCommandId enum and String()/FromString() methods
//   - zz_commands_structs_generated.go: DebuggerCommand and DebuggerCommandResult structs
//   - zz_execute_generated.go: Execute() method that dispatches commands to interface methods
//   - zz_commands_documentation_schema.go: CommandsDocsSchema with runtime documentation
//
// The generation is driven by the go:generate directive in command.go and uses the generator
// code in generator.go. For detailed information, see GENERATOR.md.
