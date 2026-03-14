package graphql_test

import (
	"testing"

	differgraphql "github.com/DriftAgent/api-drift-engine/internal/differ/graphql"
	parsergraphql "github.com/DriftAgent/api-drift-engine/internal/parser/graphql"
	"github.com/DriftAgent/api-drift-engine/pkg/schema"
)

const testdataDir = "../../testdata/"

// loadFixtures parses the base/head SDL fixtures used across all differ tests.
func loadFixtures(t *testing.T) (base, head *schema.GQLSchema) {
	t.Helper()
	var err error
	base, err = parsergraphql.Parse(testdataDir + "base.graphql")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	head, err = parsergraphql.Parse(testdataDir + "head.graphql")
	if err != nil {
		t.Fatalf("parse head: %v", err)
	}
	return
}

// findChange returns the first Change matching the given type and location substring.
func findChange(changes []schema.Change, ct schema.ChangeType, locSubstr string) *schema.Change {
	for i := range changes {
		if changes[i].Type == ct {
			if locSubstr == "" || contains(changes[i].Location, locSubstr) {
				return &changes[i]
			}
		}
	}
	return nil
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (s == sub || len(s) > 0 && searchSubstr(s, sub)))
}

func searchSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// --------------------------------------------------------------------------
// Field-level changes
// --------------------------------------------------------------------------

func TestDiffGQL_FieldRemoved(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// User.address was removed in head
	c := findChange(changes, schema.ChangeTypeGQLFieldRemoved, "User.address")
	if c == nil {
		t.Error("expected gql_field_removed for User.address")
	}
}

func TestDiffGQL_FieldAdded(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// User.createdAt was added in head
	c := findChange(changes, schema.ChangeTypeGQLFieldAdded, "User.createdAt")
	if c == nil {
		t.Error("expected gql_field_added for User.createdAt")
	}
}

// --------------------------------------------------------------------------
// Mutation-level changes
// --------------------------------------------------------------------------

func TestDiffGQL_MutationFieldRemoved(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// Mutation.deleteUser removed
	c := findChange(changes, schema.ChangeTypeGQLFieldRemoved, "Mutation.deleteUser")
	if c == nil {
		t.Error("expected gql_field_removed for Mutation.deleteUser")
	}
}

// --------------------------------------------------------------------------
// Argument changes
// --------------------------------------------------------------------------

func TestDiffGQL_ArgAdded_Optional(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// Query.user gains optional arg includeDeleted: Boolean (non-breaking)
	c := findChange(changes, schema.ChangeTypeGQLArgAdded, "Query.user(arg:includeDeleted)")
	if c == nil {
		t.Fatal("expected gql_arg_added for Query.user(arg:includeDeleted)")
	}
	if c.After != "Boolean" {
		t.Errorf("expected arg type 'Boolean', got '%s'", c.After)
	}
}

func TestDiffGQL_ArgAdded_OnSearch(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// Query.search gains optional arg type: String
	c := findChange(changes, schema.ChangeTypeGQLArgAdded, "Query.search(arg:type)")
	if c == nil {
		t.Error("expected gql_arg_added for Query.search(arg:type)")
	}
}

// --------------------------------------------------------------------------
// Enum changes
// --------------------------------------------------------------------------

func TestDiffGQL_EnumValueRemoved(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// UserRole.EDITOR removed
	c := findChange(changes, schema.ChangeTypeGQLEnumValueRemoved, "UserRole.EDITOR")
	if c == nil {
		t.Fatal("expected gql_enum_value_removed for UserRole.EDITOR")
	}
	if c.Before != "EDITOR" {
		t.Errorf("expected Before='EDITOR', got '%s'", c.Before)
	}
}

func TestDiffGQL_EnumValueNotFalsePositive(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// ADMIN and VIEWER still exist — must NOT be reported as removed
	for _, c := range changes {
		if c.Type == schema.ChangeTypeGQLEnumValueRemoved {
			if contains(c.Location, "UserRole.ADMIN") || contains(c.Location, "UserRole.VIEWER") {
				t.Errorf("false positive: %s should not be reported removed", c.Location)
			}
		}
	}
}

