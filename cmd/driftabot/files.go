package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// swaggerSpecExists reports whether a swagger/openapi spec file already exists
// in common locations under dir.
func swaggerSpecExists(dir string) bool {
	candidates := []string{
		"swagger.json", "swagger.yaml", "swagger.yml",
		"openapi.json", "openapi.yaml", "openapi.yml",
		"docs/swagger.json", "docs/swagger.yaml",
		"api/swagger.json", "api/openapi.json",
	}
	for _, rel := range candidates {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			return true
		}
	}
	return false
}

// swaggerScriptExists reports whether a swagger generation script or tsoa config exists in dir.
func swaggerScriptExists(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "tsoa.json")); err == nil {
		return true
	}
	return findSwaggerScript(dir) != ""
}

// findSwaggerScript returns the relative path of the first swagger generation
// script found in dir, or empty string if none is found.
func findSwaggerScript(dir string) string {
	candidates := []string{
		"driftabot/scripts/generate-swagger.ts",
		"driftabot/scripts/generate-swagger.js",
		"scripts/generate-swagger.ts",
		"scripts/generate-swagger.js",
		"src/generate-swagger.ts",
		"generate-swagger.ts",
	}
	for _, rel := range candidates {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			return rel
		}
	}
	return ""
}

// findSchemaFile finds the generated OpenAPI schema file in dir.
func findSchemaFile(dir string) (string, error) {
	for _, name := range []string{"swagger.yaml", "swagger.json", "docs.yaml", "docs.json"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no schema file found in %s", dir)
}

// findGraphQLFile finds the generated GraphQL schema file in dir.
func findGraphQLFile(dir string) (string, error) {
	for _, name := range []string{"schema.graphql", "schema.gql"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no GraphQL schema file found in %s", dir)
}

// findProtoFile finds the generated proto schema file in dir.
func findProtoFile(dir string) (string, error) {
	p := filepath.Join(dir, "schema.proto")
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("no .proto file found in %s", dir)
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
