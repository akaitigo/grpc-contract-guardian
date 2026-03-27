// Package buf parses the output of `buf breaking` and structures it into
// categorized breaking changes for downstream analysis and reporting.
package buf

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// ChangeCategory classifies the type of breaking change.
type ChangeCategory string

const (
	CategoryFieldRemoved          ChangeCategory = "FIELD_REMOVED"
	CategoryFieldTypeChanged      ChangeCategory = "FIELD_TYPE_CHANGED"
	CategoryFieldReserved         ChangeCategory = "FIELD_RESERVED"
	CategoryServiceRemoved        ChangeCategory = "SERVICE_REMOVED"
	CategoryMethodRemoved         ChangeCategory = "METHOD_REMOVED"
	CategoryMethodSignatureChanged ChangeCategory = "METHOD_SIGNATURE_CHANGED"
	CategoryMessageRemoved        ChangeCategory = "MESSAGE_REMOVED"
	CategoryEnumRemoved           ChangeCategory = "ENUM_REMOVED"
	CategoryEnumValueRemoved      ChangeCategory = "ENUM_VALUE_REMOVED"
	CategoryUnknown               ChangeCategory = "UNKNOWN"
)

// Severity indicates the impact level of a breaking change.
type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)

// BreakingChange represents a single breaking change detected by buf.
type BreakingChange struct {
	File           string
	Line           int
	Column         int
	Category       ChangeCategory
	Severity       Severity
	Description    string
	AffectedEntity string
}

// BreakingReport is the structured result of parsing buf breaking output.
type BreakingReport struct {
	Changes    []BreakingChange
	TotalCount int
}

// categorySeverity maps categories to their default severity.
var categorySeverity = map[ChangeCategory]Severity{
	CategoryFieldRemoved:           SeverityHigh,
	CategoryFieldTypeChanged:       SeverityHigh,
	CategoryServiceRemoved:         SeverityHigh,
	CategoryMethodRemoved:          SeverityHigh,
	CategoryMethodSignatureChanged: SeverityHigh,
	CategoryMessageRemoved:         SeverityHigh,
	CategoryEnumRemoved:            SeverityMedium,
	CategoryEnumValueRemoved:       SeverityMedium,
	CategoryFieldReserved:          SeverityLow,
	CategoryUnknown:                SeverityMedium,
}

// categoryPatterns maps substrings in buf output to categories.
var categoryPatterns = []struct {
	pattern  string
	category ChangeCategory
}{
	// Signature changes must match before service/method removal (more specific first)
	{"changed request type", CategoryMethodSignatureChanged},
	{"changed response type", CategoryMethodSignatureChanged},
	{"input type changed", CategoryMethodSignatureChanged},
	{"output type changed", CategoryMethodSignatureChanged},
	{"request type changed", CategoryMethodSignatureChanged},
	{"response type changed", CategoryMethodSignatureChanged},
	// Field changes
	{"field type changed", CategoryFieldTypeChanged},
	{"changed type", CategoryFieldTypeChanged},
	{"Previously present field", CategoryFieldRemoved},
	// Service/method removal (after signature changes)
	{"Previously present service", CategoryServiceRemoved},
	{"Previously present method", CategoryMethodRemoved},
	// Message removal
	{"Previously present message", CategoryMessageRemoved},
	// Enum changes (enum value before enum)
	{"Previously present enum value", CategoryEnumValueRemoved},
	{"Previously present enum", CategoryEnumRemoved},
	// Reserved
	{"reserved", CategoryFieldReserved},
}

// ParseOutput parses the raw text output of `buf breaking` into a structured report.
// buf breaking output format: file:line:column:message
// Empty input returns a report with zero changes (no breaking changes detected).
func ParseOutput(rawOutput string) (*BreakingReport, error) {
	rawOutput = strings.TrimSpace(rawOutput)
	if rawOutput == "" {
		return &BreakingReport{Changes: nil, TotalCount: 0}, nil
	}

	var changes []BreakingChange
	scanner := bufio.NewScanner(strings.NewReader(rawOutput))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		change, err := parseLine(line)
		if err != nil {
			continue
		}

		changes = append(changes, change)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning buf output: %w", err)
	}

	return &BreakingReport{
		Changes:    changes,
		TotalCount: len(changes),
	}, nil
}

// parseLine parses a single line of buf breaking output.
// Expected format: "file:line:column:message" or "file:line:message"
func parseLine(line string) (BreakingChange, error) {
	parts := strings.SplitN(line, ":", 4)
	if len(parts) < 3 {
		return BreakingChange{}, fmt.Errorf("unexpected format: %s", line)
	}

	file := parts[0]
	lineNum, err := strconv.Atoi(parts[1])
	if err != nil {
		lineNum = 0
	}

	var col int
	var message string

	if len(parts) == 4 {
		col, err = strconv.Atoi(parts[2])
		if err != nil {
			col = 0
			message = strings.TrimSpace(parts[2] + ":" + parts[3])
		} else {
			message = strings.TrimSpace(parts[3])
		}
	} else {
		message = strings.TrimSpace(parts[2])
	}

	category := classifyChange(message)
	entity := extractEntity(message)

	return BreakingChange{
		File:           file,
		Line:           lineNum,
		Column:         col,
		Category:       category,
		Severity:       categorySeverity[category],
		Description:    message,
		AffectedEntity: entity,
	}, nil
}

// classifyChange determines the ChangeCategory from the message text.
func classifyChange(message string) ChangeCategory {
	lower := strings.ToLower(message)

	for _, p := range categoryPatterns {
		if strings.Contains(lower, strings.ToLower(p.pattern)) {
			return p.category
		}
	}

	return CategoryUnknown
}

// extractEntity attempts to extract the affected proto entity name from the message.
// buf messages typically contain quoted entity names like `"MyService.MyMethod"`.
func extractEntity(message string) string {
	start := strings.Index(message, "\"")
	if start == -1 {
		return ""
	}
	end := strings.Index(message[start+1:], "\"")
	if end == -1 {
		return ""
	}
	return message[start+1 : start+1+end]
}

// CountByCategory returns the number of changes per category.
func (r *BreakingReport) CountByCategory() map[ChangeCategory]int {
	counts := make(map[ChangeCategory]int)
	for _, c := range r.Changes {
		counts[c.Category]++
	}
	return counts
}

// CountBySeverity returns the number of changes per severity level.
func (r *BreakingReport) CountBySeverity() map[Severity]int {
	counts := make(map[Severity]int)
	for _, c := range r.Changes {
		counts[c.Severity]++
	}
	return counts
}

// HasHighSeverity returns true if any change has high severity.
func (r *BreakingReport) HasHighSeverity() bool {
	for _, c := range r.Changes {
		if c.Severity == SeverityHigh {
			return true
		}
	}
	return false
}
