package openapi

import "github.com/DriftAgent/api-drift-engine/pkg/schema"

// Severity returns the severity for OpenAPI change types.
// The second return value is false if the change type is not an OpenAPI type.
func Severity(c schema.Change) (schema.Severity, bool) {
	switch c.Type {

	// Endpoint / method removals — always breaking
	case schema.ChangeTypeEndpointRemoved,
		schema.ChangeTypeMethodRemoved:
		return schema.SeverityBreaking, true

	// Endpoint / method additions — non-breaking
	case schema.ChangeTypeEndpointAdded,
		schema.ChangeTypeMethodAdded:
		return schema.SeverityNonBreaking, true

	// Parameter removed — breaking
	case schema.ChangeTypeParamRemoved:
		return schema.SeverityBreaking, true

	// Parameter added — conservatively non-breaking (optional by default)
	case schema.ChangeTypeParamAdded:
		return schema.SeverityNonBreaking, true

	// Parameter type change — always breaking
	case schema.ChangeTypeParamTypeChanged:
		return schema.SeverityBreaking, true

	// Parameter required changed: false → true is breaking; true → false is non-breaking
	case schema.ChangeTypeParamRequiredChanged:
		if c.Before == "false" && c.After == "true" {
			return schema.SeverityBreaking, true
		}
		return schema.SeverityNonBreaking, true

	// Request body removed is breaking; added is non-breaking
	case schema.ChangeTypeRequestBodyChanged:
		if c.After == "" {
			return schema.SeverityBreaking, true
		}
		return schema.SeverityNonBreaking, true

	// Response code removed is breaking; added is non-breaking
	case schema.ChangeTypeResponseChanged:
		if c.After == "" {
			return schema.SeverityBreaking, true
		}
		return schema.SeverityNonBreaking, true

	// Field removed — breaking
	case schema.ChangeTypeFieldRemoved:
		return schema.SeverityBreaking, true

	// Field added — non-breaking
	case schema.ChangeTypeFieldAdded:
		return schema.SeverityNonBreaking, true

	// Field type changed — breaking
	case schema.ChangeTypeFieldTypeChanged:
		return schema.SeverityBreaking, true

	// Field required changed: false → true is breaking; true → false is non-breaking
	case schema.ChangeTypeFieldRequiredChanged:
		if c.Before == "false" && c.After == "true" {
			return schema.SeverityBreaking, true
		}
		return schema.SeverityNonBreaking, true

	default:
		return "", false
	}
}
