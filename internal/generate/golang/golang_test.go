package golang

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
	dir, err := os.MkdirTemp("", "drift-agent-golang-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func writeFile(t *testing.T, dir, rel, content string) string {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", rel, err)
	}
	return p
}

// --------------------------------------------------------------------------
// FindGraphQLSchema
// --------------------------------------------------------------------------

func TestFindGraphQLSchema_Empty_ReturnsEmpty(t *testing.T) {
	if got := FindGraphQLSchema(tempDir(t)); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFindGraphQLSchema_RootSchemaGraphQL(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "schema.graphql", "type Query { ping: String }")

	got := FindGraphQLSchema(dir)
	if !strings.HasSuffix(got, "schema.graphql") {
		t.Errorf("expected schema.graphql path, got %q", got)
	}
}

func TestFindGraphQLSchema_RootSchemaGQL(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "schema.gql", "type Query { ping: String }")

	got := FindGraphQLSchema(dir)
	if !strings.HasSuffix(got, "schema.gql") {
		t.Errorf("expected schema.gql path, got %q", got)
	}
}

func TestFindGraphQLSchema_SrcSubdir(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "src/schema.graphql", "type Query { ping: String }")

	got := FindGraphQLSchema(dir)
	if !strings.HasSuffix(got, filepath.Join("src", "schema.graphql")) {
		t.Errorf("expected src/schema.graphql path, got %q", got)
	}
}

func TestFindGraphQLSchema_GraphQLSubdir(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "graphql/schema.graphql", "type Query { ping: String }")

	got := FindGraphQLSchema(dir)
	if !strings.HasSuffix(got, filepath.Join("graphql", "schema.graphql")) {
		t.Errorf("expected graphql/schema.graphql path, got %q", got)
	}
}

func TestFindGraphQLSchema_APISubdir(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "api/schema.graphql", "type Query { ping: String }")

	got := FindGraphQLSchema(dir)
	if !strings.HasSuffix(got, filepath.Join("api", "schema.graphql")) {
		t.Errorf("expected api/schema.graphql path, got %q", got)
	}
}

func TestFindGraphQLSchema_PrefersSchemaGraphQLOverGQL(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "schema.graphql", "type Query { a: String }")
	writeFile(t, dir, "schema.gql", "type Query { b: String }")

	got := FindGraphQLSchema(dir)
	if !strings.HasSuffix(got, "schema.graphql") {
		t.Errorf("expected schema.graphql to win, got %q", got)
	}
}

// --------------------------------------------------------------------------
// GoGraphQL
// --------------------------------------------------------------------------

func TestGoGraphQL_NoSchema_ReturnsError(t *testing.T) {
	if err := GoGraphQL(tempDir(t), tempDir(t)); err == nil {
		t.Error("expected error when no schema file exists")
	}
}

func TestGoGraphQL_CopiesSchemaToOutputDir(t *testing.T) {
	src := tempDir(t)
	dst := tempDir(t)
	writeFile(t, src, "schema.graphql", "type Query { ping: String }")

	if err := GoGraphQL(src, dst); err != nil {
		t.Fatalf("GoGraphQL: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "schema.graphql"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(data) != "type Query { ping: String }" {
		t.Errorf("unexpected output content: %q", string(data))
	}
}

func TestGoGraphQL_CreatesOutputDir(t *testing.T) {
	src := tempDir(t)
	writeFile(t, src, "schema.graphql", "type Query { ping: String }")
	dst := filepath.Join(tempDir(t), "nested", "out")

	if err := GoGraphQL(src, dst); err != nil {
		t.Fatalf("GoGraphQL: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "schema.graphql")); err != nil {
		t.Error("expected output file to exist")
	}
}

// --------------------------------------------------------------------------
// FindProtoFile
// --------------------------------------------------------------------------

func TestFindProtoFile_Empty_ReturnsEmpty(t *testing.T) {
	if got := FindProtoFile(tempDir(t)); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFindProtoFile_RootProto(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "service.proto", `syntax = "proto3";`)

	got := FindProtoFile(dir)
	if !strings.HasSuffix(got, "service.proto") {
		t.Errorf("expected service.proto path, got %q", got)
	}
}

func TestFindProtoFile_ProtoSubdir(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "proto/user.proto", `syntax = "proto3";`)

	got := FindProtoFile(dir)
	if !strings.HasSuffix(got, "user.proto") {
		t.Errorf("expected user.proto path, got %q", got)
	}
}

func TestFindProtoFile_ProtosSubdir(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "protos/api.proto", `syntax = "proto3";`)

	got := FindProtoFile(dir)
	if !strings.HasSuffix(got, "api.proto") {
		t.Errorf("expected api.proto path, got %q", got)
	}
}

func TestFindProtoFile_SrcProtoSubdir(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "src/proto/svc.proto", `syntax = "proto3";`)

	got := FindProtoFile(dir)
	if !strings.HasSuffix(got, "svc.proto") {
		t.Errorf("expected svc.proto path, got %q", got)
	}
}

func TestFindProtoFile_GRPCSubdir(t *testing.T) {
	dir := tempDir(t)
	writeFile(t, dir, "grpc/svc.proto", `syntax = "proto3";`)

	got := FindProtoFile(dir)
	if !strings.HasSuffix(got, "svc.proto") {
		t.Errorf("expected svc.proto path, got %q", got)
	}
}

// --------------------------------------------------------------------------
// GoGRPC
// --------------------------------------------------------------------------

func TestGoGRPC_NoProto_ReturnsError(t *testing.T) {
	if err := GoGRPC(tempDir(t), tempDir(t)); err == nil {
		t.Error("expected error when no .proto file exists")
	}
}

func TestGoGRPC_CopiesProtoToOutputDir(t *testing.T) {
	src := tempDir(t)
	dst := tempDir(t)
	writeFile(t, src, "proto/service.proto", `syntax = "proto3"; service Svc {}`)

	if err := GoGRPC(src, dst); err != nil {
		t.Fatalf("GoGRPC: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "schema.proto"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(data) != `syntax = "proto3"; service Svc {}` {
		t.Errorf("unexpected output content: %q", string(data))
	}
}

func TestGoGRPC_CreatesOutputDir(t *testing.T) {
	src := tempDir(t)
	writeFile(t, src, "service.proto", `syntax = "proto3";`)
	dst := filepath.Join(tempDir(t), "nested", "out")

	if err := GoGRPC(src, dst); err != nil {
		t.Fatalf("GoGRPC: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "schema.proto")); err != nil {
		t.Error("expected output file to exist")
	}
}
