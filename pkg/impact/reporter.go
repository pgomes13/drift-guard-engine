package impact

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Report writes the impact report to w in the requested format (text, json, markdown).
func Report(w io.Writer, hits []Hit, format string) error {
	if len(hits) == 0 {
		fmt.Fprintln(w, "No references found.")
		return nil
	}
	switch format {
	case "json":
		return reportJSON(w, hits)
	case "markdown":
		return reportMarkdown(w, hits)
	default:
		return reportText(w, hits)
	}
}

// groupByChange buckets hits by their (changeType, changePath) pair.
func groupByChange(hits []Hit) map[string][]Hit {
	m := make(map[string][]Hit)
	for _, h := range hits {
		key := h.ChangeType + "\x00" + h.ChangePath
		m[key] = append(m[key], h)
	}
	return m
}

func sortedKeys(m map[string][]Hit) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func reportText(w io.Writer, hits []Hit) error {
	groups := groupByChange(hits)
	for _, key := range sortedKeys(groups) {
		parts := strings.SplitN(key, "\x00", 2)
		fmt.Fprintf(w, "Breaking change: %s\n", parts[1])
		for _, h := range groups[key] {
			fmt.Fprintf(w, "  %s:%d\t%s\n", h.File, h.LineNum, h.Line)
		}
		fmt.Fprintln(w)
	}
	return nil
}

func reportMarkdown(w io.Writer, hits []Hit) error {
	groups := groupByChange(hits)
	for _, key := range sortedKeys(groups) {
		parts := strings.SplitN(key, "\x00", 2)
		fmt.Fprintf(w, "### Breaking change: %s\n\n", parts[1])
		fmt.Fprintln(w, "| File | Line | Code |")
		fmt.Fprintln(w, "|------|------|------|")
		for _, h := range groups[key] {
			fmt.Fprintf(w, "| `%s` | %d | `%s` |\n", h.File, h.LineNum, escapeMarkdown(h.Line))
		}
		fmt.Fprintln(w)
	}
	return nil
}

func reportJSON(w io.Writer, hits []Hit) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(hits)
}

func escapeMarkdown(s string) string {
	return strings.NewReplacer("|", "\\|", "`", "'").Replace(s)
}
