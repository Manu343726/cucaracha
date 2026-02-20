// +build ignore

package main

import (
"os"
"os/exec"
)

func main() {
	cmd := exec.Command("go", "run", "./generators/docs.go", "./zz_commands_documentation_schema.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
