package main

import (
	"os"
	"path/filepath"
	"testing"
)

// --------------------------------------------------------------------------
// hasProtoFiles
// --------------------------------------------------------------------------

func TestHasProtoFiles_False_EmptyDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "drift-agent-ui-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	if hasProtoFiles(dir) {
		t.Error("expected false for empty dir")
	}
}

func TestHasProtoFiles_True_RootProto(t *testing.T) {
	dir, err := os.MkdirTemp("", "drift-agent-ui-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "service.proto"), []byte(`syntax = "proto3";`), 0o644); err != nil {
		t.Fatalf("write proto: %v", err)
	}

	if !hasProtoFiles(dir) {
		t.Error("expected true when .proto file is present")
	}
}

func TestHasProtoFiles_True_NestedProto(t *testing.T) {
	dir, err := os.MkdirTemp("", "drift-agent-ui-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	sub := filepath.Join(dir, "proto", "v1")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "user.proto"), []byte(`syntax = "proto3";`), 0o644); err != nil {
		t.Fatalf("write proto: %v", err)
	}

	if !hasProtoFiles(dir) {
		t.Error("expected true for nested .proto file")
	}
}

func TestHasProtoFiles_False_OnlyNonProtoFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "drift-agent-ui-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "schema.graphql"), []byte("type Query {}"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if hasProtoFiles(dir) {
		t.Error("expected false when no .proto files present")
	}
}

// --------------------------------------------------------------------------
// detectAPITypes
// --------------------------------------------------------------------------

func TestDetectAPITypes_AlwaysIncludesREST(t *testing.T) {
	dir, err := os.MkdirTemp("", "drift-agent-ui-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	types := detectAPITypes(dir)
	if len(types) == 0 || types[0] != "REST" {
		t.Errorf("expected REST as first type, got %v", types)
	}
}

func TestDetectAPITypes_OnlyREST_EmptyDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "drift-agent-ui-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	types := detectAPITypes(dir)
	if len(types) != 1 {
		t.Errorf("expected only REST for empty dir, got %v", types)
	}
}

func TestDetectAPITypes_IncludesGRPC_WhenProtoPresent(t *testing.T) {
	dir, err := os.MkdirTemp("", "drift-agent-ui-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "service.proto"), []byte(`syntax = "proto3";`), 0o644); err != nil {
		t.Fatalf("write proto: %v", err)
	}

	types := detectAPITypes(dir)
	if !contains(types, "gRPC") {
		t.Errorf("expected gRPC in types, got %v", types)
	}
}

func TestDetectAPITypes_IncludesGraphQL_WhenSchemaPresent(t *testing.T) {
	dir, err := os.MkdirTemp("", "drift-agent-ui-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	// Write a package.json with a GraphQL dep so DetectGraphQLInfo picks it up.
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
		"dependencies": {
			"express": "^4.0.0",
			"apollo-server-express": "^3.0.0"
		}
	}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	types := detectAPITypes(dir)
	if !contains(types, "GraphQL") {
		t.Errorf("expected GraphQL in types, got %v", types)
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
