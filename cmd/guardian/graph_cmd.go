package main

import (
	"fmt"
	"os"

	"github.com/akaitigo/grpc-contract-guardian/internal/analyzer"
	"github.com/akaitigo/grpc-contract-guardian/internal/graph"
	"github.com/spf13/cobra"
)

func newGraphCmd() *cobra.Command {
	var (
		output    string
		protoRoot string
	)

	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Output service dependency graph",
		Long:  "Analyzes .proto files and outputs the service dependency graph in DOT or text format.",
		RunE: func(cmd *cobra.Command, args []string) error {
			protoFiles, err := findProtoFiles(protoRoot)
			if err != nil {
				return fmt.Errorf("finding proto files: %w", err)
			}

			if len(protoFiles) == 0 {
				return fmt.Errorf("no .proto files found in %s", protoRoot)
			}

			parsed, err := analyzer.AnalyzeAll(protoFiles)
			if err != nil {
				return fmt.Errorf("analyzing proto files: %w", err)
			}

			g := graph.BuildFromProtoFiles(parsed)

			switch output {
			case formatText:
				return g.WriteText(os.Stdout)
			case "dot":
				return g.WriteDOT(os.Stdout)
			default:
				return fmt.Errorf("unsupported output format: %s (use text or dot)", output)
			}
		},
	}

	cmd.Flags().StringVar(&output, "output", formatText, "Output format: text, dot")
	cmd.Flags().StringVar(&protoRoot, "proto-root", ".", "Root directory for .proto files")

	return cmd
}
