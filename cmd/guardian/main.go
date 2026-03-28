package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "guardian",
		Short:   "gRPC/Proto backward compatibility checker + impact visualizer",
		Long:    "guardian analyzes .proto files for backward compatibility and visualizes the impact of breaking changes across service dependencies.",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newCheckCmd())
	root.AddCommand(newGraphCmd())
	root.AddCommand(newVersionCmd())

	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print guardian version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("guardian %s\n", version)
		},
	}
}
