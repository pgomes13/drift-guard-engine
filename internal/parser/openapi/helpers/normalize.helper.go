package helpers

import (
	"fmt"

	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

// Normalize converts a raw OpenAPI document into a normalized Schema.
func Normalize(raw *RawOpenAPI) *schema.Schema {
	s := &schema.Schema{
		Title:   raw.Info.Title,
		Version: raw.Info.Version,
	}

	for path, item := range raw.Paths {
		endpoint := schema.Endpoint{Path: path}
		ops := map[string]*RawOperation{
			"GET":     item.Get,
			"POST":    item.Post,
			"PUT":     item.Put,
			"PATCH":   item.Patch,
			"DELETE":  item.Delete,
			"HEAD":    item.Head,
			"OPTIONS": item.Options,
		}
		for method, rawOp := range ops {
			if rawOp == nil {
				continue
			}
			op := schema.Operation{
				Method:      method,
				OperationID: rawOp.OperationID,
				Parameters:  NormalizeParams(rawOp.Parameters),
				RequestBody: NormalizeRequestBody(rawOp.RequestBody),
				Responses:   NormalizeResponses(rawOp.Responses),
			}
			endpoint.Operations = append(endpoint.Operations, op)
		}
		s.Endpoints = append(s.Endpoints, endpoint)
	}

	return s
}

// NormalizeParams converts raw parameters into normalized Parameter values.
func NormalizeParams(raw []RawParameter) []schema.Parameter {
	params := make([]schema.Parameter, 0, len(raw))
	for _, p := range raw {
		params = append(params, schema.Parameter{
			Name:     p.Name,
			In:       p.In,
			Required: p.Required,
			Type:     p.Schema.Type,
			Format:   p.Schema.Format,
			Ref:      p.Ref,
		})
	}
	return params
}

// NormalizeRequestBody converts a raw request body into a normalized RequestBody.
func NormalizeRequestBody(raw *RawRequestBody) *schema.RequestBody {
	if raw == nil {
		return nil
	}
	rb := &schema.RequestBody{Required: raw.Required}
	for _, media := range raw.Content {
		rb.Properties = NormalizeProperties(media.Schema)
		break // use first content type
	}
	return rb
}

// NormalizeResponses converts raw responses into normalized Response values.
func NormalizeResponses(raw map[string]RawResponse) []schema.Response {
	responses := make([]schema.Response, 0, len(raw))
	for code, resp := range raw {
		r := schema.Response{StatusCode: code}
		for _, media := range resp.Content {
			r.Properties = NormalizeProperties(media.Schema)
			break
		}
		responses = append(responses, r)
	}
	return responses
}

// NormalizeProperties converts a raw schema into a list of normalized Property values.
func NormalizeProperties(s RawSchema) []schema.Property {
	var props []schema.Property
	requiredSet := make(map[string]bool)
	for _, r := range s.Required {
		requiredSet[r] = true
	}
	for name, ps := range s.Properties {
		prop := schema.Property{
			Name:        name,
			Type:        ps.Type,
			Format:      ps.Format,
			Required:    requiredSet[name],
			Ref:         ps.Ref,
			Description: ps.Description,
		}
		for _, e := range ps.Enum {
			prop.Enum = append(prop.Enum, fmt.Sprintf("%v", e))
		}
		if ps.Items != nil {
			items := schema.Property{
				Type:   ps.Items.Type,
				Ref:    ps.Items.Ref,
				Format: ps.Items.Format,
			}
			prop.Items = &items
		}
		props = append(props, prop)
	}
	return props
}
