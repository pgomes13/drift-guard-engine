package openapi_test

import (
	"testing"

	differopenapi "github.com/DriftAgent/api-drift-engine/internal/differ/openapi"
	parseropenapi "github.com/DriftAgent/api-drift-engine/internal/parser/openapi"
	"github.com/DriftAgent/api-drift-engine/pkg/schema"
)

const testdataDir = "../../testdata/"

// loadFixtures parses the base/head OpenAPI fixtures used across all differ tests.
func loadFixtures(t *testing.T) (base, head *schema.Schema) {
	t.Helper()
	var err error
	base, err = parseropenapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	head, err = parseropenapi.Parse(testdataDir + "head.yaml")
	if err != nil {
		t.Fatalf("parse head: %v", err)
	}
	return
}

// findChange returns the first Change matching the given type and optional
// path/location substrings (empty string matches any).
func findChange(changes []schema.Change, ct schema.ChangeType, pathSubstr, locSubstr string) *schema.Change {
	for i := range changes {
		c := &changes[i]
		if c.Type != ct {
			continue
		}
		if pathSubstr != "" && !contains(c.Path, pathSubstr) {
			continue
		}
		if locSubstr != "" && !contains(c.Location, locSubstr) {
			continue
		}
		return c
	}
	return nil
}

func contains(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// --------------------------------------------------------------------------
// Method-level changes
// --------------------------------------------------------------------------

func TestDiff_MethodRemoved(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differopenapi.Diff(base, head)

	// DELETE /users/{id} was removed in head
	c := findChange(changes, schema.ChangeTypeMethodRemoved, "/users/{id}", "")
	if c == nil {
		t.Fatal("expected method_removed for DELETE /users/{id}")
	}
	if c.Method != "DELETE" {
		t.Errorf("expected Method='DELETE', got '%s'", c.Method)
	}
}

// --------------------------------------------------------------------------
// Endpoint-level changes
// --------------------------------------------------------------------------

func TestDiff_EndpointAdded(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differopenapi.Diff(base, head)

	// /posts is a new endpoint in head
	c := findChange(changes, schema.ChangeTypeEndpointAdded, "/posts", "")
	if c == nil {
		t.Error("expected endpoint_added for /posts")
	}
}

// --------------------------------------------------------------------------
// Parameter changes
// --------------------------------------------------------------------------

func TestDiff_ParamTypeChanged(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differopenapi.Diff(base, head)

	// GET /users/{id} param 'id' changed from string to integer
	c := findChange(changes, schema.ChangeTypeParamTypeChanged, "/users/{id}", "path.id")
	if c == nil {
		t.Fatal("expected param_type_changed for id on GET /users/{id}")
	}
	if c.Before != "string" {
		t.Errorf("expected Before='string', got '%s'", c.Before)
	}
	if c.After != "integer" {
		t.Errorf("expected After='integer', got '%s'", c.After)
	}
}

// --------------------------------------------------------------------------
// Response field changes
// --------------------------------------------------------------------------

func TestDiff_FieldRemoved_InResponse(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differopenapi.Diff(base, head)

	// GET /users response 200: 'total' field removed
	c := findChange(changes, schema.ChangeTypeFieldRemoved, "/users", "total")
	if c == nil {
		t.Error("expected field_removed for 'total' in GET /users response")
	}
}

// --------------------------------------------------------------------------
// Request body field changes
// --------------------------------------------------------------------------

func TestDiff_FieldAdded_InRequestBody(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differopenapi.Diff(base, head)

	// POST /users request body: 'role' field added (non-breaking)
	c := findChange(changes, schema.ChangeTypeFieldAdded, "/users", "role")
	if c == nil {
		t.Error("expected field_added for 'role' in POST /users request body")
	}
}

// --------------------------------------------------------------------------
// No false positives
// --------------------------------------------------------------------------

func TestDiff_NoFalsePositive_UnchangedEndpoint(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differopenapi.Diff(base, head)

	// GET /users itself is not removed — only fields inside it changed
	for _, c := range changes {
		if c.Type == schema.ChangeTypeEndpointRemoved && c.Path == "/users" {
			t.Errorf("false positive: /users should not be reported as removed")
		}
	}
}

// --------------------------------------------------------------------------
// Change count sanity
// --------------------------------------------------------------------------

func TestDiff_TotalChanges(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differopenapi.Diff(base, head)

	if len(changes) == 0 {
		t.Error("expected at least one change between base and head fixtures")
	}

	byType := map[schema.ChangeType]int{}
	for _, c := range changes {
		byType[c.Type]++
	}

	required := []schema.ChangeType{
		schema.ChangeTypeMethodRemoved,    // DELETE /users/{id}
		schema.ChangeTypeEndpointAdded,    // /posts
		schema.ChangeTypeParamTypeChanged, // id: string → integer
		schema.ChangeTypeFieldRemoved,     // total removed from response
		schema.ChangeTypeFieldAdded,       // role added to request body
	}

	for _, ct := range required {
		if byType[ct] == 0 {
			t.Errorf("expected at least one change of type %s", ct)
		}
	}
}

func TestDiff_IdenticalSchemas_NoChanges(t *testing.T) {
	base, err := parseropenapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	changes := differopenapi.Diff(base, base)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for identical schemas, got %d: %v", len(changes), changes)
	}
}
