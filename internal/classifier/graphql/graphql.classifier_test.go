package graphql_test

import (
	"testing"

	"github.com/DriftAgent/api-drift-engine/internal/classifier"
	differgraphql "github.com/DriftAgent/api-drift-engine/internal/differ/graphql"
	parsergraphql "github.com/DriftAgent/api-drift-engine/internal/parser/graphql"
	"github.com/DriftAgent/api-drift-engine/pkg/schema"
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
	// Type-level
	{
		name:     "type removed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLTypeRemoved, Before: "OBJECT"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "type added is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLTypeAdded, After: "OBJECT"},
		expected: schema.SeverityNonBreaking,
	},
	{
		name:     "type kind changed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLTypeKindChanged, Before: "OBJECT", After: "INTERFACE"},
		expected: schema.SeverityBreaking,
	},

	// Output fields
	{
		name:     "field removed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLFieldRemoved},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "field added is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLFieldAdded},
		expected: schema.SeverityNonBreaking,
	},
	{
		name:     "field deprecated is info",
		change:   schema.Change{Type: schema.ChangeTypeGQLFieldDeprecated},
		expected: schema.SeverityInfo,
	},
	{
		name:     "field type changed (unrelated types) is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLFieldTypeChanged, Before: "String", After: "Int"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "field type String! to String is breaking (nullability relaxed on output)",
		change:   schema.Change{Type: schema.ChangeTypeGQLFieldTypeChanged, Before: "String!", After: "String"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "field type String to String! is non-breaking (nullability tightened on output)",
		change:   schema.Change{Type: schema.ChangeTypeGQLFieldTypeChanged, Before: "String", After: "String!"},
		expected: schema.SeverityNonBreaking,
	},

	// Arguments
	{
		name:     "arg removed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLArgRemoved, Before: "String!"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "arg added optional is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLArgAdded, After: "Boolean"},
		expected: schema.SeverityNonBreaking,
	},
	{
		name:     "arg added required is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLArgAdded, After: "String!"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "arg type String to String! is breaking (input direction)",
		change:   schema.Change{Type: schema.ChangeTypeGQLArgTypeChanged, Before: "String", After: "String!"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "arg type String! to String is non-breaking (input direction)",
		change:   schema.Change{Type: schema.ChangeTypeGQLArgTypeChanged, Before: "String!", After: "String"},
		expected: schema.SeverityNonBreaking,
	},
	{
		name:     "arg type changed to unrelated type is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLArgTypeChanged, Before: "String", After: "Int"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "arg default changed is info",
		change:   schema.Change{Type: schema.ChangeTypeGQLArgDefaultChanged, Before: "10", After: "20"},
		expected: schema.SeverityInfo,
	},

	// Enums
	{
		name:     "enum value removed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLEnumValueRemoved, Before: "ADMIN"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "enum value added is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLEnumValueAdded, After: "SUPERADMIN"},
		expected: schema.SeverityNonBreaking,
	},

	// Union
	{
		name:     "union member removed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLUnionMemberRemoved, Before: "User"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "union member added is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLUnionMemberAdded, After: "Post"},
		expected: schema.SeverityNonBreaking,
	},

	// Input fields
	{
		name:     "input field removed is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLInputFieldRemoved, Before: "String!"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "input field added optional is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLInputFieldAdded, After: "String"},
		expected: schema.SeverityNonBreaking,
	},
	{
		name:     "input field added required is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLInputFieldAdded, After: "String!"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "input field type nullable to non-null is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLInputFieldTypeChanged, Before: "UserRole", After: "UserRole!"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "input field type non-null to nullable is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLInputFieldTypeChanged, Before: "UserRole!", After: "UserRole"},
		expected: schema.SeverityNonBreaking,
	},
	{
		name:     "input field type changed to unrelated type is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLInputFieldTypeChanged, Before: "String", After: "Int"},
		expected: schema.SeverityBreaking,
	},

	// Interfaces
	{
		name:     "interface removed from type is breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLInterfaceRemoved, Before: "Node"},
		expected: schema.SeverityBreaking,
	},
	{
		name:     "interface added to type is non-breaking",
		change:   schema.Change{Type: schema.ChangeTypeGQLInterfaceAdded, After: "Node"},
		expected: schema.SeverityNonBreaking,
	},
}

func TestClassify_GraphQL_SeverityRules(t *testing.T) {
	for _, tc := range severityCases {
		tc := tc
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
// Summary counts
// --------------------------------------------------------------------------

func TestClassify_GraphQL_SummaryCounts(t *testing.T) {
	changes := []schema.Change{
		{Type: schema.ChangeTypeGQLFieldRemoved},                  // breaking
		{Type: schema.ChangeTypeGQLEnumValueRemoved},              // breaking
		{Type: schema.ChangeTypeGQLFieldAdded},                    // non-breaking
		{Type: schema.ChangeTypeGQLTypeAdded, After: "OBJECT"},    // non-breaking
		{Type: schema.ChangeTypeGQLFieldDeprecated},               // info
	}

	result := classifier.Classify("base", "head", changes)

	if result.Summary.Total != 5 {
		t.Errorf("total = %d, want 5", result.Summary.Total)
	}
	if result.Summary.Breaking != 2 {
		t.Errorf("breaking = %d, want 2", result.Summary.Breaking)
	}
	if result.Summary.NonBreaking != 2 {
		t.Errorf("non_breaking = %d, want 2", result.Summary.NonBreaking)
	}
	if result.Summary.Info != 1 {
		t.Errorf("info = %d, want 1", result.Summary.Info)
	}
}

// --------------------------------------------------------------------------
// Integration: fixture files → full pipeline
// --------------------------------------------------------------------------

func TestClassify_GraphQL_FixtureBreakingChanges(t *testing.T) {
	base, err := parsergraphql.Parse(testdataDir + "base.graphql")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	head, err := parsergraphql.Parse(testdataDir + "head.graphql")
	if err != nil {
		t.Fatalf("parse head: %v", err)
	}

	changes := differgraphql.Diff(base, head)
	result := classifier.Classify("base.graphql", "head.graphql", changes)

	if result.Summary.Breaking == 0 {
		t.Error("expected at least one breaking change from fixture diff")
	}

	// Known breaking changes from fixture design:
	// 1. User.address removed
	// 2. Mutation.deleteUser removed
	// 3. UserRole.EDITOR removed
	// 4. CreateUserInput.role made required (UserRole → UserRole!)
	if result.Summary.Breaking < 4 {
		t.Errorf("expected at least 4 breaking changes, got %d", result.Summary.Breaking)
	}
}

func TestClassify_GraphQL_FixtureNonBreakingChanges(t *testing.T) {
	base, err := parsergraphql.Parse(testdataDir + "base.graphql")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	head, err := parsergraphql.Parse(testdataDir + "head.graphql")
	if err != nil {
		t.Fatalf("parse head: %v", err)
	}

	changes := differgraphql.Diff(base, head)
	result := classifier.Classify("base.graphql", "head.graphql", changes)

	if result.Summary.NonBreaking == 0 {
		t.Error("expected at least one non-breaking change from fixture diff")
	}
}

func TestClassify_GraphQL_IdenticalSchemas_NoChanges(t *testing.T) {
	base, err := parsergraphql.Parse(testdataDir + "base.graphql")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	result := classifier.Classify("base.graphql", "base.graphql", differgraphql.Diff(base, base))

	if result.Summary.Total != 0 {
		t.Errorf("expected 0 changes for identical schemas, got %d", result.Summary.Total)
	}
}
