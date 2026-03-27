package analyzer_test

import (
	"testing"

	"github.com/akaitigo/grpc-contract-guardian/internal/analyzer"
)

func TestAnalyze_ReturnsProtoFileWithPath(t *testing.T) {
	t.Parallel()

	result, err := analyzer.Analyze("../../testdata/user.proto")
	if err != nil {
		t.Fatalf("Analyze returned unexpected error: %v", err)
	}

	if result.Path != "../../testdata/user.proto" {
		t.Errorf("Path = %q, want testdata path", result.Path)
	}
}

func TestAnalyze_Package(t *testing.T) {
	t.Parallel()

	result, err := analyzer.Analyze("../../testdata/user.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Package != "example.user.v1" {
		t.Errorf("Package = %q, want %q", result.Package, "example.user.v1")
	}
}

func TestAnalyze_Services(t *testing.T) {
	t.Parallel()

	result, err := analyzer.Analyze("../../testdata/user.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(result.Services))
	}

	svc := result.Services[0]
	if svc.Name != "UserService" {
		t.Errorf("service name = %q, want %q", svc.Name, "UserService")
	}

	if len(svc.Methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(svc.Methods))
	}

	if svc.Methods[0].Name != "GetUser" {
		t.Errorf("method[0] = %q, want %q", svc.Methods[0].Name, "GetUser")
	}
	if svc.Methods[0].InputType != "GetUserRequest" {
		t.Errorf("method[0].InputType = %q, want %q", svc.Methods[0].InputType, "GetUserRequest")
	}
	if svc.Methods[0].OutputType != "GetUserResponse" {
		t.Errorf("method[0].OutputType = %q, want %q", svc.Methods[0].OutputType, "GetUserResponse")
	}
}

func TestAnalyze_Messages(t *testing.T) {
	t.Parallel()

	result, err := analyzer.Analyze("../../testdata/user.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// user.proto has: GetUserRequest, GetUserResponse, ListUsersRequest, ListUsersResponse, User, Address
	if len(result.Messages) < 5 {
		t.Fatalf("expected at least 5 messages, got %d", len(result.Messages))
	}

	// Find User message
	var userMsg *analyzer.Message
	for i := range result.Messages {
		if result.Messages[i].Name == "User" {
			userMsg = &result.Messages[i]
			break
		}
	}

	if userMsg == nil {
		t.Fatal("User message not found")
	}

	if len(userMsg.Fields) != 4 {
		t.Fatalf("User fields: expected 4, got %d", len(userMsg.Fields))
	}

	// Check that Address field is detected as message type reference
	var hasAddress bool
	for _, f := range userMsg.Fields {
		if f.Name == "address" && f.Type == "Address" {
			hasAddress = true
		}
	}
	if !hasAddress {
		t.Error("User.address field with type Address not found")
	}
}

func TestAnalyze_MessageTypes(t *testing.T) {
	t.Parallel()

	result, err := analyzer.Analyze("../../testdata/user.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	types := result.MessageTypes()
	if len(types) == 0 {
		t.Fatal("expected at least 1 message type reference")
	}

	var hasUser, hasAddress bool
	for _, tt := range types {
		if tt == "User" {
			hasUser = true
		}
		if tt == "Address" {
			hasAddress = true
		}
	}

	if !hasUser {
		t.Error("expected User in message types")
	}
	if !hasAddress {
		t.Error("expected Address in message types")
	}
}

func TestAnalyze_NonexistentFile(t *testing.T) {
	t.Parallel()

	_, err := analyzer.Analyze("nonexistent.proto")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestAnalyzeAll(t *testing.T) {
	t.Parallel()

	results, err := analyzer.AnalyzeAll([]string{"../../testdata/user.proto"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
