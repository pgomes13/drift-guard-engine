// Package openapi computes the structural diff between two OpenAPI schemas,
// producing a flat list of Change values for downstream classification.
package openapi

import (
	"fmt"

	"drift-guard-diff-engine/pkg/schema"
)

// Diff computes all changes between base and head schemas.
func Diff(base, head *schema.Schema) []schema.Change {
	var changes []schema.Change

	baseEndpoints := indexEndpoints(base)
	headEndpoints := indexEndpoints(head)

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
		changes = append(changes, diffOperations(path, baseEndpoint, headEndpoint)...)
	}

	return changes
}

func indexEndpoints(s *schema.Schema) map[string]schema.Endpoint {
	m := make(map[string]schema.Endpoint, len(s.Endpoints))
	for _, e := range s.Endpoints {
		m[e.Path] = e
	}
	return m
}

func indexOperations(endpoint schema.Endpoint) map[string]schema.Operation {
	m := make(map[string]schema.Operation, len(endpoint.Operations))
	for _, op := range endpoint.Operations {
		m[op.Method] = op
	}
	return m
}

func diffOperations(path string, base, head schema.Endpoint) []schema.Change {
	var changes []schema.Change

	baseOps := indexOperations(base)
	headOps := indexOperations(head)

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
		changes = append(changes, diffParams(path, method, baseOp.Parameters, headOp.Parameters)...)
		changes = append(changes, diffRequestBody(path, method, baseOp.RequestBody, headOp.RequestBody)...)
		changes = append(changes, diffResponses(path, method, baseOp.Responses, headOp.Responses)...)
	}

	return changes
}

func indexParams(params []schema.Parameter) map[string]schema.Parameter {
	m := make(map[string]schema.Parameter, len(params))
	for _, p := range params {
		key := p.In + ":" + p.Name
		m[key] = p
	}
	return m
}

func diffParams(path, method string, base, head []schema.Parameter) []schema.Change {
	var changes []schema.Change

	baseParams := indexParams(base)
	headParams := indexParams(head)

	for key, bp := range baseParams {
		hp, exists := headParams[key]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeParamRemoved,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("param.%s.%s", bp.In, bp.Name),
				Description: fmt.Sprintf("Parameter '%s' (in %s) was removed from %s %s", bp.Name, bp.In, method, path),
			})
			continue
		}
		if bp.Type != hp.Type {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeParamTypeChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("param.%s.%s.type", bp.In, bp.Name),
				Description: fmt.Sprintf("Parameter '%s' type changed from '%s' to '%s' in %s %s", bp.Name, bp.Type, hp.Type, method, path),
				Before:      bp.Type,
				After:       hp.Type,
			})
		}
		if bp.Required != hp.Required {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeParamRequiredChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("param.%s.%s.required", bp.In, bp.Name),
				Description: fmt.Sprintf("Parameter '%s' required changed from %v to %v in %s %s", bp.Name, bp.Required, hp.Required, method, path),
				Before:      fmt.Sprintf("%v", bp.Required),
				After:       fmt.Sprintf("%v", hp.Required),
			})
		}
	}

	for key, hp := range headParams {
		if _, exists := baseParams[key]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeParamAdded,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("param.%s.%s", hp.In, hp.Name),
				Description: fmt.Sprintf("Parameter '%s' (in %s) was added to %s %s", hp.Name, hp.In, method, path),
			})
		}
	}

	return changes
}

func diffRequestBody(path, method string, base, head *schema.RequestBody) []schema.Change {
	if base == nil && head == nil {
		return nil
	}
	if base == nil && head != nil {
		return []schema.Change{{
			Type:        schema.ChangeTypeRequestBodyChanged,
			Path:        path,
			Method:      method,
			Location:    "request.body",
			Description: fmt.Sprintf("Request body was added to %s %s", method, path),
		}}
	}
	if base != nil && head == nil {
		return []schema.Change{{
			Type:        schema.ChangeTypeRequestBodyChanged,
			Path:        path,
			Method:      method,
			Location:    "request.body",
			Description: fmt.Sprintf("Request body was removed from %s %s", method, path),
		}}
	}
	return diffProperties(path, method, "request.body", base.Properties, head.Properties)
}

func diffResponses(path, method string, base, head []schema.Response) []schema.Change {
	var changes []schema.Change

	baseMap := make(map[string]schema.Response)
	for _, r := range base {
		baseMap[r.StatusCode] = r
	}
	headMap := make(map[string]schema.Response)
	for _, r := range head {
		headMap[r.StatusCode] = r
	}

	for code, br := range baseMap {
		hr, exists := headMap[code]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeResponseChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("response.%s", code),
				Description: fmt.Sprintf("Response %s was removed from %s %s", code, method, path),
			})
			continue
		}
		changes = append(changes, diffProperties(path, method, fmt.Sprintf("response.%s", code), br.Properties, hr.Properties)...)
	}

	for code := range headMap {
		if _, exists := baseMap[code]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeResponseChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("response.%s", code),
				Description: fmt.Sprintf("Response %s was added to %s %s", code, method, path),
			})
		}
	}

	return changes
}

func indexProperties(props []schema.Property) map[string]schema.Property {
	m := make(map[string]schema.Property, len(props))
	for _, p := range props {
		m[p.Name] = p
	}
	return m
}

func diffProperties(path, method, location string, base, head []schema.Property) []schema.Change {
	var changes []schema.Change

	baseProps := indexProperties(base)
	headProps := indexProperties(head)

	for name, bp := range baseProps {
		hp, exists := headProps[name]
		if !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeFieldRemoved,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("%s.%s", location, name),
				Description: fmt.Sprintf("Field '%s' was removed from %s in %s %s", name, location, method, path),
			})
			continue
		}
		if bp.Type != hp.Type {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeFieldTypeChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("%s.%s.type", location, name),
				Description: fmt.Sprintf("Field '%s' type changed from '%s' to '%s' in %s of %s %s", name, bp.Type, hp.Type, location, method, path),
				Before:      bp.Type,
				After:       hp.Type,
			})
		}
		if bp.Required != hp.Required {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeFieldRequiredChanged,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("%s.%s.required", location, name),
				Description: fmt.Sprintf("Field '%s' required changed from %v to %v in %s of %s %s", name, bp.Required, hp.Required, location, method, path),
				Before:      fmt.Sprintf("%v", bp.Required),
				After:       fmt.Sprintf("%v", hp.Required),
			})
		}
	}

	for name := range headProps {
		if _, exists := baseProps[name]; !exists {
			changes = append(changes, schema.Change{
				Type:        schema.ChangeTypeFieldAdded,
				Path:        path,
				Method:      method,
				Location:    fmt.Sprintf("%s.%s", location, name),
				Description: fmt.Sprintf("Field '%s' was added to %s in %s %s", name, location, method, path),
			})
		}
	}

	return changes
}
