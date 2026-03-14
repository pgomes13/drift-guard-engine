package graphql

import (
	"strings"

	"github.com/pgomes13/api-drift-engine/pkg/schema"
)

// Severity returns the severity for GraphQL change types.
// The second return value is false if the change type is not a GraphQL type.
func Severity(c schema.Change) (schema.Severity, bool) {
	switch c.Type {

	// Type removed — always breaking
	case schema.ChangeTypeGQLTypeRemoved:
		return schema.SeverityBreaking, true

	// Type added — non-breaking
	case schema.ChangeTypeGQLTypeAdded:
		return schema.SeverityNonBreaking, true

	// Type kind changed (e.g. Object → Interface) — always breaking
	case schema.ChangeTypeGQLTypeKindChanged:
		return schema.SeverityBreaking, true

	// Output field removed — breaking
	case schema.ChangeTypeGQLFieldRemoved:
		return schema.SeverityBreaking, true

	// Output field added — non-breaking
	case schema.ChangeTypeGQLFieldAdded:
		return schema.SeverityNonBreaking, true

	// Output field deprecated — informational (not yet removed)
	case schema.ChangeTypeGQLFieldDeprecated:
		return schema.SeverityInfo, true

	// Output field type changed — apply nullability rules:
	//   String! → String  : breaking (consumers relied on non-null guarantee)
	//   String  → String! : non-breaking (consumers already handled null)
	//   Any other type change: breaking
	case schema.ChangeTypeGQLFieldTypeChanged:
		if isNullabilityRelaxed(c.Before, c.After) {
			return schema.SeverityBreaking, true
		}
		if isNullabilityTightened(c.Before, c.After) {
			return schema.SeverityNonBreaking, true
		}
		return schema.SeverityBreaking, true

	// Argument removed from a field — breaking
	case schema.ChangeTypeGQLArgRemoved:
		return schema.SeverityBreaking, true

	// Argument added:
	//   required (Type!) with no default → breaking
	//   optional (Type)                  → non-breaking
	case schema.ChangeTypeGQLArgAdded:
		if isRequiredType(c.After) {
			return schema.SeverityBreaking, true
		}
		return schema.SeverityNonBreaking, true

	// Argument type changed (input direction):
	//   String  → String! : breaking (callers not providing it will now fail)
	//   String! → String  : non-breaking
	//   Any other type change: breaking
	case schema.ChangeTypeGQLArgTypeChanged:
		if isNullabilityTightened(c.Before, c.After) {
			return schema.SeverityBreaking, true
		}
		if isNullabilityRelaxed(c.Before, c.After) {
			return schema.SeverityNonBreaking, true
		}
		return schema.SeverityBreaking, true

	// Argument default changed — informational
	case schema.ChangeTypeGQLArgDefaultChanged:
		return schema.SeverityInfo, true

	// Enum value removed — breaking (consumers may send/receive that value)
	case schema.ChangeTypeGQLEnumValueRemoved:
		return schema.SeverityBreaking, true

	// Enum value added — non-breaking for existing consumers
	case schema.ChangeTypeGQLEnumValueAdded:
		return schema.SeverityNonBreaking, true

	// Union member removed — breaking
	case schema.ChangeTypeGQLUnionMemberRemoved:
		return schema.SeverityBreaking, true

	// Union member added — non-breaking
	case schema.ChangeTypeGQLUnionMemberAdded:
		return schema.SeverityNonBreaking, true

	// Input field removed — breaking
	case schema.ChangeTypeGQLInputFieldRemoved:
		return schema.SeverityBreaking, true

	// Input field added:
	//   required (Type!) → breaking; optional (Type) → non-breaking
	case schema.ChangeTypeGQLInputFieldAdded:
		if isRequiredType(c.After) {
			return schema.SeverityBreaking, true
		}
		return schema.SeverityNonBreaking, true

	// Input field type changed (input direction) — same rules as arg type
	case schema.ChangeTypeGQLInputFieldTypeChanged:
		if isNullabilityTightened(c.Before, c.After) {
			return schema.SeverityBreaking, true
		}
		if isNullabilityRelaxed(c.Before, c.After) {
			return schema.SeverityNonBreaking, true
		}
		return schema.SeverityBreaking, true

	// Interface removed from an object type — breaking
	case schema.ChangeTypeGQLInterfaceRemoved:
		return schema.SeverityBreaking, true

	// Interface added to an object type — non-breaking
	case schema.ChangeTypeGQLInterfaceAdded:
		return schema.SeverityNonBreaking, true

	default:
		return "", false
	}
}

// isRequiredType returns true when a GraphQL type string is non-nullable
// (ends with "!"). e.g. "String!" → true, "String" → false.
func isRequiredType(t string) bool {
	return strings.HasSuffix(strings.TrimSpace(t), "!")
}

// isNullabilityRelaxed returns true when a non-null type becomes nullable.
// e.g. "String!" → "String" (output field loses its non-null guarantee).
func isNullabilityRelaxed(before, after string) bool {
	return isRequiredType(before) && !isRequiredType(after) &&
		strings.TrimSuffix(strings.TrimSpace(before), "!") == strings.TrimSpace(after)
}

// isNullabilityTightened returns true when a nullable type becomes non-null.
// e.g. "String" → "String!" (type is now guaranteed non-null).
func isNullabilityTightened(before, after string) bool {
	return !isRequiredType(before) && isRequiredType(after) &&
		strings.TrimSpace(before) == strings.TrimSuffix(strings.TrimSpace(after), "!")
}