// --------------------------------------------------------------------------
// Input field changes
// --------------------------------------------------------------------------

func TestDiffGQL_InputFieldTypeChanged_MadeRequired(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// CreateUserInput.role changed from UserRole (nullable) to UserRole! (non-null) — breaking
	c := findChange(changes, schema.ChangeTypeGQLInputFieldTypeChanged, "CreateUserInput.role")
	if c == nil {
		t.Fatal("expected gql_input_field_type_changed for CreateUserInput.role")
	}
	if c.Before != "UserRole" {
		t.Errorf("expected Before='UserRole', got '%s'", c.Before)
	}
	if c.After != "UserRole!" {
		t.Errorf("expected After='UserRole!', got '%s'", c.After)
	}
}

func TestDiffGQL_InputFieldAdded_Optional(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// UpdateUserInput.notes added (nullable String — non-breaking)
	c := findChange(changes, schema.ChangeTypeGQLInputFieldAdded, "UpdateUserInput.notes")
	if c == nil {
		t.Error("expected gql_input_field_added for UpdateUserInput.notes")
	}
}

// --------------------------------------------------------------------------
// Union changes
// --------------------------------------------------------------------------

func TestDiffGQL_UnionMemberAdded(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// SearchResult gains Post member
	c := findChange(changes, schema.ChangeTypeGQLUnionMemberAdded, "Post")
	if c == nil {
		t.Error("expected gql_union_member_added for Post in SearchResult")
	}
}

// --------------------------------------------------------------------------
// Type-level changes
// --------------------------------------------------------------------------

func TestDiffGQL_TypeAdded(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// Post type is new in head
	c := findChange(changes, schema.ChangeTypeGQLTypeAdded, "Post")
	if c == nil {
		t.Fatal("expected gql_type_added for Post")
	}
	if c.After != string(schema.GQLTypeKindObject) {
		t.Errorf("expected After='OBJECT', got '%s'", c.After)
	}
}

func TestDiffGQL_NoFalsePositive_UnchangedType(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	// Node interface is unchanged — should not appear in changes
	for _, c := range changes {
		if contains(c.Location, "Node") {
			t.Errorf("unexpected change involving 'Node': %s %s", c.Type, c.Location)
		}
	}
}

// --------------------------------------------------------------------------
// Change count sanity
// --------------------------------------------------------------------------

func TestDiffGQL_TotalChanges(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgraphql.Diff(base, head)

	if len(changes) == 0 {
		t.Error("expected at least one change between base and head fixtures")
	}

	// Count breaking-candidate types to ensure nothing obvious is silently dropped
	breaking := map[schema.ChangeType]int{}
	for _, c := range changes {
		breaking[c.Type]++
	}

	// We know these must be present from the fixture design
	required := []schema.ChangeType{
		schema.ChangeTypeGQLFieldRemoved,          // User.address, Mutation.deleteUser
		schema.ChangeTypeGQLEnumValueRemoved,      // UserRole.EDITOR
		schema.ChangeTypeGQLInputFieldTypeChanged, // CreateUserInput.role
		schema.ChangeTypeGQLFieldAdded,            // User.createdAt, Address.postcode
		schema.ChangeTypeGQLArgAdded,              // Query.user includeDeleted
		schema.ChangeTypeGQLUnionMemberAdded,      // SearchResult | Post
		schema.ChangeTypeGQLTypeAdded,             // Post
	}

	for _, ct := range required {
		if breaking[ct] == 0 {
			t.Errorf("expected at least one change of type %s", ct)
		}
	}
}

func TestDiffGQL_IdenticalSchemas_NoChanges(t *testing.T) {
	base, err := parsergraphql.Parse(testdataDir + "base.graphql")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	changes := differgraphql.Diff(base, base)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for identical schemas, got %d: %v", len(changes), changes)
	}
}
