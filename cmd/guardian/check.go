package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/akaitigo/grpc-contract-guardian/internal/analyzer"
	"github.com/akaitigo/grpc-contract-guardian/internal/buf"
	"github.com/akaitigo/grpc-contract-guardian/internal/graph"
	"github.com/akaitigo/grpc-contract-guardian/internal/reporter"
	"github.com/spf13/cobra"
)

const (
	// defaultExecTimeout is the maximum duration for external commands (buf, gh).
	defaultExecTimeout = 2 * time.Minute
)

const (
	formatText   = "text"
	formatGitHub = "github"
)

// branchNameRe validates git branch names (no spaces, control characters, or special sequences).
var branchNameRe = regexp.MustCompile(`^[a-zA-Z0-9._\-/]+$`)

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
			// Validate --against flag
			if !branchNameRe.MatchString(against) {
				return fmt.Errorf("invalid branch name for --against: %q", against)
			}

			// Validate --repo flag when github format is used
			if format == formatGitHub {
				if repo == "" {
					return fmt.Errorf("--repo is required for github format")
				}
				parts := strings.SplitN(repo, "/", 2)
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return fmt.Errorf("--repo must be in owner/repo format (e.g., myorg/myrepo), got %q", repo)
				}
			}

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

			if len(protoFiles) == 0 {
				return fmt.Errorf("no .proto files found in %q. Use --proto-root to specify the directory containing your proto files", protoRoot)
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
			case formatText:
				return reporter.WriteImpactText(cmd.OutOrStdout(), impactReport)
			case formatGitHub:
				if prNumber == 0 {
					return fmt.Errorf("--pr is required for github format")
				}
				parts := strings.SplitN(repo, "/", 2)
				body, err := reporter.PostToGitHubPR(impactReport, parts[0], parts[1], prNumber, dryRun)
				if err != nil {
					return err
				}
				if dryRun {
					fmt.Fprint(cmd.OutOrStdout(), body)
				} else {
					fmt.Fprintf(cmd.ErrOrStderr(), "Posted impact report to PR #%d\n", prNumber)
				}
				return nil
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}
		},
	}

	cmd.Flags().StringVar(&against, "against", "main", "Git ref to compare against")
	cmd.Flags().StringVar(&format, "format", formatText, "Output format: text, github")
	cmd.Flags().IntVar(&prNumber, "pr", 0, "GitHub PR number (required for github format)")
	cmd.Flags().StringVar(&repo, "repo", "", "GitHub repository (owner/repo)")
	cmd.Flags().StringVar(&protoRoot, "proto-root", ".", "Root directory for .proto files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print output without posting to GitHub")

	return cmd
}

func runBufBreaking(against, protoRoot string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultExecTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "buf", "breaking", protoRoot, "--against", fmt.Sprintf(".git#branch=%s", against)) // #nosec G204 -- arguments from trusted CLI flags
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			switch exitErr.ExitCode() {
			case 1:
				// buf breaking returns exit code 1 when breaking changes are found
				return string(out), nil
			default:
				// Exit code 2+ indicates a buf configuration or runtime error
				return "", fmt.Errorf("buf breaking failed (exit code %d): %s", exitErr.ExitCode(), strings.TrimSpace(string(out)))
			}
		}
		// buf not installed or other OS-level error
		if strings.Contains(string(out), "not found") || strings.Contains(err.Error(), "not found") {
			return "", fmt.Errorf("buf is not installed. Install from https://buf.build/docs/installation")
		}
		return "", fmt.Errorf("buf breaking: %w\n%s", err, out)
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
