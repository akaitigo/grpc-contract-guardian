package analyzer_test

import (
	"testing"

	"github.com/akaitigo/grpc-contract-guardian/internal/analyzer"
)

func TestAnalyze_ReturnsProtoFileWithPath(t *testing.T) {
	t.Parallel()

	path := "testdata/example.proto"
	result, err := analyzer.Analyze(path)
	if err != nil {
		t.Fatalf("Analyze(%q) returned unexpected error: %v", path, err)
	}

	if result == nil {
		t.Fatal("Analyze() returned nil")
	}

	if result.Path != path {
		t.Errorf("Analyze(%q).Path = %q, want %q", path, result.Path, path)
	}
}

func TestProtoFile_StructFields(t *testing.T) {
	t.Parallel()

	pf := &analyzer.ProtoFile{
		Path:    "test.proto",
		Package: "example.v1",
		Services: []analyzer.Service{
			{
				Name: "ExampleService",
				Methods: []analyzer.Method{
					{
						Name:       "GetItem",
						InputType:  "example.v1.GetItemRequest",
						OutputType: "example.v1.GetItemResponse",
					},
				},
			},
		},
		Messages: []analyzer.Message{
			{
				Name: "GetItemRequest",
				Fields: []analyzer.Field{
					{Name: "id", Number: 1, Type: "string"},
				},
			},
		},
	}

	if len(pf.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(pf.Services))
	}

	if pf.Services[0].Name != "ExampleService" {
		t.Errorf("service name = %q, want %q", pf.Services[0].Name, "ExampleService")
	}

	if len(pf.Services[0].Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(pf.Services[0].Methods))
	}

	if pf.Services[0].Methods[0].Name != "GetItem" {
		t.Errorf("method name = %q, want %q", pf.Services[0].Methods[0].Name, "GetItem")
	}

	if len(pf.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(pf.Messages))
	}

	if len(pf.Messages[0].Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(pf.Messages[0].Fields))
	}

	if pf.Messages[0].Fields[0].Number != 1 {
		t.Errorf("field number = %d, want %d", pf.Messages[0].Fields[0].Number, 1)
	}
}
