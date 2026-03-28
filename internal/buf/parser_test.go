package buf_test

import (
	"testing"

	"github.com/akaitigo/grpc-contract-guardian/internal/buf"
)

func TestParseOutput_Empty(t *testing.T) {
	t.Parallel()

	report, err := buf.ParseOutput("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalCount != 0 {
		t.Errorf("expected 0 changes, got %d", report.TotalCount)
	}
}

func TestParseOutput_SingleFieldRemoved(t *testing.T) {
	t.Parallel()

	input := `user/v1/user.proto:10:3:Previously present field "5" with name "email" on message "User" was deleted.`

	report, err := buf.ParseOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.TotalCount != 1 {
		t.Fatalf("expected 1 change, got %d", report.TotalCount)
	}

	c := report.Changes[0]
	if c.File != "user/v1/user.proto" {
		t.Errorf("file = %q, want %q", c.File, "user/v1/user.proto")
	}
	if c.Line != 10 {
		t.Errorf("line = %d, want %d", c.Line, 10)
	}
	if c.Column != 3 {
		t.Errorf("column = %d, want %d", c.Column, 3)
	}
	if c.Category != buf.CategoryFieldRemoved {
		t.Errorf("category = %q, want %q", c.Category, buf.CategoryFieldRemoved)
	}
	if c.Severity != buf.SeverityHigh {
		t.Errorf("severity = %q, want %q", c.Severity, buf.SeverityHigh)
	}
}

func TestParseOutput_MultipleChanges(t *testing.T) {
	t.Parallel()

	input := `order/v1/order.proto:5:1:Previously present service "OrderService" was deleted.
order/v1/order.proto:20:1:Previously present enum value "3" on enum "OrderStatus" was deleted.
payment/v1/payment.proto:15:3:Previously present field "2" with name "amount" on message "Payment" was deleted.`

	report, err := buf.ParseOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.TotalCount != 3 {
		t.Fatalf("expected 3 changes, got %d", report.TotalCount)
	}

	if report.Changes[0].Category != buf.CategoryServiceRemoved {
		t.Errorf("change[0] category = %q, want SERVICE_REMOVED", report.Changes[0].Category)
	}
	if report.Changes[1].Category != buf.CategoryEnumValueRemoved {
		t.Errorf("change[1] category = %q, want ENUM_VALUE_REMOVED", report.Changes[1].Category)
	}
	if report.Changes[2].Category != buf.CategoryFieldRemoved {
		t.Errorf("change[2] category = %q, want FIELD_REMOVED", report.Changes[2].Category)
	}
}

func TestParseOutput_MethodSignatureChange(t *testing.T) {
	t.Parallel()

	input := `api/v1/api.proto:30:3:Method "GetUser" on service "UserService" changed request type from "GetUserRequest" to "GetUserRequestV2".`

	report, err := buf.ParseOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.TotalCount != 1 {
		t.Fatalf("expected 1 change, got %d", report.TotalCount)
	}

	if report.Changes[0].Category != buf.CategoryMethodSignatureChanged {
		t.Errorf("category = %q, want METHOD_SIGNATURE_CHANGED", report.Changes[0].Category)
	}
}

func TestParseOutput_EntityExtraction(t *testing.T) {
	t.Parallel()

	input := `user/v1/user.proto:10:3:Previously present field "5" with name "email" on message "User" was deleted.`

	report, err := buf.ParseOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Field numbers (pure digits like "5") should be skipped; first non-numeric quoted value is "email"
	if report.Changes[0].AffectedEntity != "email" {
		t.Errorf("entity = %q, want %q", report.Changes[0].AffectedEntity, "email")
	}
}

func TestBreakingReport_CountByCategory(t *testing.T) {
	t.Parallel()

	report := &buf.BreakingReport{
		Changes: []buf.BreakingChange{
			{Category: buf.CategoryFieldRemoved},
			{Category: buf.CategoryFieldRemoved},
			{Category: buf.CategoryServiceRemoved},
		},
		TotalCount: 3,
	}

	counts := report.CountByCategory()
	if counts[buf.CategoryFieldRemoved] != 2 {
		t.Errorf("FIELD_REMOVED count = %d, want 2", counts[buf.CategoryFieldRemoved])
	}
	if counts[buf.CategoryServiceRemoved] != 1 {
		t.Errorf("SERVICE_REMOVED count = %d, want 1", counts[buf.CategoryServiceRemoved])
	}
}

func TestBreakingReport_CountBySeverity(t *testing.T) {
	t.Parallel()

	report := &buf.BreakingReport{
		Changes: []buf.BreakingChange{
			{Severity: buf.SeverityHigh},
			{Severity: buf.SeverityHigh},
			{Severity: buf.SeverityMedium},
		},
		TotalCount: 3,
	}

	counts := report.CountBySeverity()
	if counts[buf.SeverityHigh] != 2 {
		t.Errorf("high count = %d, want 2", counts[buf.SeverityHigh])
	}
}

func TestBreakingReport_HasHighSeverity(t *testing.T) {
	t.Parallel()

	withHigh := &buf.BreakingReport{
		Changes: []buf.BreakingChange{{Severity: buf.SeverityHigh}},
	}
	if !withHigh.HasHighSeverity() {
		t.Error("expected HasHighSeverity() = true")
	}

	withoutHigh := &buf.BreakingReport{
		Changes: []buf.BreakingChange{{Severity: buf.SeverityLow}},
	}
	if withoutHigh.HasHighSeverity() {
		t.Error("expected HasHighSeverity() = false")
	}
}

func TestParseOutput_SkipsEmptyLines(t *testing.T) {
	t.Parallel()

	input := `
user/v1/user.proto:10:3:Previously present field "5" on message "User" was deleted.

order/v1/order.proto:5:1:Previously present service "OrderService" was deleted.
`
	report, err := buf.ParseOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.TotalCount != 2 {
		t.Errorf("expected 2 changes, got %d", report.TotalCount)
	}
}
