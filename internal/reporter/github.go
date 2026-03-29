package reporter

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

const (
	// defaultGHTimeout is the maximum duration for gh API commands.
	defaultGHTimeout = 30 * time.Second
)

const commentMarker = "<!-- grpc-contract-guardian -->"

// CommandRunner abstracts external command execution for testability.
type CommandRunner interface {
	// Run executes a command and returns its combined output.
	Run(name string, args ...string) ([]byte, error)
}

// ExecCommandRunner is the default CommandRunner that delegates to os/exec.
type ExecCommandRunner struct{}

// Run executes a command using os/exec with a timeout.
func (r *ExecCommandRunner) Run(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGHTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204 -- arguments from trusted CLI flags
	return cmd.CombinedOutput()
}

// defaultRunner is the package-level runner, replaceable for testing.
var defaultRunner CommandRunner = &ExecCommandRunner{}

// SetCommandRunner replaces the default command runner (useful for testing).
func SetCommandRunner(r CommandRunner) {
	defaultRunner = r
}

// PostToGitHubPR posts the impact report as a PR comment.
// If a previous guardian comment exists, it updates it instead of creating a new one.
// Set dryRun=true to return the comment body without posting.
func PostToGitHubPR(report *ImpactReport, owner, repo string, prNumber int, dryRun bool) (string, error) {
	var body bytes.Buffer

	if _, err := fmt.Fprintln(&body, commentMarker); err != nil {
		return "", err
	}

	if err := WriteImpactGitHub(&body, report); err != nil {
		return "", fmt.Errorf("generating report: %w", err)
	}

	commentBody := body.String()

	if dryRun {
		return commentBody, nil
	}

	// Check for existing guardian comment
	existingID, err := findExistingComment(owner, repo, prNumber)
	if err != nil {
		// If we fail to list comments, fall through to create a new one.
		// This avoids duplicate comments in rare cases, but is more robust
		// than failing the entire operation.
		log.Printf("warning: failed to check for existing comment: %v", err)
	}

	if existingID != "" {
		return commentBody, updateComment(owner, repo, existingID, commentBody)
	}

	return commentBody, createComment(owner, repo, prNumber, commentBody)
}

func findExistingComment(owner, repo string, prNumber int) (string, error) {
	out, err := defaultRunner.Run("gh", "api",
		fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, repo, prNumber),
		"--jq", fmt.Sprintf(`.[] | select(.body | contains(%q)) | .id`, commentMarker),
	)
	if err != nil {
		return "", fmt.Errorf("listing PR comments: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	return strings.TrimSpace(string(out)), nil
}

func createComment(owner, repo string, prNumber int, body string) error {
	out, err := defaultRunner.Run("gh", "api",
		fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, repo, prNumber),
		"-f", fmt.Sprintf("body=%s", body),
	)
	if err != nil {
		return fmt.Errorf("creating comment: %w\n%s", err, out)
	}

	return nil
}

func updateComment(owner, repo, commentID, body string) error {
	out, err := defaultRunner.Run("gh", "api",
		fmt.Sprintf("repos/%s/%s/issues/comments/%s", owner, repo, commentID),
		"-X", "PATCH",
		"-f", fmt.Sprintf("body=%s", body),
	)
	if err != nil {
		return fmt.Errorf("updating comment: %w\n%s", err, out)
	}

	return nil
}
