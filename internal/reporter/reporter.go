// Package reporter generates human-readable reports from breaking change
// analysis and dependency graph data. Supports text output for terminals
// and Markdown output for GitHub PR comments.
package reporter

import (
	"fmt"
	"io"

	"github.com/akaitigo/grpc-contract-guardian/internal/buf"
)

// Format specifies the output format for reports.
type Format string

const (
	// FormatText outputs plain text suitable for terminal display.
	FormatText Format = "text"
	// FormatGitHub outputs Markdown suitable for GitHub PR comments.
	FormatGitHub Format = "github"
)

// Report generates a formatted report from a breaking change report.
// The output is written to the provided writer in the specified format.
func Report(w io.Writer, report *buf.BreakingReport, format Format) error {
	if report == nil {
		return fmt.Errorf("report is nil")
	}

	switch format {
	case FormatText:
		return writeTextReport(w, report)
	case FormatGitHub:
		return writeGitHubReport(w, report)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func writeTextReport(w io.Writer, report *buf.BreakingReport) error {
	if report.TotalCount == 0 {
		_, err := fmt.Fprintln(w, "No breaking changes detected.")
		return err
	}

	if _, err := fmt.Fprintf(w, "Breaking Changes: %d\n", report.TotalCount); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "---"); err != nil {
		return err
	}

	for _, c := range report.Changes {
		if _, err := fmt.Fprintf(w, "[%s] %s:%d - %s\n", c.Category, c.File, c.Line, c.Description); err != nil {
			return err
		}
	}

	return nil
}

func writeGitHubReport(w io.Writer, report *buf.BreakingReport) error {
	if report.TotalCount == 0 {
		_, err := fmt.Fprintln(w, "## :white_check_mark: No Breaking Changes\n\nAll proto definitions are backward compatible.")
		return err
	}

	if _, err := fmt.Fprintf(w, "## :warning: Breaking Changes Detected (%d)\n\n", report.TotalCount); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "| Category | File | Description |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "|----------|------|-------------|"); err != nil {
		return err
	}

	for _, c := range report.Changes {
		if _, err := fmt.Fprintf(w, "| `%s` | `%s:%d` | %s |\n", c.Category, c.File, c.Line, c.Description); err != nil {
			return err
		}
	}

	return nil
}
