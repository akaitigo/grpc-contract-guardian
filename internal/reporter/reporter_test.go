package reporter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/akaitigo/grpc-contract-guardian/internal/buf"
	"github.com/akaitigo/grpc-contract-guardian/internal/reporter"
)

func TestReport_NilReport(t *testing.T) {
	t.Parallel()

	var w bytes.Buffer
	err := reporter.Report(&w, nil, reporter.FormatText)
	if err == nil {
		t.Fatal("expected error for nil report")
	}
}

func TestReport_NoBreakingChanges_Text(t *testing.T) {
	t.Parallel()

	report := &buf.BreakingReport{TotalCount: 0}
	var w bytes.Buffer

	if err := reporter.Report(&w, report, reporter.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(w.String(), "No breaking changes") {
		t.Error("expected 'No breaking changes' message")
	}
}

func TestReport_WithChanges_Text(t *testing.T) {
	t.Parallel()

	report := &buf.BreakingReport{
		TotalCount: 1,
		Changes: []buf.BreakingChange{
			{
				File:        "user/v1/user.proto",
				Line:        10,
				Category:    buf.CategoryFieldRemoved,
				Description: "Field 'email' was removed",
			},
		},
	}

	var w bytes.Buffer
	if err := reporter.Report(&w, report, reporter.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := w.String()
	if !strings.Contains(output, "Breaking Changes: 1") {
		t.Error("missing change count")
	}
	if !strings.Contains(output, "FIELD_REMOVED") {
		t.Error("missing category")
	}
}

func TestReport_WithChanges_GitHub(t *testing.T) {
	t.Parallel()

	report := &buf.BreakingReport{
		TotalCount: 1,
		Changes: []buf.BreakingChange{
			{
				File:        "order/v1/order.proto",
				Line:        5,
				Category:    buf.CategoryServiceRemoved,
				Description: "Service 'OrderService' was removed",
			},
		},
	}

	var w bytes.Buffer
	if err := reporter.Report(&w, report, reporter.FormatGitHub); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := w.String()
	if !strings.Contains(output, ":warning:") {
		t.Error("GitHub output missing warning emoji")
	}
	if !strings.Contains(output, "| Category |") {
		t.Error("GitHub output missing table header")
	}
}

func TestReport_UnsupportedFormat(t *testing.T) {
	t.Parallel()

	report := &buf.BreakingReport{TotalCount: 0}
	var w bytes.Buffer

	err := reporter.Report(&w, report, "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}
