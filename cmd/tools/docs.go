package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/utils"
	"github.com/spf13/cobra"
)

var module string
var supportedModules = map[string]func() string{
	"cpu.machine_code": func() string { return mc.Descriptor.DocString() },
}

var docsCmd = &cobra.Command{
	Use:   "docs module",
	Short: "Show cucaracha documentation",
	Long: `Dumps the documentation of the specified cucaracha module.
By default the tool dumps the documentation to stdout, but it can be redirected to a file using the --output flag.

Supported modules:
` + strings.Join(utils.Map(utils.Keys(supportedModules), func(module string) string { return "  " + module }), "\n"),
	Args:      cobra.MatchAll(cobra.OnlyValidArgs, cobra.MaximumNArgs(1), cobra.MinimumNArgs(1)),
	ValidArgs: utils.Keys(supportedModules),
	Run: func(cmd *cobra.Command, args []string) {
		module = args[0]
		outputFile, _ := cmd.Flags().GetString("output")
		if outputFile != "" {
			file, err := os.Create(outputFile)
			if err != nil {
				fmt.Println("Error creating file:", err)
				os.Exit(1)
			}
			defer file.Close()
			fmt.Fprintln(file, supportedModules[module]())
		} else {
			fmt.Println(supportedModules[module]())
		}
	},
}

func init() {
	ToolsCmd.AddCommand(docsCmd)
	docsCmd.Flags().StringP("output", "o", "", "Output file. If not specified, the documentation is dumped to stdout.")
}
