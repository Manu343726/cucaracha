package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/llvm"
)

func main() {
	outputFile := flag.String("outputFile", "Cucaracha.td", "LLVM Target description output file")

	flag.Parse()

	g, err := llvm.NewGenerator()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing llvm.Generator: %v\n", err)
		os.Exit(1)
	}

	if *outputFile == "stdout" {
		err = g.GenerateTo(os.Stdout)
	} else {
		err = g.Generate(*outputFile)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating target descriptor file: %v\n", err)
		os.Exit(2)
	}
}
