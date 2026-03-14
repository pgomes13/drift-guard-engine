// Package helpers provides internal diffing utilities for the OpenAPI differ.
package helpers

import (
	"fmt"

	"github.com/DriftAgent/api-drift-engine/pkg/schema"
)

func IndexEndpoints(s *schema.Schema) map[string]schema.Endpoint {
	m := make(map[string]schema.Endpoint, len(s.Endpoints))
	for _, e := range s.Endpoints {
		m[e.Path] = e
	}
	return m
}

func IndexOperations(endpoint schema.Endpoint) map[string]schema.Operation {
	m := make(map[string]schema.Operation, len(endpoint.Operations))
	for _, op := range endpoint.Operations {
		m[op.Method] = op
	}
	return m
}

func DiffOperations(path string, base, head schema.Endpoint) []schema.Change {
	var changes []schema.Change

	baseOps := IndexOperations(base)
	headOps := IndexOperations(head)

	for method := range baseOps {
		if _, exists := headOps[method]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeMethodRemoved,
				Path:        path,
				Method:      method,
				Description: fmt.Sprintf("%s %s was removed", method, path),
			})
		}
	}

	for method := range headOps {
		if _, exists := baseOps[method]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeMethodAdded,
				Path:        path,
				Method:      method,
				Description: fmt.Sprintf("%s %s was added", method, path),
			})
		}
	}

	for method, baseOp := range baseOps {
		headOp, exists := headOps[method]
		if !exists {
			continue
		}
		changes = append(changes, DiffParams(path, method, baseOp.Parameters, headOp.Parameters)...)
		changes = append(changes, DiffRequestBody(path, method, baseOp.RequestBody, headOp.RequestBody)...)
		changes = append(changes, DiffResponses(path, method, baseOp.Responses, headOp.Responses)...)
	}

	return changes
}
