// Package main is the entry point for the azfloci CLI.
package main

import (
	"fmt"
	"os"

	"github.com/ankurCES/floc-zure/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
