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
	Type   string
	Number int32
}

var (
	packageRe = regexp.MustCompile(`^package\s+([\w.]+)\s*;`)
	importRe  = regexp.MustCompile(`^import\s+"([^"]+)"\s*;`)
	serviceRe = regexp.MustCompile(`^service\s+(\w+)\s*\{`)
	rpcRe     = regexp.MustCompile(`^\s*rpc\s+(\w+)\s*\(\s*(\w+)\s*\)\s*returns\s*\(\s*([\w.]+)\s*\)`)
	messageRe = regexp.MustCompile(`^message\s+(\w+)\s*\{`)
	fieldRe   = regexp.MustCompile(`^\s*(repeated\s+)?(\w+(?:\.\w+)*)\s+(\w+)\s*=\s*(\d+)`)
)

// parseState holds mutable state during proto file parsing.
type parseState struct {
	pf         *ProtoFile
	curSvc     *Service
	curMsg     *Message
	braceDepth int
	inService  bool
	inMessage  bool
}

// Analyze parses a .proto file and returns its structured representation.
func Analyze(path string) (*ProtoFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	s := &parseState{pf: &ProtoFile{Path: path}}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "syntax") {
			continue
		}

		s.processLine(line)
	}

	s.flushPending()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", path, err)
	}

	return s.pf, nil
}

func (s *parseState) processLine(line string) {
	if m := packageRe.FindStringSubmatch(line); m != nil {
		s.pf.Package = m[1]
		return
	}

	if m := importRe.FindStringSubmatch(line); m != nil {
		s.pf.Imports = append(s.pf.Imports, m[1])
		return
	}

	s.braceDepth += strings.Count(line, "{") - strings.Count(line, "}")

	if s.braceDepth <= 0 {
		s.closeBlock()
		s.braceDepth = 0
		return
	}

	s.parseContent(line)

	if line == "}" {
		s.closeBlock()
	}
}

func (s *parseState) closeBlock() {
	if s.inService && s.curSvc != nil {
		s.pf.Services = append(s.pf.Services, *s.curSvc)
		s.curSvc = nil
		s.inService = false
	}

	if s.inMessage && s.curMsg != nil {
		s.pf.Messages = append(s.pf.Messages, *s.curMsg)
		s.curMsg = nil
		s.inMessage = false
	}
}

func (s *parseState) parseContent(line string) {
	if !s.inService && !s.inMessage {
		if m := serviceRe.FindStringSubmatch(line); m != nil {
			s.inService = true
			s.curSvc = &Service{Name: m[1]}
			return
		}

		if m := messageRe.FindStringSubmatch(line); m != nil {
			s.inMessage = true
			s.curMsg = &Message{Name: m[1]}
			return
		}
	}

	if s.inService && s.curSvc != nil {
		if m := rpcRe.FindStringSubmatch(line); m != nil {
			s.curSvc.Methods = append(s.curSvc.Methods, Method{
				Name:       m[1],
				InputType:  m[2],
				OutputType: m[3],
			})
			return
		}
	}

	if s.inMessage && s.curMsg != nil {
		s.parseField(line)
	}
}

func (s *parseState) parseField(line string) {
	m := fieldRe.FindStringSubmatch(line)
	if m == nil {
		return
	}

	var num int32
	if _, err := fmt.Sscanf(m[4], "%d", &num); err == nil {
		s.curMsg.Fields = append(s.curMsg.Fields, Field{
			Name:   m[3],
			Type:   strings.TrimPrefix(m[2], "repeated "),
			Number: num,
		})
	}
}

// flushPending closes any unclosed blocks at EOF.
func (s *parseState) flushPending() {
	if s.curSvc != nil {
		s.pf.Services = append(s.pf.Services, *s.curSvc)
	}

	if s.curMsg != nil {
		s.pf.Messages = append(s.pf.Messages, *s.curMsg)
	}
}

// AnalyzeAll parses multiple .proto files and returns all results.
func AnalyzeAll(paths []string) ([]*ProtoFile, error) {
	results := make([]*ProtoFile, 0, len(paths))
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
	return !primitives[t] && t != ""
}
