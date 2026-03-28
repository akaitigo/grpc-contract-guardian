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
	outputs map[string]mockResult
	calls   [][]string
}

type mockResult struct {
	err    error
	output []byte
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

func (m *mockRunner) hasCallContaining(substrs ...string) bool {
	for _, call := range m.calls {
		joined := strings.Join(call, " ")
		allMatch := true
		for _, sub := range substrs {
			if !strings.Contains(joined, sub) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return true
		}
	}
	return false
}

func setupMockRunner(t *testing.T, mock *mockRunner) {
	t.Helper()
	reporter.SetCommandRunner(mock)
	t.Cleanup(func() { reporter.SetCommandRunner(&reporter.ExecCommandRunner{}) })
}

func buildEmptyReport() *reporter.ImpactReport {
	return reporter.AnalyzeImpact(&buf.BreakingReport{TotalCount: 0}, graph.NewGraph())
}

func buildReportWithChange() *reporter.ImpactReport {
	breaking := &buf.BreakingReport{
		TotalCount: 1,
		Changes: []buf.BreakingChange{
			{File: "a.proto", Line: 1, Category: buf.CategoryFieldRemoved, Severity: buf.SeverityHigh, Description: "field removed"},
		},
	}
	return reporter.AnalyzeImpact(breaking, graph.NewGraph())
}

func TestGitHubReporter_CreateComment(t *testing.T) {
	mock := newMockRunner()
	mock.outputs["issues/1/comments"] = mockResult{output: []byte(""), err: nil}
	setupMockRunner(t, mock)

	body, err := reporter.PostToGitHubPR(buildEmptyReport(), "testowner", "testrepo", 1, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(body, "grpc-contract-guardian") {
		t.Error("body missing guardian marker")
	}
	if !mock.hasCallContaining("issues/1/comments", "body=") {
		t.Error("expected gh api create comment call")
	}
}

func TestGitHubReporter_UpdateExistingComment(t *testing.T) {
	mock := newMockRunner()
	mock.outputs["--jq"] = mockResult{output: []byte("42"), err: nil}
	setupMockRunner(t, mock)

	body, err := reporter.PostToGitHubPR(buildReportWithChange(), "testowner", "testrepo", 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(body, "grpc-contract-guardian") {
		t.Error("body missing guardian marker")
	}
	if !mock.hasCallContaining("PATCH", "comments/42") {
		t.Error("expected gh api update comment call with PATCH")
	}
}

func TestGitHubReporter_CreateCommentError(t *testing.T) {
	mock := newMockRunner()
	mock.outputs["--jq"] = mockResult{output: []byte(""), err: nil}
	mock.outputs["body="] = mockResult{output: []byte("forbidden"), err: fmt.Errorf("exit status 1")}
	setupMockRunner(t, mock)

	_, err := reporter.PostToGitHubPR(buildEmptyReport(), "testowner", "testrepo", 1, false)
	if err == nil {
		t.Fatal("expected error when create comment fails")
	}
}
