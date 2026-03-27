package reporter

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const commentMarker = "<!-- grpc-contract-guardian -->"

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
	existingID := findExistingComment(owner, repo, prNumber)

	if existingID != "" {
		return commentBody, updateComment(owner, repo, existingID, commentBody)
	}

	return commentBody, createComment(owner, repo, prNumber, commentBody)
}

func findExistingComment(owner, repo string, prNumber int) string {
	cmd := exec.Command("gh", "api", // #nosec G204 -- arguments from trusted CLI flags
		fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, repo, prNumber),
		"--jq", fmt.Sprintf(`.[] | select(.body | contains(%q)) | .id`, commentMarker),
	)

	out, err := cmd.Output()
	if err != nil {
		return "" // not found is OK
	}

	return strings.TrimSpace(string(out))
}

func createComment(owner, repo string, prNumber int, body string) error {
	cmd := exec.Command("gh", "api", // #nosec G204 -- arguments from trusted CLI flags
		fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, repo, prNumber),
		"-f", fmt.Sprintf("body=%s", body),
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating comment: %w\n%s", err, out)
	}

	return nil
}

func updateComment(owner, repo, commentID, body string) error {
	cmd := exec.Command("gh", "api", // #nosec G204 -- arguments from trusted CLI flags
		fmt.Sprintf("repos/%s/%s/issues/comments/%s", owner, repo, commentID),
		"-X", "PATCH",
		"-f", fmt.Sprintf("body=%s", body),
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("updating comment: %w\n%s", err, out)
	}

	return nil
}
