package languages_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DriftaBot/driftabot-engine/internal/languages"
)

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

// makeTempDir creates a temporary directory and returns its path.
// The caller is responsible for cleanup via t.Cleanup or defer.
func makeTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "driftabot-detect-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

const nestPackageJSON = `{
  "dependencies": {
    "@nestjs/core": "^9.0.0",
    "@nestjs/common": "^9.0.0"
  }
}`

const expressPackageJSON = `{
  "dependencies": {
    "express": "^4.18.0"
  }
}`

const genericNodePackageJSON = `{
  "dependencies": {
    "axios": "^1.0.0"
  }
}`

const nestGraphQLPackageJSON = `{
  "dependencies": {
    "@nestjs/core": "^9.0.0",
    "@nestjs/graphql": "^11.0.0"
  }
}`

// --------------------------------------------------------------------------
// DetectGenerator
// --------------------------------------------------------------------------

func TestDetectGenerator_Go(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")

	gen, err := languages.DetectGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil generator for Go project")
	}
}

func TestDetectGenerator_GoLegacy_GopkgTOML(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "Gopkg.toml", "[[constraint]]\n  name = \"github.com/gorilla/mux\"\n")
	writeFile(t, dir, "main.go", "package main\n")

	gen, err := languages.DetectGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error for legacy Go project: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil generator for legacy Go project")
	}
}

func TestDetectGenerator_GoLegacy_GoSourceOnly(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "main.go", "package main\n")

	gen, err := languages.DetectGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error for Go-source-only project: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil generator for Go-source-only project")
	}
}

func TestDetectGenerator_GoGin(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n\nrequire github.com/gin-gonic/gin v1.9.1\n")

	gen, err := languages.DetectGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil generator for Go (Gin) project")
	}
}

func TestDetectGenerator_GoEcho(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n\nrequire github.com/labstack/echo/v4 v4.11.0\n")

	gen, err := languages.DetectGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil generator for Go (Echo) project")
	}
}

func TestDetectGenerator_NestJS(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", nestPackageJSON)

	gen, err := languages.DetectGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil generator for NestJS project")
	}
}

func TestDetectGenerator_Express(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", expressPackageJSON)

	gen, err := languages.DetectGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil generator for Express project")
	}
}

func TestDetectGenerator_GenericNode(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", genericNodePackageJSON)

	gen, err := languages.DetectGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil generator for generic Node.js project")
	}
}

func TestDetectGenerator_Python_ReturnsError(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "pyproject.toml", "[tool.poetry]\nname = \"app\"\n")

	_, err := languages.DetectGenerator(dir)
	if err == nil {
		t.Fatal("expected error for Python project")
	}
}

func TestDetectGenerator_Unknown_ReturnsError(t *testing.T) {
	dir := makeTempDir(t)

	_, err := languages.DetectGenerator(dir)
	if err == nil {
		t.Fatal("expected error for unknown project type")
	}
}

// --------------------------------------------------------------------------
// DetectProjectInfo
// --------------------------------------------------------------------------

func TestDetectProjectInfo_Go(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TypeName != "Go" {
		t.Errorf("expected TypeName='Go', got '%s'", info.TypeName)
	}
	if info.Generate == nil {
		t.Error("expected non-nil Generate func")
	}
}

func TestDetectProjectInfo_GoGin(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n\nrequire github.com/gin-gonic/gin v1.9.1\n")

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TypeName != "Go (Gin)" {
		t.Errorf("expected TypeName='Go (Gin)', got '%s'", info.TypeName)
	}
	if info.Generate == nil {
		t.Error("expected non-nil Generate func")
	}
}

func TestDetectProjectInfo_GoEcho(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n\nrequire github.com/labstack/echo/v4 v4.11.0\n")

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TypeName != "Go (Echo)" {
		t.Errorf("expected TypeName='Go (Echo)', got '%s'", info.TypeName)
	}
}

func TestDetectProjectInfo_GoFiber(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n\nrequire github.com/gofiber/fiber/v2 v2.52.0\n")

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TypeName != "Go (Fiber)" {
		t.Errorf("expected TypeName='Go (Fiber)', got '%s'", info.TypeName)
	}
}

