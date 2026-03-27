// Package graph builds and outputs service dependency graphs from parsed proto definitions.
// It constructs a directed graph where nodes are services/messages and edges represent
// dependencies (e.g., a service method using a message type).
package graph

import (
	"fmt"
	"io"
	"strings"

	"github.com/akaitigo/grpc-contract-guardian/internal/analyzer"
)

// Node represents a node in the dependency graph.
type Node struct {
	// ID is the unique identifier (e.g., "myservice.v1.UserService").
	ID string
	// Kind is the node type: "service", "message", or "field".
	Kind string
	// Label is the human-readable display name.
	Label string
}

// Edge represents a directed dependency between two nodes.
type Edge struct {
	// From is the source node ID.
	From string
	// To is the target node ID.
	To string
	// Label describes the relationship (e.g., "uses", "input", "output").
	Label string
}

// DependencyGraph holds the service dependency graph.
type DependencyGraph struct {
	Nodes []Node
	Edges []Edge
}

// NewGraph creates an empty dependency graph.
func NewGraph() *DependencyGraph {
	return &DependencyGraph{}
}

// AddNode adds a node to the graph. Duplicate IDs are ignored.
func (g *DependencyGraph) AddNode(node Node) {
	for _, n := range g.Nodes {
		if n.ID == node.ID {
			return
		}
	}
	g.Nodes = append(g.Nodes, node)
}

// AddEdge adds a directed edge to the graph.
func (g *DependencyGraph) AddEdge(edge Edge) {
	g.Edges = append(g.Edges, edge)
}

// WriteDOT outputs the graph in Graphviz DOT format.
func (g *DependencyGraph) WriteDOT(w io.Writer) error {
	if _, err := fmt.Fprintln(w, "digraph dependencies {"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "  rankdir=LR;"); err != nil {
		return err
	}

	for _, n := range g.Nodes {
		shape := "box"
		if n.Kind == "message" {
			shape = "ellipse"
		}
		if _, err := fmt.Fprintf(w, "  %q [label=%q shape=%s];\n", n.ID, n.Label, shape); err != nil {
			return err
		}
	}

	for _, e := range g.Edges {
		if _, err := fmt.Fprintf(w, "  %q -> %q [label=%q];\n", e.From, e.To, e.Label); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintln(w, "}")
	return err
}

// WriteText outputs the graph in a human-readable text format.
func (g *DependencyGraph) WriteText(w io.Writer) error {
	for _, n := range g.Nodes {
		if n.Kind != "service" {
			continue
		}

		if _, err := fmt.Fprintf(w, "[%s] %s\n", strings.ToUpper(n.Kind), n.Label); err != nil {
			return err
		}

		for _, e := range g.Edges {
			if e.From == n.ID {
				if _, err := fmt.Fprintf(w, "  -> %s (%s)\n", e.To, e.Label); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// BuildFromProtoFiles constructs a dependency graph from parsed proto files.
func BuildFromProtoFiles(files []*analyzer.ProtoFile) *DependencyGraph {
	g := NewGraph()

	for _, pf := range files {
		prefix := pf.Package
		if prefix != "" {
			prefix += "."
		}

		// Add message nodes
		for _, msg := range pf.Messages {
			g.AddNode(Node{
				ID:    prefix + msg.Name,
				Kind:  "message",
				Label: msg.Name,
			})

			// Add edges for message-type fields
			for _, f := range msg.Fields {
				if isMessageType(f.Type) {
					target := f.Type
					if !strings.Contains(target, ".") {
						target = prefix + target
					}
					g.AddEdge(Edge{
						From:  prefix + msg.Name,
						To:    target,
						Label: "field:" + f.Name,
					})
				}
			}
		}

		// Add service nodes and method edges
		for _, svc := range pf.Services {
			svcID := prefix + svc.Name
			g.AddNode(Node{
				ID:    svcID,
				Kind:  "service",
				Label: svc.Name,
			})

			for _, m := range svc.Methods {
				inputID := m.InputType
				if !strings.Contains(inputID, ".") {
					inputID = prefix + inputID
				}
				outputID := m.OutputType
				if !strings.Contains(outputID, ".") {
					outputID = prefix + outputID
				}

				g.AddEdge(Edge{From: svcID, To: inputID, Label: "input:" + m.Name})
				g.AddEdge(Edge{From: svcID, To: outputID, Label: "output:" + m.Name})
			}
		}
	}

	return g
}

// isMessageType returns true if the type is not a protobuf primitive.
func isMessageType(t string) bool {
	primitives := map[string]bool{
		"double": true, "float": true, "int32": true, "int64": true,
		"uint32": true, "uint64": true, "sint32": true, "sint64": true,
		"fixed32": true, "fixed64": true, "sfixed32": true, "sfixed64": true,
		"bool": true, "string": true, "bytes": true,
	}
	return !primitives[t] && len(t) > 0
}
