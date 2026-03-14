package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/DriftAgent/api-drift-engine/pkg/schema"
)

// Format controls the output format of the report.
type Format string

const (
	FormatJSON     Format = "json"
	FormatText     Format = "text"
	FormatGitHub   Format = "github"   // GitHub Actions workflow commands
	FormatMarkdown Format = "markdown" // GitHub-flavored markdown table
)

// Write outputs the DiffResult to w in the requested format.
func Write(w io.Writer, result schema.DiffResult, format Format) error {
	switch format {
	case FormatJSON:
		return writeJSON(w, result)
	case FormatGitHub:
		return writeGitHub(w, result)
	case FormatMarkdown:
		return writeMarkdown(w, result)
	default:
		return writeText(w, result)
	}
}

func writeJSON(w io.Writer, result schema.DiffResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func writeText(w io.Writer, result schema.DiffResult) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprintf(tw, "Total: %d\tBreaking: %d\tNon-Breaking: %d\tInfo: %d\n\n",
		result.Summary.Total,
		result.Summary.Breaking,
		result.Summary.NonBreaking,
		result.Summary.Info,
	)

	if len(result.Changes) == 0 {
		fmt.Fprintln(tw, "No changes detected.")
		return tw.Flush()
	}

	fmt.Fprintf(tw, "SEVERITY\tTYPE\tPATH\tDESCRIPTION\n")
	fmt.Fprintf(tw, "%s\n", strings.Repeat("-", 100))

	for _, c := range result.Changes {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			severityLabel(c.Severity),
			c.Type,
			c.Path,
			c.Description,
		)
	}

	return tw.Flush()
}

func writeMarkdown(w io.Writer, result schema.DiffResult) error {
	fmt.Fprintf(w, "**Total: %d** | Breaking: %d | Non-Breaking: %d | Info: %d\n\n",
		result.Summary.Total,
		result.Summary.Breaking,
		result.Summary.NonBreaking,
		result.Summary.Info,
	)

	if len(result.Changes) == 0 {
		fmt.Fprintln(w, "No changes detected.")
		return nil
	}

	fmt.Fprintln(w, "| Severity | Type | Path | Description |")
	fmt.Fprintln(w, "|----------|------|------|-------------|")

	for _, c := range result.Changes {
		fmt.Fprintf(w, "| %s | %s | %s | %s |\n",
			severityLabel(c.Severity),
			c.Type,
			c.Path,
			c.Description,
		)
	}

	return nil
}

// writeGitHub emits GitHub Actions workflow commands so that breaking changes
// appear as error annotations and non-breaking changes appear as warnings
// directly on the PR diff.
func writeGitHub(w io.Writer, result schema.DiffResult) error {
	for _, c := range result.Changes {
		switch c.Severity {
		case schema.SeverityBreaking:
			fmt.Fprintf(w, "::error title=Breaking Change::%s\n", c.Description)
		case schema.SeverityNonBreaking:
			fmt.Fprintf(w, "::warning title=Non-Breaking Change::%s\n", c.Description)
		default:
			fmt.Fprintf(w, "::notice title=Info::%s\n", c.Description)
		}
	}

	if result.Summary.Breaking > 0 {
		fmt.Fprintf(w, "::error title=API Contract Violation::%d breaking change(s) detected between %s and %s\n",
			result.Summary.Breaking, result.BaseFile, result.HeadFile)
	}

	return nil
}

func severityLabel(s schema.Severity) string {
	switch s {
	case schema.SeverityBreaking:
		return "[BREAKING]"
	case schema.SeverityNonBreaking:
		return "[non-breaking]"
	default:
		return "[info]"
	}
}

// HasBreakingChanges returns true if the result contains any breaking changes.
func HasBreakingChanges(result schema.DiffResult) bool {
	return result.Summary.Breaking > 0
}
