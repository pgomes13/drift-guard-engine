package schema

// Severity represents how impactful a change is to API consumers.
type Severity string

const (
	SeverityBreaking    Severity = "breaking"
	SeverityNonBreaking Severity = "non-breaking"
	SeverityInfo        Severity = "info"
)

// ChangeType describes the nature of a diff between two schemas.
type ChangeType string

// Change represents a single detected difference between base and head schemas.
type Change struct {
	Type        ChangeType `json:"type"`
	Severity    Severity   `json:"severity"`
	Path        string     `json:"path"`     // e.g. "/users/{id}"
	Method      string     `json:"method"`   // e.g. "GET", empty if path-level
	Location    string     `json:"location"` // e.g. "request.body.email", "response.200.id"
	Description string     `json:"description"`
	Before      string     `json:"before,omitempty"`
	After       string     `json:"after,omitempty"`
}

// DiffResult holds the full output of a schema diff operation.
type DiffResult struct {
	BaseFile string   `json:"base_file"`
	HeadFile string   `json:"head_file"`
	Changes  []Change `json:"changes"`
	Summary  Summary  `json:"summary"`
}

// Summary aggregates change counts by severity.
type Summary struct {
	Total       int `json:"total"`
	Breaking    int `json:"breaking"`
	NonBreaking int `json:"non_breaking"`
	Info        int `json:"info"`
}
