package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/akaitigo/grpc-contract-guardian/internal/buf"
	"github.com/akaitigo/grpc-contract-guardian/internal/graph"
)

// ImpactReport combines breaking changes with dependency graph to show
// which services are affected by each change.
type ImpactReport struct {
	Breaking *buf.BreakingReport
	Graph    *graph.DependencyGraph
	Impacts  []Impact
}

// Impact represents the downstream effect of a breaking change.
type Impact struct {
	AffectedServices []string
	AffectedPath     []string
	Change           buf.BreakingChange
}

// AnalyzeImpact computes which services are affected by each breaking change
// using the dependency graph.
func AnalyzeImpact(breaking *buf.BreakingReport, g *graph.DependencyGraph) *ImpactReport {
	report := &ImpactReport{
		Breaking: breaking,
		Graph:    g,
	}

	if breaking == nil || len(breaking.Changes) == 0 {
		return report
	}

	for _, change := range breaking.Changes {
		impact := Impact{Change: change}

		// Find services that depend on the affected entity
		entity := change.AffectedEntity
		if entity != "" {
			impact.AffectedServices = findDependentServices(g, entity)
			impact.AffectedPath = tracePath(g, entity)
		}

		report.Impacts = append(report.Impacts, impact)
	}

	return report
}

// matchesEntity returns true if nodeID matches entity.
// If entity is already a FQN (contains "."), only exact match is accepted.
// Otherwise, nodeID must be an exact match or end with ".<entity>" (suffix match),
// preventing partial matches like "UserProfile" matching entity "User".
func matchesEntity(nodeID, entity string) bool {
	if nodeID == entity {
		return true
	}
	// If entity is a FQN, require exact match only (already checked above).
	if strings.Contains(entity, ".") {
		return false
	}
	return strings.HasSuffix(nodeID, "."+entity)
}

// findDependentServices walks the graph backwards to find all services
// that transitively depend on the given entity.
func findDependentServices(g *graph.DependencyGraph, entity string) []string {
	visited := make(map[string]bool)
	var services []string

	// Find all nodes that have edges pointing to the entity (exact or FQN suffix match)
	var queue []string
	for _, e := range g.Edges {
		if matchesEntity(e.To, entity) {
			if !visited[e.From] {
				visited[e.From] = true
				queue = append(queue, e.From)
			}
		}
	}

	// BFS to find all upstream services
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check if current is a service node
		for _, n := range g.Nodes {
			if n.ID == current && n.Kind == "service" {
				services = append(services, n.Label)
			}
		}

		// Find nodes that depend on current
		for _, e := range g.Edges {
			if e.To == current && !visited[e.From] {
				visited[e.From] = true
				queue = append(queue, e.From)
			}
		}
	}

	return services
}

// tracePath returns the dependency path from services to the affected entity.
func tracePath(g *graph.DependencyGraph, entity string) []string {
	var paths []string
	for _, e := range g.Edges {
		if matchesEntity(e.To, entity) {
			paths = append(paths, fmt.Sprintf("%s -[%s]-> %s", e.From, e.Label, e.To))
		}
	}
	return paths
}

// WriteImpactText writes the impact report in terminal-friendly text format.
func WriteImpactText(w io.Writer, report *ImpactReport) error {
	if report.Breaking == nil || report.Breaking.TotalCount == 0 {
		_, err := fmt.Fprintln(w, "No breaking changes detected.")
		return err
	}

	if _, err := fmt.Fprintf(w, "=== Breaking Change Impact Report ===\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Total: %d breaking change(s)\n\n", report.Breaking.TotalCount); err != nil {
		return err
	}

	if err := writeTextSeveritySummary(w, report.Breaking); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "\n--- Details ---"); err != nil {
		return err
	}

	return writeTextDetails(w, report.Impacts)
}

