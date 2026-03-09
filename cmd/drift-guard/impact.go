package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pgomes13/drift-guard-engine/pkg/impact"
	"github.com/pgomes13/drift-guard-engine/pkg/schema"
)

var impactCmd = &cobra.Command{
	Use:   "impact",
	Short: "Scan source files for references to breaking API changes",
	Long: `impact reads a drift-guard JSON diff and scans a source directory for
code references to each breaking change, so you know exactly which files and
lines need to be updated.

The diff JSON can be supplied via --diff or piped from another drift-guard command:

  drift-guard openapi --base old.yaml --head new.yaml --format json \
    | drift-guard impact --scan ./services`,
	Example: `  drift-guard impact --diff /tmp/diff.json --scan .
  drift-guard openapi --base base.yaml --head head.yaml --format json | drift-guard impact --scan ./services
  drift-guard impact --diff /tmp/diff.json --scan . --format markdown`,
	RunE: runImpact,
}

var (
	flagImpactDiff   string
	flagImpactScan   string
	flagImpactFormat string
)

func init() {
	impactCmd.Flags().StringVar(&flagImpactDiff, "diff", "-", "Path to JSON diff file (use - or omit to read from stdin)")
	impactCmd.Flags().StringVar(&flagImpactScan, "scan", ".", "Directory to scan for source references")
	impactCmd.Flags().StringVar(&flagImpactFormat, "format", "text", "Output format: text, json, markdown")
}

func runImpact(cmd *cobra.Command, args []string) error {
	// Read the diff JSON from file or stdin.
	var r io.Reader
	if flagImpactDiff == "" || flagImpactDiff == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(flagImpactDiff)
		if err != nil {
			return fmt.Errorf("opening diff file: %w", err)
		}
		defer f.Close()
		r = f
	}

	var result schema.DiffResult
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return fmt.Errorf("parsing diff JSON: %w", err)
	}

	// Filter to breaking changes only.
	var breaking []schema.Change
	for _, c := range result.Changes {
		if c.Severity == schema.SeverityBreaking {
			breaking = append(breaking, c)
		}
	}
	if len(breaking) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No breaking changes to scan for.")
		return nil
	}

	// Scan source files for each breaking change.
	var allHits []impact.Hit
	for _, c := range breaking {
		terms := impact.ExtractTerms(c)
		if len(terms) == 0 {
			continue
		}
		hits, err := impact.Scan(flagImpactScan, terms, changeLabel(c), string(c.Type))
		if err != nil {
			return fmt.Errorf("scanning %s: %w", flagImpactScan, err)
		}
		allHits = append(allHits, hits...)
	}

	return impact.Report(cmd.OutOrStdout(), allHits, flagImpactFormat)
}

// changeLabel returns a short human-readable label for a change used in reports.
func changeLabel(c schema.Change) string {
	if c.Method != "" && c.Path != "" {
		return fmt.Sprintf("%s %s (%s)", c.Method, c.Path, c.Type)
	}
	if c.Path != "" {
		return fmt.Sprintf("%s (%s)", c.Path, c.Type)
	}
	if c.Location != "" {
		return fmt.Sprintf("%s (%s)", c.Location, c.Type)
	}
	return strings.ReplaceAll(string(c.Type), "_", " ")
}
