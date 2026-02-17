package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Manu343726/cucaracha/pkg/debugger"
	"github.com/Manu343726/cucaracha/pkg/ui"
	"github.com/Manu343726/cucaracha/pkg/ui/repl"
	"github.com/Manu343726/cucaracha/pkg/ui/tui/tview"
	"github.com/spf13/cobra"
)

// TuiCmd represents the debug command
var TuiCmd = &cobra.Command{
	Use:   "debug [program]",
	Short: "Launch the interactive debugger",
	Long: `Launch the interactive debugger for the Cucaracha emulator.

The optional program argument is a path to a .c or .cpp file to debug.
Use --runtime to specify the runtime (default: interpreter).
Use --system to specify a system configuration file (uses embedded default if not provided).`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Create the underlying debugger and wrap it for UI
		underlying := debugger.NewDebugger()
		uiDebugger := debugger.NewDebuggerForUI(underlying)
		var loadArgs ui.LoadArgs

		// Extract program path if provided
		if len(args) > 0 {
			ext := filepath.Ext(args[0])
			if ext == ".c" || ext == ".cpp" {
				loadArgs.ProgramPath = &args[0]
			} else if ext == ".yaml" || ext == ".yml" {
				loadArgs.FullDescriptorPath = &args[0]
			} else {
				fmt.Fprintf(os.Stderr, "Error: unrecognized file extension '%s' for argument '%s'\n", ext, args[0])
				os.Exit(1)
			}
		}

		// Set system config path if provided
		if debugSystem != "" {
			loadArgs.SystemConfigPath = &debugSystem
		}

		// Parse runtime type
		runtimeType, err := ui.RuntimeTypeFromString(debugRuntime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid runtime type: %v\n", err)
			os.Exit(1)
		}
		loadArgs.Runtime = &runtimeType

		// Use REPL if --tui flag is not provided, otherwise use TUI
		if !useTUI {
			// Use REPL mode
			replInstance := repl.NewREPL(uiDebugger)
			replInstance.Run()
		} else {
			// Force true color support
			// These must be set BEFORE tcell/tview initializes
			os.Setenv("COLORTERM", "truecolor")
			// Also try setting TERM to a 256-color capable terminal
			if term := os.Getenv("TERM"); term == "" || term == "dumb" {
				os.Setenv("TERM", "xterm-256color")
			}

			// Create the TUI with the underlying debugger
			program := tview.NewDebuggerTUI(uiDebugger,
				tview.WithoutCatchPanics(),
				tview.WithLoadArgs(&loadArgs),
			)

			if err := program.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	TuiCmd.Flags().StringVar(&debugRuntime, "runtime", "interpreter", "Runtime to use (interpreter)")
	TuiCmd.Flags().StringVar(&debugSystem, "system", "", "Path to system configuration file (uses embedded default if not provided)")
	TuiCmd.Flags().BoolVar(&useTUI, "tui", false, "Use the TUI interface (default: use REPL)")
	RootCmd.AddCommand(TuiCmd)
}
