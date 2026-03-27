// Package analyzer provides proto file parsing and dependency extraction.
// It reads .proto files and builds a structured representation of services,
// messages, and their field-level dependencies.
package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ProtoFile represents a parsed .proto file with its services and messages.
type ProtoFile struct {
	Path     string
	Package  string
	Services []Service
	Messages []Message
	Imports  []string
}

// Service represents a gRPC service definition.
type Service struct {
	Name    string
	Methods []Method
}

// Method represents an RPC method in a service.
type Method struct {
	Name       string
	InputType  string
	OutputType string
}

// Message represents a protobuf message definition.
type Message struct {
	Name   string
	Fields []Field
}

// Field represents a field in a protobuf message.
type Field struct {
	Name   string
	Number int32
	Type   string
}

var (
	packageRe = regexp.MustCompile(`^package\s+([\w.]+)\s*;`)
	importRe  = regexp.MustCompile(`^import\s+"([^"]+)"\s*;`)
	serviceRe = regexp.MustCompile(`^service\s+(\w+)\s*\{`)
	rpcRe     = regexp.MustCompile(`^\s*rpc\s+(\w+)\s*\(\s*(\w+)\s*\)\s*returns\s*\(\s*([\w.]+)\s*\)`)
	messageRe = regexp.MustCompile(`^message\s+(\w+)\s*\{`)
	fieldRe   = regexp.MustCompile(`^\s*(repeated\s+)?(\w+(?:\.\w+)*)\s+(\w+)\s*=\s*(\d+)`)
)

// Analyze parses a .proto file and returns its structured representation.
func Analyze(path string) (*ProtoFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	pf := &ProtoFile{Path: path}
	scanner := bufio.NewScanner(f)

	var (
		inService bool
		inMessage bool
		curSvc    *Service
		curMsg    *Message
		braceDepth int
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "syntax") {
			continue
		}

		// Package and import are top-level, parse before brace tracking
		if m := packageRe.FindStringSubmatch(line); m != nil {
			pf.Package = m[1]
			continue
		}

		// Import
		if m := importRe.FindStringSubmatch(line); m != nil {
			pf.Imports = append(pf.Imports, m[1])
			continue
		}

		// Track brace depth for nested scope
		braceDepth += strings.Count(line, "{") - strings.Count(line, "}")

		if braceDepth <= 0 {
			if inService && curSvc != nil {
				pf.Services = append(pf.Services, *curSvc)
				curSvc = nil
				inService = false
			}
			if inMessage && curMsg != nil {
				pf.Messages = append(pf.Messages, *curMsg)
				curMsg = nil
				inMessage = false
			}
			braceDepth = 0
			continue
		}

		// Service start
		if !inService && !inMessage {
			if m := serviceRe.FindStringSubmatch(line); m != nil {
				inService = true
				curSvc = &Service{Name: m[1]}
				continue
			}
		}

		// Message start
		if !inService && !inMessage {
			if m := messageRe.FindStringSubmatch(line); m != nil {
				inMessage = true
				curMsg = &Message{Name: m[1]}
				continue
			}
		}

		// RPC method inside service
		if inService && curSvc != nil {
			if m := rpcRe.FindStringSubmatch(line); m != nil {
				curSvc.Methods = append(curSvc.Methods, Method{
					Name:       m[1],
					InputType:  m[2],
					OutputType: m[3],
				})
				continue
			}
		}

		// Field inside message
		if inMessage && curMsg != nil {
			if m := fieldRe.FindStringSubmatch(line); m != nil {
				var num int32
				if _, err := fmt.Sscanf(m[4], "%d", &num); err == nil {
					curMsg.Fields = append(curMsg.Fields, Field{
						Name:   m[3],
						Number: num,
						Type:   strings.TrimPrefix(m[2], "repeated "),
					})
				}
				continue
			}
		}

		// Closing brace for service/message
		if line == "}" {
			if inService && curSvc != nil {
				pf.Services = append(pf.Services, *curSvc)
				curSvc = nil
				inService = false
			}
			if inMessage && curMsg != nil {
				pf.Messages = append(pf.Messages, *curMsg)
				curMsg = nil
				inMessage = false
			}
		}
	}

	// Handle unclosed blocks at EOF
	if curSvc != nil {
		pf.Services = append(pf.Services, *curSvc)
	}
	if curMsg != nil {
		pf.Messages = append(pf.Messages, *curMsg)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", path, err)
	}

	return pf, nil
}

// AnalyzeAll parses multiple .proto files and returns all results.
func AnalyzeAll(paths []string) ([]*ProtoFile, error) {
	var results []*ProtoFile
	for _, p := range paths {
		pf, err := Analyze(p)
		if err != nil {
			return nil, fmt.Errorf("analyzing %s: %w", p, err)
		}
		results = append(results, pf)
	}
	return results, nil
}

// MessageTypes returns a set of all message type names referenced as field types.
func (pf *ProtoFile) MessageTypes() []string {
	seen := make(map[string]bool)
	var types []string
	for _, msg := range pf.Messages {
		for _, f := range msg.Fields {
			if isMessageType(f.Type) && !seen[f.Type] {
				seen[f.Type] = true
				types = append(types, f.Type)
			}
		}
	}
	return types
}

// isMessageType returns true if the type name looks like a message reference
// (starts with uppercase, not a primitive type).
func isMessageType(t string) bool {
	primitives := map[string]bool{
		"double": true, "float": true, "int32": true, "int64": true,
		"uint32": true, "uint64": true, "sint32": true, "sint64": true,
		"fixed32": true, "fixed64": true, "sfixed32": true, "sfixed64": true,
		"bool": true, "string": true, "bytes": true,
	}
	return !primitives[t] && len(t) > 0
}
