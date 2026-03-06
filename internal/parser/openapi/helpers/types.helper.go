// Package helpers provides internal types and normalization utilities for the OpenAPI parser.
package helpers

// RawOpenAPI is the raw unmarshalled OpenAPI 3.x document.
type RawOpenAPI struct {
	Info struct {
		Title   string `json:"title" yaml:"title"`
		Version string `json:"version" yaml:"version"`
	} `json:"info" yaml:"info"`
	Paths      map[string]RawPathItem `json:"paths" yaml:"paths"`
	Components RawComponents          `json:"components" yaml:"components"`
}

// RawComponents holds reusable component definitions.
type RawComponents struct {
	Schemas map[string]RawSchema `json:"schemas" yaml:"schemas"`
}

// RawPathItem holds all HTTP method operations for a path.
type RawPathItem struct {
	Get     *RawOperation `json:"get" yaml:"get"`
	Post    *RawOperation `json:"post" yaml:"post"`
	Put     *RawOperation `json:"put" yaml:"put"`
	Patch   *RawOperation `json:"patch" yaml:"patch"`
	Delete  *RawOperation `json:"delete" yaml:"delete"`
	Head    *RawOperation `json:"head" yaml:"head"`
	Options *RawOperation `json:"options" yaml:"options"`
}

// RawOperation represents a single HTTP operation.
type RawOperation struct {
	OperationID string                 `json:"operationId" yaml:"operationId"`
	Parameters  []RawParameter         `json:"parameters" yaml:"parameters"`
	RequestBody *RawRequestBody        `json:"requestBody" yaml:"requestBody"`
	Responses   map[string]RawResponse `json:"responses" yaml:"responses"`
}

// RawParameter represents a single query/path/header parameter.
type RawParameter struct {
	Name     string    `json:"name" yaml:"name"`
	In       string    `json:"in" yaml:"in"`
	Required bool      `json:"required" yaml:"required"`
	Ref      string    `json:"$ref" yaml:"$ref"`
	Schema   RawSchema `json:"schema" yaml:"schema"`
}

// RawRequestBody represents a request body definition.
type RawRequestBody struct {
	Required bool                   `json:"required" yaml:"required"`
	Content  map[string]RawMediaType `json:"content" yaml:"content"`
}

// RawMediaType wraps a schema for a specific media type.
type RawMediaType struct {
	Schema RawSchema `json:"schema" yaml:"schema"`
}

// RawResponse represents a response definition.
type RawResponse struct {
	Ref     string                  `json:"$ref" yaml:"$ref"`
	Content map[string]RawMediaType `json:"content" yaml:"content"`
}

// RawSchema represents a JSON Schema / OpenAPI schema object.
type RawSchema struct {
	Ref         string               `json:"$ref" yaml:"$ref"`
	Type        string               `json:"type" yaml:"type"`
	Format      string               `json:"format" yaml:"format"`
	Description string               `json:"description" yaml:"description"`
	Required    []string             `json:"required" yaml:"required"`
	Properties  map[string]RawSchema `json:"properties" yaml:"properties"`
	Items       *RawSchema           `json:"items" yaml:"items"`
	Enum        []interface{}        `json:"enum" yaml:"enum"`
}
