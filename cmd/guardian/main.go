// Package main is the entry point for the grpc-contract-guardian CLI.
// guardian analyzes .proto files for backward compatibility and visualizes
// the impact of breaking changes across service dependencies.
package main

import (
	"fmt"
	"os"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("guardian %s\n", version)
		return
	}

	fmt.Fprintln(os.Stderr, "grpc-contract-guardian: proto backward compatibility checker + impact visualizer")
	fmt.Fprintln(os.Stderr, "Usage: guardian <command> [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  check   Run breaking change detection and show impact")
	fmt.Fprintln(os.Stderr, "  graph   Output service dependency graph")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Use \"guardian <command> --help\" for more information about a command.")
	os.Exit(1)
}
