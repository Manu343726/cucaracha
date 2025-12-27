package mc

import (
	"fmt"
	"os"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/llvm"
	"github.com/spf13/cobra"
)

var outputFile string
var clangPath string

// generateLlvmTablegenCmd represents the generateLlvmTablegen command
var generateLlvmTablegenCmd = &cobra.Command{
	Use:   "generateLlvmTablegen",
	Short: "Generate LLVM tablegen descriptor files for the Cucaracha architecture",
	Long: `Cucaracha includes a forked LLVM toolchain implementing the Cucaracha architecture
so that C/C++ code can be run in a cucaracha environment (Interpreter, hardware, etc). This command bootstraps
the tablegen files needed by the LLVM backend to implement code generation for the Cucaracha architecture.

See https://llvm.org/docs/CodeGenerator.html for more information about LLVM code generation.`,
	Run: func(cmd *cobra.Command, args []string) {
		g, err := llvm.NewGenerator()

		if err != nil {
			fmt.Fprintf(os.Stderr, "error initializing llvm.Generator: %v\n", err)
			os.Exit(1)
		}

		if len(outputFile) == 0 {
			err = g.GenerateTo(os.Stdout)
		} else {
			err = g.Generate(outputFile)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating target descriptor file: %v\n", err)
			os.Exit(2)
		}
	},
}

var clangVersionCmd = &cobra.Command{
	Use:   "clangVersion",
	Short: "Output the clang version used by the Cucaracha toolchain",
	Run: func(cmd *cobra.Command, args []string) {
		driver, err := llvm.NewClangDriver(clangPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding clang: %v\n", err)
			os.Exit(1)
		}

		version, err := driver.Version()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting clang version: %v\n", err)
			os.Exit(2)
		}

		fmt.Println(version)
	},
}

func init() {
	McCmd.AddCommand(generateLlvmTablegenCmd, clangVersionCmd)
	generateLlvmTablegenCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file. If omitted, the output will be written to stdout")
	clangVersionCmd.Flags().StringVarP(&clangPath, "clang-path", "p", "", "Full path to the clang executable")
}
