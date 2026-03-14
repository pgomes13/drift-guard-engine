package openapi_test

import (
	"testing"

	"github.com/DriftaBot/driftabot-engine/internal/classifier"
	differopenapi "github.com/DriftaBot/driftabot-engine/internal/differ/openapi"
	parseropenapi "github.com/DriftaBot/driftabot-engine/internal/parser/openapi"
	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

const testdataDir = "../../testdata/"

// --------------------------------------------------------------------------
// Severity rules — table-driven
// --------------------------------------------------------------------------

type severityCase struct {
	name     string
	change   schema.Change
	expected schema.Severity
}

var severityCases = []severityCase{
	// Endpoint / method
	{name: "endpoint removed is breaking", change: schema.Change{Type: schema.ChangeTypeEndpointRemoved}, expected: schema.SeverityBreaking},
	{name: "endpoint added is non-breaking", change: schema.Change{Type: schema.ChangeTypeEndpointAdded}, expected: schema.SeverityNonBreaking},
	{name: "method removed is breaking", change: schema.Change{Type: schema.ChangeTypeMethodRemoved}, expected: schema.SeverityBreaking},
	{name: "method added is non-breaking", change: schema.Change{Type: schema.ChangeTypeMethodAdded}, expected: schema.SeverityNonBreaking},

	// Parameters
	{name: "param removed is breaking", change: schema.Change{Type: schema.ChangeTypeParamRemoved}, expected: schema.SeverityBreaking},
	{name: "param added is non-breaking", change: schema.Change{Type: schema.ChangeTypeParamAdded}, expected: schema.SeverityNonBreaking},
	{name: "param type changed is breaking", change: schema.Change{Type: schema.ChangeTypeParamTypeChanged, Before: "string", After: "integer"}, expected: schema.SeverityBreaking},
	{name: "param required false→true is breaking", change: schema.Change{Type: schema.ChangeTypeParamRequiredChanged, Before: "false", After: "true"}, expected: schema.SeverityBreaking},
	{name: "param required true→false is non-breaking", change: schema.Change{Type: schema.ChangeTypeParamRequiredChanged, Before: "true", After: "false"}, expected: schema.SeverityNonBreaking},

	// Request body
	{name: "request body removed is breaking", change: schema.Change{Type: schema.ChangeTypeRequestBodyChanged, Before: "present", After: ""}, expected: schema.SeverityBreaking},
	{name: "request body added is non-breaking", change: schema.Change{Type: schema.ChangeTypeRequestBodyChanged, Before: "", After: "present"}, expected: schema.SeverityNonBreaking},

	// Responses
	{name: "response removed is breaking", change: schema.Change{Type: schema.ChangeTypeResponseChanged, Before: "present", After: ""}, expected: schema.SeverityBreaking},
	{name: "response added is non-breaking", change: schema.Change{Type: schema.ChangeTypeResponseChanged, Before: "", After: "present"}, expected: schema.SeverityNonBreaking},

	// Fields
	{name: "field removed is breaking", change: schema.Change{Type: schema.ChangeTypeFieldRemoved}, expected: schema.SeverityBreaking},
	{name: "field added is non-breaking", change: schema.Change{Type: schema.ChangeTypeFieldAdded}, expected: schema.SeverityNonBreaking},
	{name: "field type changed is breaking", change: schema.Change{Type: schema.ChangeTypeFieldTypeChanged, Before: "string", After: "integer"}, expected: schema.SeverityBreaking},
	{name: "field required false→true is breaking", change: schema.Change{Type: schema.ChangeTypeFieldRequiredChanged, Before: "false", After: "true"}, expected: schema.SeverityBreaking},
	{name: "field required true→false is non-breaking", change: schema.Change{Type: schema.ChangeTypeFieldRequiredChanged, Before: "true", After: "false"}, expected: schema.SeverityNonBreaking},
}

func TestClassify_OpenAPI_SeverityRules(t *testing.T) {
	for _, tc := range severityCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.Classify("base", "head", []schema.Change{tc.change})
			if len(result.Changes) != 1 {
				t.Fatalf("expected 1 classified change, got %d", len(result.Changes))
			}
			got := result.Changes[0].Severity
			if got != tc.expected {
				t.Errorf("severity = %s, want %s", got, tc.expected)
			}
		})
	}
}

// --------------------------------------------------------------------------
// Integration: fixture files → full pipeline
// --------------------------------------------------------------------------

func TestClassify_OpenAPI_FixtureBreakingChanges(t *testing.T) {
	base, err := parseropenapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	head, err := parseropenapi.Parse(testdataDir + "head.yaml")
	if err != nil {
		t.Fatalf("parse head: %v", err)
	}

	changes := differopenapi.Diff(base, head)
	result := classifier.Classify("base.yaml", "head.yaml", changes)

	if result.Summary.Breaking == 0 {
		t.Error("expected at least one breaking change from fixture diff")
	}

	// Known breaking changes from fixture design:
	// 1. DELETE /users/{id} removed (method_removed)
	// 2. GET /users response 'total' field removed (field_removed)
	// 3. GET /users/{id} param 'id' type string→integer (param_type_changed)
	if result.Summary.Breaking < 3 {
		t.Errorf("expected at least 3 breaking changes, got %d", result.Summary.Breaking)
	}
}

func TestClassify_OpenAPI_FixtureNonBreakingChanges(t *testing.T) {
	base, err := parseropenapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	head, err := parseropenapi.Parse(testdataDir + "head.yaml")
	if err != nil {
		t.Fatalf("parse head: %v", err)
	}

	changes := differopenapi.Diff(base, head)
	result := classifier.Classify("base.yaml", "head.yaml", changes)

	if result.Summary.NonBreaking == 0 {
		t.Error("expected at least one non-breaking change from fixture diff")
	}
}

func TestClassify_OpenAPI_IdenticalSchemas_NoChanges(t *testing.T) {
	base, err := parseropenapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	result := classifier.Classify("base.yaml", "base.yaml", differopenapi.Diff(base, base))

	if result.Summary.Total != 0 {
		t.Errorf("expected 0 changes for identical schemas, got %d", result.Summary.Total)
	}
}
