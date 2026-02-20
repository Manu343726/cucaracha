package debugger

// Implements a channel-based asynchronous debugger UI
type AsyncDebuggerUI struct {
	debugger CommandBasedDebugger

	commands chan *asyncDebuggerCommand
}

type AsyncDebuggerCommandResultCallback func(result *DebuggerCommandResult, err error)

type asyncDebuggerCommand struct {
	cmd      *DebuggerCommand
	callback AsyncDebuggerCommandResultCallback
}

// Creates a new AsyncDebuggerUI
func NewAsyncDebuggerUI(debugger CommandBasedDebugger) *AsyncDebuggerUI {
	ui := &AsyncDebuggerUI{
		debugger: debugger,
		commands: make(chan *asyncDebuggerCommand, 10),
	}

	go ui.commandLoop()

	return ui
}

func (ui *AsyncDebuggerUI) SetEventCallback(callback DebuggerEventCallback) {
	ui.debugger.SetEventCallback(callback)
}

// Internal command processing loop
func (ui *AsyncDebuggerUI) commandLoop() {
	for asyncCmd := range ui.commands {
		result, err := ui.debugger.Execute(asyncCmd.cmd)
		asyncCmd.callback(result, err)
	}
}

// Sends a command to the debugger and returns the result asynchronously
func (ui *AsyncDebuggerUI) Execute(cmd *DebuggerCommand, callback AsyncDebuggerCommandResultCallback) {
	if cmd.Command == DebuggerCommandInterrupt {
		// Explicitly handle interrupt command with a direct call to avoid potential deadlocks
		// with the command loop stuck in a long-running command.
		result, err := ui.debugger.Execute(cmd)
		callback(result, err)
		return
	}

	asyncCmd := &asyncDebuggerCommand{
		cmd:      cmd,
		callback: callback,
	}
	ui.commands <- asyncCmd
}
