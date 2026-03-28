package main

import (
	"bytes"
	"testing"
)

func TestNewGraphCmd_DefaultFlags(t *testing.T) {
	t.Parallel()

	cmd := newGraphCmd()

	if cmd.Use != "graph" {
		t.Errorf("Use = %q, want %q", cmd.Use, "graph")
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		t.Fatalf("getting --output flag: %v", err)
	}
	if output != "text" {
		t.Errorf("default --output = %q, want %q", output, "text")
	}

	protoRoot, err := cmd.Flags().GetString("proto-root")
	if err != nil {
		t.Fatalf("getting --proto-root flag: %v", err)
	}
	if protoRoot != "." {
		t.Errorf("default --proto-root = %q, want %q", protoRoot, ".")
	}
}

func TestNewGraphCmd_TextOutput(t *testing.T) {
	t.Parallel()

	cmd := newGraphCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--proto-root", "../../testdata", "--output", "text"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewGraphCmd_DotOutput(t *testing.T) {
	t.Parallel()

	cmd := newGraphCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--proto-root", "../../testdata", "--output", "dot"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewGraphCmd_InvalidFormat(t *testing.T) {
	t.Parallel()

	cmd := newGraphCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--proto-root", "../../testdata", "--output", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid output format")
	}
}

func TestNewGraphCmd_NoProtoFiles(t *testing.T) {
	t.Parallel()

	cmd := newGraphCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	// Use a temp dir with no proto files
	cmd.SetArgs([]string{"--proto-root", t.TempDir(), "--output", "text"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no .proto files found")
	}
}
