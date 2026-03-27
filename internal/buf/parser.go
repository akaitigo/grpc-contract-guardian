// Package buf parses the output of `buf breaking` and structures it into
// categorized breaking changes for downstream analysis and reporting.
package buf

// ChangeCategory classifies the type of breaking change.
type ChangeCategory string

const (
	// CategoryFieldRemoved indicates a field was removed from a message.
	CategoryFieldRemoved ChangeCategory = "FIELD_REMOVED"
	// CategoryFieldTypeChanged indicates a field's type was changed.
	CategoryFieldTypeChanged ChangeCategory = "FIELD_TYPE_CHANGED"
	// CategoryFieldReserved indicates a field number conflict with reserved range.
	CategoryFieldReserved ChangeCategory = "FIELD_RESERVED"
	// CategoryServiceRemoved indicates a service was removed.
	CategoryServiceRemoved ChangeCategory = "SERVICE_REMOVED"
	// CategoryMethodRemoved indicates an RPC method was removed.
	CategoryMethodRemoved ChangeCategory = "METHOD_REMOVED"
	// CategoryMethodSignatureChanged indicates a method's input/output type changed.
	CategoryMethodSignatureChanged ChangeCategory = "METHOD_SIGNATURE_CHANGED"
	// CategoryMessageRemoved indicates a message was removed.
	CategoryMessageRemoved ChangeCategory = "MESSAGE_REMOVED"
	// CategoryEnumRemoved indicates an enum was removed.
	CategoryEnumRemoved ChangeCategory = "ENUM_REMOVED"
	// CategoryEnumValueRemoved indicates an enum value was removed.
	CategoryEnumValueRemoved ChangeCategory = "ENUM_VALUE_REMOVED"
	// CategoryUnknown indicates an unrecognized change type.
	CategoryUnknown ChangeCategory = "UNKNOWN"
)

// BreakingChange represents a single breaking change detected by buf.
type BreakingChange struct {
	// File is the proto file where the change was detected.
	File string
	// Line is the line number in the proto file (0 if unknown).
	Line int
	// Category classifies the type of breaking change.
	Category ChangeCategory
	// Description is the human-readable description from buf.
	Description string
	// AffectedEntity is the fully qualified name of the affected proto entity.
	AffectedEntity string
}

// BreakingReport is the structured result of parsing buf breaking output.
type BreakingReport struct {
	// Changes is the list of all detected breaking changes.
	Changes []BreakingChange
	// TotalCount is the total number of breaking changes.
	TotalCount int
}

// ParseOutput parses the raw text output of `buf breaking` into a structured report.
// Returns an error if the output cannot be parsed.
func ParseOutput(rawOutput string) (*BreakingReport, error) {
	// Placeholder: will be implemented in Issue #3
	return &BreakingReport{
		Changes:    nil,
		TotalCount: 0,
	}, nil
}
