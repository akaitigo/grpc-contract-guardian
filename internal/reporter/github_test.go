package reporter_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akaitigo/grpc-contract-guardian/internal/buf"
	"github.com/akaitigo/grpc-contract-guardian/internal/graph"
	"github.com/akaitigo/grpc-contract-guardian/internal/reporter"
)

// mockRunner records command invocations and returns preconfigured responses.
type mockRunner struct {
	calls   [][]string
	outputs map[string]mockResult
}

type mockResult struct {
	output []byte
	err    error
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		outputs: make(map[string]mockResult),
	}
}

func (m *mockRunner) Run(name string, args ...string) ([]byte, error) {
	call := append([]string{name}, args...)
	m.calls = append(m.calls, call)

	key := strings.Join(call, " ")
	for pattern, result := range m.outputs {
		if strings.Contains(key, pattern) {
			return result.output, result.err
		}
	}

	return []byte(""), nil
}

func TestPostToGitHubPR_CreateComment(t *testing.T) {
	t.Parallel()

	mock := newMockRunner()
	// No existing comment found
	mock.outputs["issues/1/comments"] = mockResult{output: []byte(""), err: nil}

	reporter.SetCommandRunner(mock)
	t.Cleanup(func() { reporter.SetCommandRunner(&reporter.ExecCommandRunner{}) })

	breaking := &buf.BreakingReport{TotalCount: 0}
	g := graph.NewGraph()
	report := reporter.AnalyzeImpact(breaking, g)

	body, err := reporter.PostToGitHubPR(report, "testowner", "testrepo", 1, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(body, "grpc-contract-guardian") {
		t.Error("body missing guardian marker")
	}

	// Verify that gh api was called to create a comment
	found := false
	for _, call := range mock.calls {
		joined := strings.Join(call, " ")
		if strings.Contains(joined, "issues/1/comments") && strings.Contains(joined, "body=") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected gh api create comment call")
	}
}

func TestPostToGitHubPR_UpdateExistingComment(t *testing.T) {
	t.Parallel()

	mock := newMockRunner()
	// Existing comment found with ID 42
	mock.outputs["--jq"] = mockResult{output: []byte("42"), err: nil}

	reporter.SetCommandRunner(mock)
	t.Cleanup(func() { reporter.SetCommandRunner(&reporter.ExecCommandRunner{}) })

	breaking := &buf.BreakingReport{
		TotalCount: 1,
		Changes: []buf.BreakingChange{
			{File: "a.proto", Line: 1, Category: buf.CategoryFieldRemoved, Severity: buf.SeverityHigh, Description: "field removed"},
		},
	}
	g := graph.NewGraph()
	report := reporter.AnalyzeImpact(breaking, g)

	body, err := reporter.PostToGitHubPR(report, "testowner", "testrepo", 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(body, "grpc-contract-guardian") {
		t.Error("body missing guardian marker")
	}

	// Verify that gh api was called to update (PATCH) the comment
	found := false
	for _, call := range mock.calls {
		joined := strings.Join(call, " ")
		if strings.Contains(joined, "PATCH") && strings.Contains(joined, "comments/42") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected gh api update comment call with PATCH")
	}
}

func TestPostToGitHubPR_CreateCommentError(t *testing.T) {
	t.Parallel()

	mock := newMockRunner()
	// findExistingComment returns empty (no existing comment)
	mock.outputs["--jq"] = mockResult{output: []byte(""), err: nil}
	// createComment fails
	mock.outputs["body="] = mockResult{output: []byte("forbidden"), err: fmt.Errorf("exit status 1")}

	reporter.SetCommandRunner(mock)
	t.Cleanup(func() { reporter.SetCommandRunner(&reporter.ExecCommandRunner{}) })

	breaking := &buf.BreakingReport{TotalCount: 0}
	g := graph.NewGraph()
	report := reporter.AnalyzeImpact(breaking, g)

	_, err := reporter.PostToGitHubPR(report, "testowner", "testrepo", 1, false)
	if err == nil {
		t.Fatal("expected error when create comment fails")
	}
}
