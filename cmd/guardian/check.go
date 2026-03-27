package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/akaitigo/grpc-contract-guardian/internal/analyzer"
	"github.com/akaitigo/grpc-contract-guardian/internal/buf"
	"github.com/akaitigo/grpc-contract-guardian/internal/graph"
	"github.com/akaitigo/grpc-contract-guardian/internal/reporter"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	var (
		against   string
		format    string
		prNumber  int
		repo      string
		protoRoot string
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run breaking change detection and show impact",
		Long:  "Runs buf breaking against the specified ref, analyzes the dependency graph, and reports the impact of breaking changes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Run buf breaking
			bufOutput, err := runBufBreaking(against, protoRoot)
			if err != nil {
				return fmt.Errorf("buf breaking: %w", err)
			}

			// 2. Parse breaking changes
			breakingReport, err := buf.ParseOutput(bufOutput)
			if err != nil {
				return fmt.Errorf("parsing buf output: %w", err)
			}

			// 3. Analyze proto files for dependency graph
			protoFiles, err := findProtoFiles(protoRoot)
			if err != nil {
				return fmt.Errorf("finding proto files: %w", err)
			}

			parsed, err := analyzer.AnalyzeAll(protoFiles)
			if err != nil {
				return fmt.Errorf("analyzing proto files: %w", err)
			}

			g := graph.BuildFromProtoFiles(parsed)

			// 4. Generate impact report
			impactReport := reporter.AnalyzeImpact(breakingReport, g)

			// 5. Output
			switch format {
			case "text":
				return reporter.WriteImpactText(os.Stdout, impactReport)
			case "github":
				if prNumber == 0 {
					return fmt.Errorf("--pr is required for github format")
				}
				parts := strings.SplitN(repo, "/", 2)
				if len(parts) != 2 {
					return fmt.Errorf("--repo must be owner/repo format")
				}
				body, err := reporter.PostToGitHubPR(impactReport, parts[0], parts[1], prNumber, dryRun)
				if err != nil {
					return err
				}
				if dryRun {
					fmt.Print(body)
				} else {
					fmt.Fprintf(os.Stderr, "Posted impact report to PR #%d\n", prNumber)
				}
				return nil
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}
		},
	}

	cmd.Flags().StringVar(&against, "against", "main", "Git ref to compare against")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, github")
	cmd.Flags().IntVar(&prNumber, "pr", 0, "GitHub PR number (required for github format)")
	cmd.Flags().StringVar(&repo, "repo", "", "GitHub repository (owner/repo)")
	cmd.Flags().StringVar(&protoRoot, "proto-root", ".", "Root directory for .proto files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print output without posting to GitHub")

	return cmd
}

func runBufBreaking(against, protoRoot string) (string, error) {
	cmd := exec.Command("buf", "breaking", protoRoot, "--against", fmt.Sprintf(".git#branch=%s", against))
	out, err := cmd.CombinedOutput()
	if err != nil {
		// buf breaking returns exit code 1 when breaking changes are found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return string(out), nil
		}
		// buf not installed or other error
		if strings.Contains(string(out), "not found") || strings.Contains(err.Error(), "not found") {
			return "", fmt.Errorf("buf is not installed. Install from https://buf.build/docs/installation")
		}
		return string(out), nil
	}
	return string(out), nil
}

func findProtoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".proto") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
