package openapi_test

import (
	"os"
	"testing"

	"github.com/pgomes13/api-drift-engine/internal/parser/openapi"
	"github.com/pgomes13/api-drift-engine/pkg/schema"
)

const testdataDir = "../../testdata/"

func TestParse_ReturnsSchema(t *testing.T) {
	s, err := openapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil schema")
	}
}

func TestParse_Title(t *testing.T) {
	s, err := openapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Title != "Users API" {
		t.Errorf("expected title 'Users API', got '%s'", s.Title)
	}
}

func TestParse_EndpointCount(t *testing.T) {
	s, err := openapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// base.yaml has /users and /users/{id}
	if len(s.Endpoints) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(s.Endpoints))
	}
}

func TestParse_MethodsOnEndpoint(t *testing.T) {
	s, err := openapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ep := findEndpoint(s.Endpoints, "/users/{id}")
	if ep == nil {
		t.Fatal("endpoint '/users/{id}' not found")
	}
	// GET and DELETE
	if len(ep.Operations) != 2 {
		t.Errorf("expected 2 operations on /users/{id}, got %d", len(ep.Operations))
	}
}

func TestParse_PathParameter(t *testing.T) {
	s, err := openapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ep := findEndpoint(s.Endpoints, "/users/{id}")
	if ep == nil {
		t.Fatal("endpoint '/users/{id}' not found")
	}
	op := findOperation(ep.Operations, "GET")
	if op == nil {
		t.Fatal("GET operation not found on /users/{id}")
	}
	p := findParam(op.Parameters, "id")
	if p == nil {
		t.Fatal("param 'id' not found")
	}
	if p.In != "path" {
		t.Errorf("expected In='path', got '%s'", p.In)
	}
	if p.Type != "string" {
		t.Errorf("expected Type='string', got '%s'", p.Type)
	}
	if !p.Required {
		t.Error("expected param 'id' to be required")
	}
}

func TestParse_QueryParameter(t *testing.T) {
	s, err := openapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ep := findEndpoint(s.Endpoints, "/users")
	if ep == nil {
		t.Fatal("endpoint '/users' not found")
	}
	op := findOperation(ep.Operations, "GET")
	if op == nil {
		t.Fatal("GET operation not found on /users")
	}
	p := findParam(op.Parameters, "limit")
	if p == nil {
		t.Fatal("param 'limit' not found")
	}
	if p.In != "query" {
		t.Errorf("expected In='query', got '%s'", p.In)
	}
	if p.Required {
		t.Error("expected param 'limit' to be optional")
	}
}

func TestParse_RequestBodyProperties(t *testing.T) {
	s, err := openapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ep := findEndpoint(s.Endpoints, "/users")
	if ep == nil {
		t.Fatal("endpoint '/users' not found")
	}
	op := findOperation(ep.Operations, "POST")
	if op == nil {
		t.Fatal("POST operation not found on /users")
	}
	if op.RequestBody == nil {
		t.Fatal("expected request body on POST /users")
	}
	if !op.RequestBody.Required {
		t.Error("expected request body to be required")
	}
	if findProperty(op.RequestBody.Properties, "name") == nil {
		t.Error("expected property 'name' in request body")
	}
	if findProperty(op.RequestBody.Properties, "email") == nil {
		t.Error("expected property 'email' in request body")
	}
}

func TestParse_ResponseProperties(t *testing.T) {
	s, err := openapi.Parse(testdataDir + "base.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ep := findEndpoint(s.Endpoints, "/users")
	if ep == nil {
		t.Fatal("endpoint '/users' not found")
	}
	op := findOperation(ep.Operations, "GET")
	if op == nil {
		t.Fatal("GET operation not found on /users")
	}
	resp := findResponse(op.Responses, "200")
	if resp == nil {
		t.Fatal("response '200' not found on GET /users")
	}
	if findProperty(resp.Properties, "total") == nil {
		t.Error("expected property 'total' in GET /users 200 response")
	}
}

func TestParse_JSON(t *testing.T) {
	const jsonContent = `{
  "openapi": "3.0.0",
  "info": {"title": "JSON API", "version": "1.0.0"},
  "paths": {
    "/items": {
      "get": {
        "operationId": "listItems",
        "responses": {"200": {}}
      }
    }
  }
}`
	f, err := os.CreateTemp("", "openapi-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(jsonContent); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()

	s, err := openapi.Parse(f.Name())
	if err != nil {
		t.Fatalf("unexpected error parsing JSON: %v", err)
	}
	if s.Title != "JSON API" {
		t.Errorf("expected title 'JSON API', got '%s'", s.Title)
	}
	if len(s.Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(s.Endpoints))
	}
}

func TestParse_MissingFile(t *testing.T) {
	_, err := openapi.Parse("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

func findEndpoint(endpoints []schema.Endpoint, path string) *schema.Endpoint {
	for i := range endpoints {
		if endpoints[i].Path == path {
			return &endpoints[i]
		}
	}
	return nil
}

func findOperation(ops []schema.Operation, method string) *schema.Operation {
	for i := range ops {
		if ops[i].Method == method {
			return &ops[i]
		}
	}
	return nil
}

func findParam(params []schema.Parameter, name string) *schema.Parameter {
	for i := range params {
		if params[i].Name == name {
			return &params[i]
		}
	}
	return nil
}

func findProperty(props []schema.Property, name string) *schema.Property {
	for i := range props {
		if props[i].Name == name {
			return &props[i]
		}
	}
	return nil
}

func findResponse(responses []schema.Response, code string) *schema.Response {
	for i := range responses {
		if responses[i].StatusCode == code {
			return &responses[i]
		}
	}
	return nil
}
