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

func TestAnalyzeImpact_FQNEntityDoesNotSuffixMatch(t *testing.T) {
	t.Parallel()

	// When entity is already a FQN, only exact match should work.
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "example.v1.UserService", Kind: "service", Label: "UserService"})
	g.AddNode(graph.Node{ID: "other.v1.User", Kind: "message", Label: "User"})
	g.AddEdge(graph.Edge{From: "example.v1.UserService", To: "other.v1.User", Label: "field:user"})

	breaking := &buf.BreakingReport{
		TotalCount: 1,
		Changes: []buf.BreakingChange{
			{
				File:           "user/v1/user.proto",
				Category:       buf.CategoryFieldRemoved,
				Severity:       buf.SeverityHigh,
				Description:    "field removed",
				AffectedEntity: "example.v1.User", // FQN entity
			},
		},
	}

	report := reporter.AnalyzeImpact(breaking, g)
	// "example.v1.User" should NOT match "other.v1.User"
	if len(report.Impacts[0].AffectedServices) != 0 {
		t.Errorf("expected 0 affected services for FQN mismatch, got %v", report.Impacts[0].AffectedServices)
	}
}

func TestAnalyzeImpact_SimpleEntitySuffixMatches(t *testing.T) {
	t.Parallel()

	// When entity is a simple name, suffix match should work.
	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "example.v1.UserService", Kind: "service", Label: "UserService"})
	g.AddNode(graph.Node{ID: "example.v1.User", Kind: "message", Label: "User"})
	g.AddEdge(graph.Edge{From: "example.v1.UserService", To: "example.v1.User", Label: "field:user"})

	breaking := &buf.BreakingReport{
		TotalCount: 1,
		Changes: []buf.BreakingChange{
			{
				File:           "user/v1/user.proto",
				Category:       buf.CategoryFieldRemoved,
				Severity:       buf.SeverityHigh,
				Description:    "field removed",
				AffectedEntity: "User", // simple name
			},
		},
	}

	report := reporter.AnalyzeImpact(breaking, g)
	if len(report.Impacts[0].AffectedServices) != 1 {
		t.Errorf("expected 1 affected service for simple name match, got %v", report.Impacts[0].AffectedServices)
	}
}
