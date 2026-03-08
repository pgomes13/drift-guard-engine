package main

import "testing"

// --------------------------------------------------------------------------
// openAPIScaffoldNeeded
// --------------------------------------------------------------------------

func TestOpenAPIScaffoldNeeded_Go_False(t *testing.T) {
	for _, typeName := range []string{"Go", "Go (Gin)", "Go (Echo)", "Go (Fiber)", "Go (Chi)"} {
		if openAPIScaffoldNeeded(typeName) {
			t.Errorf("expected openAPIScaffoldNeeded(%q) = false", typeName)
		}
	}
}

func TestOpenAPIScaffoldNeeded_NestJS_True(t *testing.T) {
	if !openAPIScaffoldNeeded("NestJS") {
		t.Error("expected openAPIScaffoldNeeded(\"NestJS\") = true")
	}
}

func TestOpenAPIScaffoldNeeded_Express_True(t *testing.T) {
	if !openAPIScaffoldNeeded("Express") {
		t.Error("expected openAPIScaffoldNeeded(\"Express\") = true")
	}
}

func TestOpenAPIScaffoldNeeded_NodeJS_True(t *testing.T) {
	if !openAPIScaffoldNeeded("Node.js") {
		t.Error("expected openAPIScaffoldNeeded(\"Node.js\") = true")
	}
}

// --------------------------------------------------------------------------
// Go project + OpenAPI spec scenarios
// --------------------------------------------------------------------------

// swag init writes its output to docs/ by default.
func TestGoProject_SwaggerSpecExists_DocsSwaggerYAML(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "go.mod")
	touch(t, dir, "docs/swagger.yaml")

	if !swaggerSpecExists(dir) {
		t.Error("expected swaggerSpecExists=true for docs/swagger.yaml (swag default output)")
	}
}

func TestGoProject_SwaggerSpecExists_DocsSwaggerJSON(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "go.mod")
	touch(t, dir, "docs/swagger.json")

	if !swaggerSpecExists(dir) {
		t.Error("expected swaggerSpecExists=true for docs/swagger.json")
	}
}

func TestGoProject_NoSpec_SwaggerSpecExists_False(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "go.mod")

	if swaggerSpecExists(dir) {
		t.Error("expected swaggerSpecExists=false when no spec file exists")
	}
}

// A Go project with docs/swagger.yaml should need no scaffold regardless.
func TestGoGin_WithSpec_NoScaffoldNeeded(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "go.mod")
	touch(t, dir, "docs/swagger.yaml")

	specPresent := swaggerSpecExists(dir)
	scaffoldNeeded := openAPIScaffoldNeeded("Go (Gin)")

	if !specPresent {
		t.Error("expected spec to be present")
	}
	if scaffoldNeeded {
		t.Error("expected no scaffold needed for Go (Gin)")
	}
}

// A Go project without any spec still needs no scaffold — swag is run directly.
func TestGoGin_WithoutSpec_NoScaffoldNeeded(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "go.mod")

	specPresent := swaggerSpecExists(dir) || swaggerScriptExists(dir)
	scaffoldNeeded := openAPIScaffoldNeeded("Go (Gin)")

	if specPresent {
		t.Error("expected no spec present")
	}
	if scaffoldNeeded {
		t.Error("expected no scaffold needed for Go (Gin) even without spec")
	}
}
