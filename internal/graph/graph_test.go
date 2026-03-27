package graph_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/akaitigo/grpc-contract-guardian/internal/analyzer"
	"github.com/akaitigo/grpc-contract-guardian/internal/graph"
)

func TestNewGraph_Empty(t *testing.T) {
	t.Parallel()

	g := graph.NewGraph()
	if len(g.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(g.Edges))
	}
}

func TestAddNode_Deduplication(t *testing.T) {
	t.Parallel()

	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "svc.v1.User", Kind: "service", Label: "UserService"})
	g.AddNode(graph.Node{ID: "svc.v1.User", Kind: "service", Label: "UserService"})

	if len(g.Nodes) != 1 {
		t.Errorf("expected 1 node after dedup, got %d", len(g.Nodes))
	}
}

func TestWriteDOT(t *testing.T) {
	t.Parallel()

	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "svc", Kind: "service", Label: "MyService"})
	g.AddNode(graph.Node{ID: "msg", Kind: "message", Label: "Request"})
	g.AddEdge(graph.Edge{From: "svc", To: "msg", Label: "input"})

	var buf bytes.Buffer
	if err := g.WriteDOT(&buf); err != nil {
		t.Fatalf("WriteDOT error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "digraph dependencies") {
		t.Error("DOT output missing 'digraph dependencies'")
	}
	if !strings.Contains(output, "MyService") {
		t.Error("DOT output missing node label")
	}
	if !strings.Contains(output, "input") {
		t.Error("DOT output missing edge label")
	}
}

func TestWriteText(t *testing.T) {
	t.Parallel()

	g := graph.NewGraph()
	g.AddNode(graph.Node{ID: "svc", Kind: "service", Label: "UserService"})
	g.AddNode(graph.Node{ID: "msg", Kind: "message", Label: "GetUserRequest"})
	g.AddEdge(graph.Edge{From: "svc", To: "msg", Label: "uses"})

	var buf bytes.Buffer
	if err := g.WriteText(&buf); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[SERVICE]") {
		t.Error("text output missing [SERVICE]")
	}
	if !strings.Contains(output, "UserService") {
		t.Error("text output missing service name")
	}
}

func TestBuildFromProtoFiles(t *testing.T) {
	t.Parallel()

	files := []*analyzer.ProtoFile{
		{
			Package: "example.v1",
			Services: []analyzer.Service{
				{
					Name: "UserService",
					Methods: []analyzer.Method{
						{Name: "GetUser", InputType: "GetUserRequest", OutputType: "GetUserResponse"},
					},
				},
			},
			Messages: []analyzer.Message{
				{Name: "GetUserRequest", Fields: []analyzer.Field{{Name: "id", Number: 1, Type: "string"}}},
				{Name: "GetUserResponse", Fields: []analyzer.Field{{Name: "user", Number: 1, Type: "User"}}},
				{Name: "User", Fields: []analyzer.Field{
					{Name: "id", Number: 1, Type: "string"},
					{Name: "name", Number: 2, Type: "string"},
				}},
			},
		},
	}

	g := graph.BuildFromProtoFiles(files)

	// Should have service + 3 messages = 4 nodes
	if len(g.Nodes) < 4 {
		t.Errorf("expected at least 4 nodes, got %d", len(g.Nodes))
	}

	// Should have edges: service→input, service→output, GetUserResponse→User
	if len(g.Edges) < 3 {
		t.Errorf("expected at least 3 edges, got %d", len(g.Edges))
	}

	// Verify DOT output works
	var buf bytes.Buffer
	if err := g.WriteDOT(&buf); err != nil {
		t.Fatalf("WriteDOT error: %v", err)
	}
	if !strings.Contains(buf.String(), "UserService") {
		t.Error("DOT output missing UserService")
	}
}

func TestBuildFromProtoFiles_Integration(t *testing.T) {
	t.Parallel()

	pf, err := analyzer.Analyze("../../testdata/user.proto")
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	g := graph.BuildFromProtoFiles([]*analyzer.ProtoFile{pf})

	if len(g.Nodes) == 0 {
		t.Fatal("expected nodes from user.proto")
	}
	if len(g.Edges) == 0 {
		t.Fatal("expected edges from user.proto")
	}

	// Verify text output has the service
	var buf bytes.Buffer
	if err := g.WriteText(&buf); err != nil {
		t.Fatalf("WriteText error: %v", err)
	}
	if !strings.Contains(buf.String(), "UserService") {
		t.Error("text output missing UserService from integration test")
	}
}
