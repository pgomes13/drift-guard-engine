package reporter_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DriftaBot/driftabot-engine/internal/reporter"
	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

// sampleResult builds a DiffResult with one breaking, one non-breaking, and one info change.
func sampleResult() schema.DiffResult {
	changes := []schema.Change{
		{
			Type:        schema.ChangeTypeMethodRemoved,
			Severity:    schema.SeverityBreaking,
			Path:        "/users/{id}",
			Method:      "DELETE",
			Description: "DELETE /users/{id} removed",
		},
		{
			Type:        schema.ChangeTypeEndpointAdded,
			Severity:    schema.SeverityNonBreaking,
			Path:        "/posts",
			Description: "/posts added",
		},
		{
			Type:        schema.ChangeTypeGQLFieldDeprecated,
			Severity:    schema.SeverityInfo,
			Path:        "User.email",
			Description: "User.email deprecated",
		},
	}
	return schema.DiffResult{
		BaseFile: "base.yaml",
		HeadFile: "head.yaml",
		Changes:  changes,
		Summary: schema.Summary{
			Total:       3,
			Breaking:    1,
			NonBreaking: 1,
			Info:        1,
		},
	}
}

func emptyResult() schema.DiffResult {
	return schema.DiffResult{
		BaseFile: "base.yaml",
		HeadFile: "head.yaml",
		Changes:  []schema.Change{},
		Summary:  schema.Summary{},
	}
}

// --------------------------------------------------------------------------
// JSON format
// --------------------------------------------------------------------------

func TestWrite_JSON_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := reporter.Write(&buf, sampleResult(), reporter.FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Errorf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
}

func TestWrite_JSON_ContainsChanges(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.FormatJSON)
	output := buf.String()
	if !strings.Contains(output, "breaking") {
		t.Error("expected JSON output to contain 'breaking'")
	}
	if !strings.Contains(output, "/users/{id}") {
		t.Error("expected JSON output to contain '/users/{id}'")
	}
}

func TestWrite_JSON_EmptyResult(t *testing.T) {
	var buf bytes.Buffer
	if err := reporter.Write(&buf, emptyResult(), reporter.FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
}

// --------------------------------------------------------------------------
// Text format
// --------------------------------------------------------------------------

func TestWrite_Text_ContainsSummary(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.FormatText)
	output := buf.String()
	if !strings.Contains(output, "Total:") {
		t.Error("expected text output to contain 'Total:'")
	}
	if !strings.Contains(output, "Breaking:") {
		t.Error("expected text output to contain 'Breaking:'")
	}
}

func TestWrite_Text_ContainsChanges(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.FormatText)
	output := buf.String()
	if !strings.Contains(output, "BREAKING") {
		t.Error("expected text output to contain 'BREAKING'")
	}
	if !strings.Contains(output, "/users/{id}") {
		t.Error("expected text output to contain '/users/{id}'")
	}
}

func TestWrite_Text_EmptyResult(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, emptyResult(), reporter.FormatText)
	output := buf.String()
	if !strings.Contains(output, "No changes detected.") {
		t.Errorf("expected 'No changes detected.' for empty result, got: %s", output)
	}
}

// --------------------------------------------------------------------------
// Markdown format
// --------------------------------------------------------------------------

func TestWrite_Markdown_ContainsTable(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.FormatMarkdown)
	output := buf.String()
	if !strings.Contains(output, "| Severity |") {
		t.Error("expected markdown output to contain table header '| Severity |'")
	}
}

func TestWrite_Markdown_ContainsSummaryLine(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.FormatMarkdown)
	output := buf.String()
	if !strings.Contains(output, "**Total:") {
		t.Error("expected markdown output to contain '**Total:'")
	}
}

func TestWrite_Markdown_ContainsBreakingRow(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.FormatMarkdown)
	output := buf.String()
	if !strings.Contains(output, "[BREAKING]") {
		t.Error("expected markdown output to contain '[BREAKING]'")
	}
}

func TestWrite_Markdown_EmptyResult(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, emptyResult(), reporter.FormatMarkdown)
	output := buf.String()
	if !strings.Contains(output, "No changes detected.") {
		t.Errorf("expected 'No changes detected.' for empty result, got: %s", output)
	}
}

// --------------------------------------------------------------------------
// GitHub format
// --------------------------------------------------------------------------

func TestWrite_GitHub_BreakingIsError(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.FormatGitHub)
	output := buf.String()
	if !strings.Contains(output, "::error") {
		t.Error("expected GitHub output to contain '::error' for breaking change")
	}
}

func TestWrite_GitHub_NonBreakingIsWarning(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.FormatGitHub)
	output := buf.String()
	if !strings.Contains(output, "::warning") {
		t.Error("expected GitHub output to contain '::warning' for non-breaking change")
	}
}

func TestWrite_GitHub_InfoIsNotice(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.FormatGitHub)
	output := buf.String()
	if !strings.Contains(output, "::notice") {
		t.Error("expected GitHub output to contain '::notice' for info change")
	}
}

func TestWrite_GitHub_SummaryErrorWhenBreaking(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.FormatGitHub)
	output := buf.String()
	if !strings.Contains(output, "API Contract Violation") {
		t.Error("expected GitHub output to contain 'API Contract Violation' summary error")
	}
}

func TestWrite_GitHub_EmptyResult_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, emptyResult(), reporter.FormatGitHub)
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty result, got: %s", buf.String())
	}
}

// --------------------------------------------------------------------------
// Default format (unrecognised → text)
// --------------------------------------------------------------------------

func TestWrite_DefaultFormat_FallsBackToText(t *testing.T) {
	var buf bytes.Buffer
	reporter.Write(&buf, sampleResult(), reporter.Format("unknown"))
	output := buf.String()
	if !strings.Contains(output, "Total:") {
		t.Error("expected default format to fall back to text output")
	}
}

// --------------------------------------------------------------------------
// HasBreakingChanges
// --------------------------------------------------------------------------

func TestHasBreakingChanges_True(t *testing.T) {
	if !reporter.HasBreakingChanges(sampleResult()) {
		t.Error("expected HasBreakingChanges=true for result with breaking changes")
	}
}

func TestHasBreakingChanges_False(t *testing.T) {
	if reporter.HasBreakingChanges(emptyResult()) {
		t.Error("expected HasBreakingChanges=false for empty result")
	}
}

func TestHasBreakingChanges_NonBreakingOnly(t *testing.T) {
	result := schema.DiffResult{
		Changes: []schema.Change{
			{Type: schema.ChangeTypeEndpointAdded, Severity: schema.SeverityNonBreaking},
		},
		Summary: schema.Summary{Total: 1, NonBreaking: 1},
	}
	if reporter.HasBreakingChanges(result) {
		t.Error("expected HasBreakingChanges=false when no breaking changes")
	}
}