func writeTextSeveritySummary(w io.Writer, breaking *buf.BreakingReport) error {
	bySev := breaking.CountBySeverity()

	if high, ok := bySev[buf.SeverityHigh]; ok && high > 0 {
		if _, err := fmt.Fprintf(w, "  HIGH:   %d\n", high); err != nil {
			return err
		}
	}

	if med, ok := bySev[buf.SeverityMedium]; ok && med > 0 {
		if _, err := fmt.Fprintf(w, "  MEDIUM: %d\n", med); err != nil {
			return err
		}
	}

	if low, ok := bySev[buf.SeverityLow]; ok && low > 0 {
		if _, err := fmt.Fprintf(w, "  LOW:    %d\n", low); err != nil {
			return err
		}
	}

	return nil
}

func writeTextDetails(w io.Writer, impacts []Impact) error {
	for i := range impacts {
		impact := &impacts[i]
		c := impact.Change

		if _, err := fmt.Fprintf(w, "\n%d. [%s] %s:%d\n", i+1, c.Category, c.File, c.Line); err != nil {
			return err
		}

		if _, err := fmt.Fprintf(w, "   %s\n", c.Description); err != nil {
			return err
		}

		if len(impact.AffectedServices) > 0 {
			if _, err := fmt.Fprintf(w, "   Affected services: %s\n", strings.Join(impact.AffectedServices, ", ")); err != nil {
				return err
			}
		}

		for _, path := range impact.AffectedPath {
			if _, err := fmt.Fprintf(w, "   Path: %s\n", path); err != nil {
				return err
			}
		}
	}

	return nil
}

// WriteImpactGitHub writes the impact report as GitHub Markdown.
func WriteImpactGitHub(w io.Writer, report *ImpactReport) error {
	if report.Breaking == nil || report.Breaking.TotalCount == 0 {
		_, err := fmt.Fprintln(w, "## :white_check_mark: No Breaking Changes\n\nAll proto definitions are backward compatible.")
		return err
	}

	if _, err := fmt.Fprintf(w, "## :warning: Breaking Change Impact Report (%d changes)\n\n", report.Breaking.TotalCount); err != nil {
		return err
	}

	if err := writeGitHubSeveritySummary(w, report.Breaking); err != nil {
		return err
	}

	if err := writeGitHubChangeTable(w, report.Impacts); err != nil {
		return err
	}

	_, err := fmt.Fprintln(w, "\n---\n*Generated by [grpc-contract-guardian](https://github.com/akaitigo/grpc-contract-guardian)*")
	return err
}

func writeGitHubSeveritySummary(w io.Writer, breaking *buf.BreakingReport) error {
	bySev := breaking.CountBySeverity()

	if _, err := fmt.Fprintln(w, "### Severity Summary"); err != nil {
		return err
	}

	for _, sev := range []buf.Severity{buf.SeverityHigh, buf.SeverityMedium, buf.SeverityLow} {
		count, ok := bySev[sev]
		if !ok || count == 0 {
			continue
		}

		emoji := severityEmoji(sev)

		if _, err := fmt.Fprintf(w, "- %s **%s**: %d\n", emoji, strings.ToUpper(string(sev)), count); err != nil {
			return err
		}
	}

	return nil
}

func severityEmoji(sev buf.Severity) string {
	switch sev {
	case buf.SeverityMedium:
		return ":orange_circle:"
	case buf.SeverityLow:
		return ":yellow_circle:"
	default:
		return ":red_circle:"
	}
}

func writeGitHubChangeTable(w io.Writer, impacts []Impact) error {
	if _, err := fmt.Fprint(w, "\n### Changes\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "| # | Severity | Category | File | Description | Affected Services |"); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "|---|----------|----------|------|-------------|-------------------|"); err != nil {
		return err
	}

	for i := range impacts {
		impact := &impacts[i]
		c := impact.Change
		svcs := "-"

		if len(impact.AffectedServices) > 0 {
			svcs = strings.Join(impact.AffectedServices, ", ")
		}

		if _, err := fmt.Fprintf(w, "| %d | `%s` | `%s` | `%s:%d` | %s | %s |\n",
			i+1, c.Severity, c.Category, c.File, c.Line, c.Description, svcs); err != nil {
			return err
		}
	}

	return nil
}
