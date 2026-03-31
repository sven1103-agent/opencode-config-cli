package main

import (
	"fmt"
	"os"

	"github.com/sven1103-agent/opencode-config-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