func TestDetectProjectInfo_GoChi(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n\nrequire github.com/go-chi/chi/v5 v5.1.0\n")

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TypeName != "Go (Chi)" {
		t.Errorf("expected TypeName='Go (Chi)', got '%s'", info.TypeName)
	}
}

func TestDetectProjectInfo_GoNoFramework(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TypeName != "Go" {
		t.Errorf("expected TypeName='Go' for plain Go project, got '%s'", info.TypeName)
	}
}

func TestDetectProjectInfo_GoGorillaMux_ModFile(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n\nrequire github.com/gorilla/mux v1.8.0\n")

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TypeName != "Go (Gorilla Mux)" {
		t.Errorf("expected TypeName='Go (Gorilla Mux)', got '%s'", info.TypeName)
	}
}

func TestDetectProjectInfo_GoGorillaMux_GopkgTOML(t *testing.T) {
	dir := makeTempDir(t)
	// No go.mod — legacy dep project
	writeFile(t, dir, "Gopkg.toml", "[[constraint]]\n  name = \"github.com/gorilla/mux\"\n  version = \"1.6.2\"\n")
	writeFile(t, dir, "main.go", "package main\n")

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TypeName != "Go (Gorilla Mux)" {
		t.Errorf("expected TypeName='Go (Gorilla Mux)', got '%s'", info.TypeName)
	}
}

func TestDetectProjectInfo_GoLegacy_NoFramework(t *testing.T) {
	dir := makeTempDir(t)
	// No go.mod — just .go source files
	writeFile(t, dir, "main.go", "package main\n")

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		t.Fatalf("unexpected error for legacy Go project: %v", err)
	}
	if info.TypeName != "Go" {
		t.Errorf("expected TypeName='Go' for legacy project without framework, got '%s'", info.TypeName)
	}
}

func TestDetectProjectInfo_NestJS(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", nestPackageJSON)

	info, err := languages.DetectProjectInfo(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TypeName != "NestJS" {
		t.Errorf("expected TypeName='NestJS', got '%s'", info.TypeName)
	}
}

func TestDetectProjectInfo_Python_ReturnsError(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "pyproject.toml", "[tool.poetry]\nname = \"app\"\n")

	_, err := languages.DetectProjectInfo(dir)
	if err == nil {
		t.Fatal("expected error for Python project")
	}
}

func TestDetectProjectInfo_Unknown_ReturnsError(t *testing.T) {
	dir := makeTempDir(t)
	_, err := languages.DetectProjectInfo(dir)
	if err == nil {
		t.Fatal("expected error for unknown project type")
	}
}

// --------------------------------------------------------------------------
// DetectGraphQLInfo
// --------------------------------------------------------------------------

func TestDetectGraphQLInfo_NestJSWithGraphQL(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", nestGraphQLPackageJSON)

	info := languages.DetectGraphQLInfo(dir)
	if info == nil {
		t.Fatal("expected non-nil GraphQL info for NestJS+GraphQL project")
	}
	if info.TypeName != "NestJS" {
		t.Errorf("expected TypeName='NestJS', got '%s'", info.TypeName)
	}
}

func TestDetectGraphQLInfo_NestJSWithSchemaFile(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", nestPackageJSON)
	writeFile(t, dir, "schema.graphql", "type Query { ping: String }")

	info := languages.DetectGraphQLInfo(dir)
	if info == nil {
		t.Fatal("expected non-nil GraphQL info for NestJS project with schema.graphql")
	}
}

func TestDetectGraphQLInfo_GoWithSchemaFile(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n\nrequire github.com/gin-gonic/gin v1.9.1\n")
	writeFile(t, dir, "schema.graphql", "type Query { ping: String }")

	info := languages.DetectGraphQLInfo(dir)
	if info == nil {
		t.Fatal("expected non-nil GraphQL info for Go project with schema.graphql")
	}
	if info.TypeName != "Go (Gin)" {
		t.Errorf("expected TypeName='Go (Gin)', got '%s'", info.TypeName)
	}
	if info.GenerateGQL == nil {
		t.Error("expected non-nil GenerateGQL func")
	}
}

