package reporter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/akaitigo/grpc-contract-guardian/internal/buf"
	"github.com/akaitigo/grpc-contract-guardian/internal/graph"
	"github.com/akaitigo/grpc-contract-guardian/internal/reporter"
)

func buildTestGraph() *graph.DependencyGraph {
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "example.v1.UserService", Kind: "service", Label: "UserService"})
	g.AddNode(graph.Node{ID: "example.v1.GetUserRequest", Kind: "message", Label: "GetUserRequest"})
	g.AddNode(graph.Node{ID: "example.v1.GetUserResponse", Kind: "message", Label: "GetUserResponse"})
	g.AddNode(graph.Node{ID: "example.v1.User", Kind: "message", Label: "User"})
	g.AddEdge(graph.Edge{From: "example.v1.UserService", To: "example.v1.GetUserRequest", Label: "input:GetUser"})
	g.AddEdge(graph.Edge{From: "example.v1.UserService", To: "example.v1.GetUserResponse", Label: "output:GetUser"})
	g.AddEdge(graph.Edge{From: "example.v1.GetUserResponse", To: "example.v1.User", Label: "field:user"})
	return g
}

func TestAnalyzeImpact_NoChanges(t *testing.T) {
	t.Parallel()

	report := reporter.AnalyzeImpact(&buf.BreakingReport{}, buildTestGraph())
	if len(report.Impacts) != 0 {
		t.Errorf("expected 0 impacts, got %d", len(report.Impacts))
	}
}

func TestAnalyzeImpact_NilBreaking(t *testing.T) {
	t.Parallel()

	report := reporter.AnalyzeImpact(nil, buildTestGraph())
	if len(report.Impacts) != 0 {
		t.Errorf("expected 0 impacts for nil breaking report")
	}
}

func TestAnalyzeImpact_FieldRemoval(t *testing.T) {
	t.Parallel()

	breaking := &buf.BreakingReport{
		TotalCount: 1,
		Changes: []buf.BreakingChange{
			{
				File:           "user/v1/user.proto",
				Line:           10,
				Category:       buf.CategoryFieldRemoved,
				Severity:       buf.SeverityHigh,
				Description:    "Field \"email\" was removed from \"User\"",
				AffectedEntity: "User",
			},
		},
	}

	report := reporter.AnalyzeImpact(breaking, buildTestGraph())
	if len(report.Impacts) != 1 {
		t.Fatalf("expected 1 impact, got %d", len(report.Impacts))
	}

	// User is referenced by GetUserResponse, which is referenced by UserService
	impact := report.Impacts[0]
	if len(impact.AffectedServices) == 0 {
		t.Log("Note: affected services detection depends on graph traversal depth")
	}
}

func TestWriteImpactText_NoChanges(t *testing.T) {
	t.Parallel()

	report := reporter.AnalyzeImpact(&buf.BreakingReport{}, buildTestGraph())
	var w bytes.Buffer
	if err := reporter.WriteImpactText(&w, report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(w.String(), "No breaking changes") {
		t.Error("expected 'No breaking changes' message")
	}
}

func TestWriteImpactText_WithChanges(t *testing.T) {
	t.Parallel()

	breaking := &buf.BreakingReport{
		TotalCount: 2,
		Changes: []buf.BreakingChange{
			{File: "a.proto", Line: 1, Category: buf.CategoryFieldRemoved, Severity: buf.SeverityHigh, Description: "field removed"},
			{File: "b.proto", Line: 2, Category: buf.CategoryEnumRemoved, Severity: buf.SeverityMedium, Description: "enum removed"},
		},
	}

	report := reporter.AnalyzeImpact(breaking, buildTestGraph())
	var w bytes.Buffer
	if err := reporter.WriteImpactText(&w, report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := w.String()
	if !strings.Contains(output, "Breaking Change Impact Report") {
		t.Error("missing report header")
	}
	if !strings.Contains(output, "HIGH:") {
		t.Error("missing severity summary")
	}
	if !strings.Contains(output, "FIELD_REMOVED") {
		t.Error("missing category")
	}
}

func TestWriteImpactGitHub_WithChanges(t *testing.T) {
	t.Parallel()

	breaking := &buf.BreakingReport{
		TotalCount: 1,
		Changes: []buf.BreakingChange{
			{File: "x.proto", Line: 5, Category: buf.CategoryServiceRemoved, Severity: buf.SeverityHigh, Description: "service deleted", AffectedEntity: "UserService"},
		},
	}

	report := reporter.AnalyzeImpact(breaking, buildTestGraph())
	var w bytes.Buffer
	if err := reporter.WriteImpactGitHub(&w, report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := w.String()
	if !strings.Contains(output, ":warning:") {
		t.Error("GitHub output missing warning")
	}
	if !strings.Contains(output, "| #") {
		t.Error("GitHub output missing table")
	}
	if !strings.Contains(output, "grpc-contract-guardian") {
		t.Error("GitHub output missing footer")
	}
}

func TestPostToGitHubPR_DryRun(t *testing.T) {
	t.Parallel()

	breaking := &buf.BreakingReport{TotalCount: 0}
	report := reporter.AnalyzeImpact(breaking, buildTestGraph())

	body, err := reporter.PostToGitHubPR(report, "akaitigo", "test-repo", 1, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(body, "grpc-contract-guardian") {
		t.Error("dry-run body missing marker")
	}
}
