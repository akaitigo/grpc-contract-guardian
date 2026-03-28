package main

import (
	"testing"
)

func TestNewCheckCmd_DefaultFlags(t *testing.T) {
	t.Parallel()

	cmd := newCheckCmd()

	if cmd.Use != "check" {
		t.Errorf("Use = %q, want %q", cmd.Use, "check")
	}

	against, err := cmd.Flags().GetString("against")
	if err != nil {
		t.Fatalf("getting --against flag: %v", err)
	}
	if against != "main" {
		t.Errorf("default --against = %q, want %q", against, "main")
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Fatalf("getting --format flag: %v", err)
	}
	if format != "text" {
		t.Errorf("default --format = %q, want %q", format, "text")
	}

	protoRoot, err := cmd.Flags().GetString("proto-root")
	if err != nil {
		t.Fatalf("getting --proto-root flag: %v", err)
	}
	if protoRoot != "." {
		t.Errorf("default --proto-root = %q, want %q", protoRoot, ".")
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		t.Fatalf("getting --dry-run flag: %v", err)
	}
	if dryRun {
		t.Error("default --dry-run should be false")
	}
}

func TestNewCheckCmd_InvalidBranch(t *testing.T) {
	t.Parallel()

	cmd := newCheckCmd()
	cmd.SetArgs([]string{"--against", "invalid branch name!"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid branch name")
	}
}

func TestNewCheckCmd_InvalidRepoFormat(t *testing.T) {
	t.Parallel()

	cmd := newCheckCmd()
	cmd.SetArgs([]string{"--format", "github", "--repo", "invalid-repo", "--pr", "1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --repo format")
	}
}

func TestNewCheckCmd_GithubFormatRequiresPR(t *testing.T) {
	t.Parallel()

	cmd := newCheckCmd()
	// Use a non-existent proto-root to avoid buf execution
	cmd.SetArgs([]string{"--format", "github", "--repo", "owner/repo", "--proto-root", "/nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --pr is not set for github format")
	}
}

func TestBranchNameValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"simple", "main", true},
		{"with-dash", "feature/my-branch", true},
		{"with-dot", "v1.0", true},
		{"with-spaces", "bad branch", false},
		{"with-special", "branch;rm -rf", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := branchNameRe.MatchString(tt.input)
			if got != tt.valid {
				t.Errorf("branchNameRe.MatchString(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestFindProtoFiles_NonexistentDir(t *testing.T) {
	t.Parallel()

	_, err := findProtoFiles("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}
