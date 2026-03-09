// Package impact scans source files for references to breaking API changes.
package impact

// Hit represents a single source code reference to a breaking change.
type Hit struct {
	File       string `json:"file"`
	LineNum    int    `json:"line_num"`
	Line       string `json:"line"`
	ChangeType string `json:"change_type"`
	ChangePath string `json:"change_path"`
}
