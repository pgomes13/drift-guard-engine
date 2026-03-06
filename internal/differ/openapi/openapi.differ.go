// Package openapi computes the structural diff between two OpenAPI schemas,
// producing a flat list of Change values for downstream classification.
package openapi

import (
	"fmt"

	"drift-guard-diff-engine/internal/differ/openapi/helpers"
	"drift-guard-diff-engine/pkg/schema"
)

// Diff computes all changes between base and head schemas.
func Diff(base, head *schema.Schema) []schema.Change {
	var changes []schema.Change

	baseEndpoints := helpers.IndexEndpoints(base)
	headEndpoints := helpers.IndexEndpoints(head)

	// Detect removed endpoints
	for path := range baseEndpoints {
		if _, exists := headEndpoints[path]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeEndpointRemoved,
				Path:        path,
				Description: fmt.Sprintf("Endpoint %s was removed", path),
			})
		}
	}

	// Detect added endpoints
	for path := range headEndpoints {
		if _, exists := baseEndpoints[path]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeEndpointAdded,
				Path:        path,
				Description: fmt.Sprintf("Endpoint %s was added", path),
			})
		}
	}

		// Diff operations on shared endpoints
	for path, baseEndpoint := range baseEndpoints {
		headEndpoint, exists := headEndpoints[path]
		if !exists {
			continue
		}
		changes = append(changes, helpers.DiffOperations(path, baseEndpoint, headEndpoint)...)
	}

	return changes
}
