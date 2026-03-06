package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"drift-guard-diff-engine/pkg/schema"

	"gopkg.in/yaml.v3"
)

// rawOpenAPI is the raw unmarshalled OpenAPI 3.x document.
type rawOpenAPI struct {
	Info struct {
		Title   string `json:"title" yaml:"title"`
		Version string `json:"version" yaml:"version"`
	} `json:"info" yaml:"info"`
	Paths      map[string]rawPathItem `json:"paths" yaml:"paths"`
	Components rawComponents          `json:"components" yaml:"components"`
}

type rawComponents struct {
	Schemas map[string]rawSchema `json:"schemas" yaml:"schemas"`
}

type rawPathItem struct {
	Get     *rawOperation `json:"get" yaml:"get"`
	Post    *rawOperation `json:"post" yaml:"post"`
	Put     *rawOperation `json:"put" yaml:"put"`
	Patch   *rawOperation `json:"patch" yaml:"patch"`
	Delete  *rawOperation `json:"delete" yaml:"delete"`
	Head    *rawOperation `json:"head" yaml:"head"`
	Options *rawOperation `json:"options" yaml:"options"`
}

type rawOperation struct {
	OperationID string                 `json:"operationId" yaml:"operationId"`
	Parameters  []rawParameter         `json:"parameters" yaml:"parameters"`
	RequestBody *rawRequestBody        `json:"requestBody" yaml:"requestBody"`
	Responses   map[string]rawResponse `json:"responses" yaml:"responses"`
}

type rawParameter struct {
	Name     string    `json:"name" yaml:"name"`
	In       string    `json:"in" yaml:"in"`
	Required bool      `json:"required" yaml:"required"`
	Ref      string    `json:"$ref" yaml:"$ref"`
	Schema   rawSchema `json:"schema" yaml:"schema"`
}

type rawRequestBody struct {
	Required bool                    `json:"required" yaml:"required"`
	Content  map[string]rawMediaType `json:"content" yaml:"content"`
}

type rawMediaType struct {
	Schema rawSchema `json:"schema" yaml:"schema"`
}

type rawResponse struct {
	Ref     string                  `json:"$ref" yaml:"$ref"`
	Content map[string]rawMediaType `json:"content" yaml:"content"`
}

type rawSchema struct {
	Ref         string               `json:"$ref" yaml:"$ref"`
	Type        string               `json:"type" yaml:"type"`
	Format      string               `json:"format" yaml:"format"`
	Description string               `json:"description" yaml:"description"`
	Required    []string             `json:"required" yaml:"required"`
	Properties  map[string]rawSchema `json:"properties" yaml:"properties"`
	Items       *rawSchema           `json:"items" yaml:"items"`
	Enum        []interface{}        `json:"enum" yaml:"enum"`
}

// Parse reads and parses an OpenAPI 3.x file (JSON or YAML) into a normalized Schema.
func Parse(path string) (*schema.Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	var raw rawOpenAPI
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing YAML %s: %w", path, err)
		}
	case ".json":
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing JSON %s: %w", path, err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &raw); err != nil {
			if err2 := json.Unmarshal(data, &raw); err2 != nil {
				return nil, fmt.Errorf("unsupported format for %s", path)
			}
		}
	}

	return normalize(&raw), nil
}

func normalize(raw *rawOpenAPI) *schema.Schema {
	s := &schema.Schema{
		Title:   raw.Info.Title,
		Version: raw.Info.Version,
	}

	for path, item := range raw.Paths {
		endpoint := schema.Endpoint{Path: path}
		ops := map[string]*rawOperation{
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
				Parameters:  normalizeParams(rawOp.Parameters),
				RequestBody: normalizeRequestBody(rawOp.RequestBody),
				Responses:   normalizeResponses(rawOp.Responses),
			}
			endpoint.Operations = append(endpoint.Operations, op)
		}
		s.Endpoints = append(s.Endpoints, endpoint)
	}

	return s
}

func normalizeParams(raw []rawParameter) []schema.Parameter {
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

func normalizeRequestBody(raw *rawRequestBody) *schema.RequestBody {
	if raw == nil {
		return nil
	}
	rb := &schema.RequestBody{Required: raw.Required}
	for _, media := range raw.Content {
		rb.Properties = normalizeProperties(media.Schema)
		break // use first content type
	}
	return rb
}

func normalizeResponses(raw map[string]rawResponse) []schema.Response {
	responses := make([]schema.Response, 0, len(raw))
	for code, resp := range raw {
		r := schema.Response{StatusCode: code}
		for _, media := range resp.Content {
			r.Properties = normalizeProperties(media.Schema)
			break
		}
		responses = append(responses, r)
	}
	return responses
}

func normalizeProperties(s rawSchema) []schema.Property {
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