func TestDetectGraphQLInfo_GoNoSchemaFile_ReturnsNil(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")

	info := languages.DetectGraphQLInfo(dir)
	if info != nil {
		t.Errorf("expected nil GraphQL info for Go project without schema, got %+v", info)
	}
}

func TestDetectGraphQLInfo_ExpressWithApollo(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", `{"dependencies": {"express": "^4.0.0", "apollo-server-express": "^3.0.0"}}`)

	info := languages.DetectGraphQLInfo(dir)
	if info == nil {
		t.Fatal("expected non-nil GraphQL info for Express+Apollo project")
	}
}

// --------------------------------------------------------------------------
// DetectGRPCInfo
// --------------------------------------------------------------------------

func TestDetectGRPCInfo_NestJSWithProto(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", nestPackageJSON)
	if err := os.MkdirAll(filepath.Join(dir, "proto"), 0755); err != nil {
		t.Fatalf("mkdir proto: %v", err)
	}
	writeFile(t, dir, "proto/user.proto", `syntax = "proto3"; service UserService {}`)

	info := languages.DetectGRPCInfo(dir)
	if info == nil {
		t.Fatal("expected non-nil gRPC info for NestJS project with .proto files")
	}
	if info.TypeName != "NestJS" {
		t.Errorf("expected TypeName='NestJS', got '%s'", info.TypeName)
	}
}

func TestDetectGRPCInfo_NoProto_ReturnsNil(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", nestPackageJSON)

	info := languages.DetectGRPCInfo(dir)
	if info != nil {
		t.Errorf("expected nil gRPC info when no .proto files, got %+v", info)
	}
}

func TestDetectGRPCInfo_GoWithProto(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n\nrequire github.com/gin-gonic/gin v1.9.1\n")
	if err := os.MkdirAll(filepath.Join(dir, "proto"), 0755); err != nil {
		t.Fatalf("mkdir proto: %v", err)
	}
	writeFile(t, dir, "proto/service.proto", `syntax = "proto3"; service UserService {}`)

	info := languages.DetectGRPCInfo(dir)
	if info == nil {
		t.Fatal("expected non-nil gRPC info for Go project with .proto files")
	}
	if info.TypeName != "Go (Gin)" {
		t.Errorf("expected TypeName='Go (Gin)', got '%s'", info.TypeName)
	}
	if info.GenerateRPC == nil {
		t.Error("expected non-nil GenerateRPC func")
	}
}

func TestDetectGRPCInfo_GoNoProto_ReturnsNil(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")

	info := languages.DetectGRPCInfo(dir)
	if info != nil {
		t.Errorf("expected nil gRPC info for Go project without .proto files, got %+v", info)
	}
}

// --------------------------------------------------------------------------
// DetectGraphQLGenerator
// --------------------------------------------------------------------------

func TestDetectGraphQLGenerator_Go(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")

	gen, err := languages.DetectGraphQLGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil GraphQL generator for Go project")
	}
}

func TestDetectGraphQLGenerator_NestJS(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", nestPackageJSON)

	gen, err := languages.DetectGraphQLGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil GraphQL generator for NestJS project")
	}
}

func TestDetectGraphQLGenerator_Unknown_ReturnsError(t *testing.T) {
	dir := makeTempDir(t)

	_, err := languages.DetectGraphQLGenerator(dir)
	if err == nil {
		t.Fatal("expected error for unknown project type")
	}
}

// --------------------------------------------------------------------------
// DetectGRPCGenerator
// --------------------------------------------------------------------------

func TestDetectGRPCGenerator_Go(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.21\n")

	gen, err := languages.DetectGRPCGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil gRPC generator for Go project")
	}
}

func TestDetectGRPCGenerator_NestJS(t *testing.T) {
	dir := makeTempDir(t)
	writeFile(t, dir, "package.json", nestPackageJSON)

	gen, err := languages.DetectGRPCGenerator(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen == nil {
		t.Fatal("expected non-nil gRPC generator for NestJS project")
	}
}

func TestDetectGRPCGenerator_Unknown_ReturnsError(t *testing.T) {
	dir := makeTempDir(t)

	_, err := languages.DetectGRPCGenerator(dir)
	if err == nil {
		t.Fatal("expected error for unknown project type")
	}
}
