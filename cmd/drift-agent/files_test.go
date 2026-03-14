package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "drift-agent-files-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func touch(t *testing.T, dir, rel string) string {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(p, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", rel, err)
	}
	return p
}

// --------------------------------------------------------------------------
// swaggerSpecExists
// --------------------------------------------------------------------------

func TestSwaggerSpecExists_False_EmptyDir(t *testing.T) {
	if swaggerSpecExists(tempDir(t)) {
		t.Error("expected false for empty dir")
	}
}

func TestSwaggerSpecExists_True_SwaggerJSON(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "swagger.json")
	if !swaggerSpecExists(dir) {
		t.Error("expected true when swagger.json present")
	}
}

func TestSwaggerSpecExists_True_OpenAPIYAML(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "openapi.yaml")
	if !swaggerSpecExists(dir) {
		t.Error("expected true when openapi.yaml present")
	}
}

func TestSwaggerSpecExists_True_DocsSwaggerJSON(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "docs/swagger.json")
	if !swaggerSpecExists(dir) {
		t.Error("expected true when docs/swagger.json present")
	}
}

func TestSwaggerSpecExists_True_APIOpenAPIJSON(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "api/openapi.json")
	if !swaggerSpecExists(dir) {
		t.Error("expected true when api/openapi.json present")
	}
}

// --------------------------------------------------------------------------
// swaggerScriptExists / findSwaggerScript
// --------------------------------------------------------------------------

func TestSwaggerScriptExists_False_EmptyDir(t *testing.T) {
	if swaggerScriptExists(tempDir(t)) {
		t.Error("expected false for empty dir")
	}
}

func TestSwaggerScriptExists_True_TsoaJSON(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "tsoa.json")
	if !swaggerScriptExists(dir) {
		t.Error("expected true when tsoa.json present")
	}
}

func TestSwaggerScriptExists_True_ScriptsTS(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "scripts/generate-swagger.ts")
	if !swaggerScriptExists(dir) {
		t.Error("expected true when scripts/generate-swagger.ts present")
	}
}

func TestFindSwaggerScript_Empty_NoMatch(t *testing.T) {
	if got := findSwaggerScript(tempDir(t)); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFindSwaggerScript_DriftAgentScriptsTS(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "drift-agent/scripts/generate-swagger.ts")
	got := findSwaggerScript(dir)
	if got != "drift-agent/scripts/generate-swagger.ts" {
		t.Errorf("expected drift-agent/scripts/generate-swagger.ts, got %q", got)
	}
}

func TestFindSwaggerScript_RootGenerateSwaggerTS(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "generate-swagger.ts")
	got := findSwaggerScript(dir)
	if got != "generate-swagger.ts" {
		t.Errorf("expected generate-swagger.ts, got %q", got)
	}
}

func TestFindSwaggerScript_PrefersFirstMatch(t *testing.T) {
	dir := tempDir(t)
	// Both present — drift-agent/scripts/ should win (it's listed first).
	touch(t, dir, "drift-agent/scripts/generate-swagger.ts")
	touch(t, dir, "generate-swagger.ts")
	got := findSwaggerScript(dir)
	if got != "drift-agent/scripts/generate-swagger.ts" {
		t.Errorf("expected first candidate to win, got %q", got)
	}
}

// --------------------------------------------------------------------------
// findSchemaFile
// --------------------------------------------------------------------------

func TestFindSchemaFile_Error_EmptyDir(t *testing.T) {
	_, err := findSchemaFile(tempDir(t))
	if err == nil {
		t.Error("expected error for empty dir")
	}
}

func TestFindSchemaFile_SwaggerYAML(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "swagger.yaml")
	got, err := findSchemaFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, "swagger.yaml") {
		t.Errorf("expected swagger.yaml path, got %q", got)
	}
}

func TestFindSchemaFile_SwaggerJSON(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "swagger.json")
	got, err := findSchemaFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, "swagger.json") {
		t.Errorf("expected swagger.json path, got %q", got)
	}
}

func TestFindSchemaFile_PrefersSwaggerYAMLOverJSON(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "swagger.yaml")
	touch(t, dir, "swagger.json")
	got, err := findSchemaFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, "swagger.yaml") {
		t.Errorf("expected swagger.yaml to win over swagger.json, got %q", got)
	}
}

// --------------------------------------------------------------------------
// findGraphQLFile
// --------------------------------------------------------------------------

func TestFindGraphQLFile_Error_EmptyDir(t *testing.T) {
	_, err := findGraphQLFile(tempDir(t))
	if err == nil {
		t.Error("expected error for empty dir")
	}
}

func TestFindGraphQLFile_SchemaGraphQL(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "schema.graphql")
	got, err := findGraphQLFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, "schema.graphql") {
		t.Errorf("expected schema.graphql path, got %q", got)
	}
}

func TestFindGraphQLFile_SchemaGQL(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "schema.gql")
	got, err := findGraphQLFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, "schema.gql") {
		t.Errorf("expected schema.gql path, got %q", got)
	}
}

// --------------------------------------------------------------------------
// findProtoFile
// --------------------------------------------------------------------------

func TestFindProtoFile_Error_EmptyDir(t *testing.T) {
	_, err := findProtoFile(tempDir(t))
	if err == nil {
		t.Error("expected error for empty dir")
	}
}

func TestFindProtoFile_SchemaProto(t *testing.T) {
	dir := tempDir(t)
	touch(t, dir, "schema.proto")
	got, err := findProtoFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, "schema.proto") {
		t.Errorf("expected schema.proto path, got %q", got)
	}
}

// --------------------------------------------------------------------------
// pathExists
// --------------------------------------------------------------------------

func TestPathExists_True(t *testing.T) {
	dir := tempDir(t)
	p := touch(t, dir, "file.txt")
	if !pathExists(p) {
		t.Error("expected true for existing file")
	}
}

func TestPathExists_False(t *testing.T) {
	if pathExists(filepath.Join(tempDir(t), "nonexistent.txt")) {
		t.Error("expected false for non-existent file")
	}
}

func TestPathExists_Directory(t *testing.T) {
	if !pathExists(tempDir(t)) {
		t.Error("expected true for existing directory")
	}
}

// --------------------------------------------------------------------------
// copyFile
// --------------------------------------------------------------------------

func TestCopyFile_CopiesContent(t *testing.T) {
	dir := tempDir(t)
	src := touch(t, dir, "src.txt")
	if err := os.WriteFile(src, []byte("hello drift-agent"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	dst := filepath.Join(dir, "dst.txt")

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "hello drift-agent" {
		t.Errorf("expected 'hello drift-agent', got %q", string(got))
	}
}

func TestCopyFile_CreatesParentDirs(t *testing.T) {
	dir := tempDir(t)
	src := touch(t, dir, "src.txt")
	dst := filepath.Join(dir, "nested", "deep", "dst.txt")

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	if !pathExists(dst) {
		t.Error("expected dst to exist after copy")
	}
}

func TestCopyFile_ErrorMissingSrc(t *testing.T) {
	dir := tempDir(t)
	err := copyFile(filepath.Join(dir, "nonexistent.txt"), filepath.Join(dir, "dst.txt"))
	if err == nil {
		t.Error("expected error for missing source file")
	}
}
