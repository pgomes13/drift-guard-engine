package compare_test

import (
	"testing"

	"github.com/pgomes13/api-drift-engine/pkg/compare"
)

const testdataDir = "../../internal/testdata/"

// --------------------------------------------------------------------------
// OpenAPI
// --------------------------------------------------------------------------

func TestOpenAPI_ReturnsResult(t *testing.T) {
	result, err := compare.OpenAPI(testdataDir+"base.yaml", testdataDir+"head.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BaseFile != testdataDir+"base.yaml" {
		t.Errorf("expected BaseFile=%q, got %q", testdataDir+"base.yaml", result.BaseFile)
	}
	if result.HeadFile != testdataDir+"head.yaml" {
		t.Errorf("expected HeadFile=%q, got %q", testdataDir+"head.yaml", result.HeadFile)
	}
}

func TestOpenAPI_DetectsBreakingChanges(t *testing.T) {
	result, err := compare.OpenAPI(testdataDir+"base.yaml", testdataDir+"head.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.Breaking == 0 {
		t.Error("expected at least one breaking change")
	}
}

func TestOpenAPI_DetectsNonBreakingChanges(t *testing.T) {
	result, err := compare.OpenAPI(testdataDir+"base.yaml", testdataDir+"head.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.NonBreaking == 0 {
		t.Error("expected at least one non-breaking change")
	}
}

func TestOpenAPI_IdenticalSchemas_NoChanges(t *testing.T) {
	result, err := compare.OpenAPI(testdataDir+"base.yaml", testdataDir+"base.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.Total != 0 {
		t.Errorf("expected 0 changes for identical schemas, got %d", result.Summary.Total)
	}
}

func TestOpenAPI_MissingBaseFile(t *testing.T) {
	_, err := compare.OpenAPI("/nonexistent/base.yaml", testdataDir+"head.yaml")
	if err == nil {
		t.Fatal("expected error for missing base file")
	}
}

func TestOpenAPI_MissingHeadFile(t *testing.T) {
	_, err := compare.OpenAPI(testdataDir+"base.yaml", "/nonexistent/head.yaml")
	if err == nil {
		t.Fatal("expected error for missing head file")
	}
}

// --------------------------------------------------------------------------
// GraphQL
// --------------------------------------------------------------------------

func TestGraphQL_ReturnsResult(t *testing.T) {
	result, err := compare.GraphQL(testdataDir+"base.graphql", testdataDir+"head.graphql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BaseFile != testdataDir+"base.graphql" {
		t.Errorf("expected BaseFile=%q, got %q", testdataDir+"base.graphql", result.BaseFile)
	}
}

func TestGraphQL_DetectsBreakingChanges(t *testing.T) {
	result, err := compare.GraphQL(testdataDir+"base.graphql", testdataDir+"head.graphql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.Breaking == 0 {
		t.Error("expected at least one breaking change")
	}
}

func TestGraphQL_DetectsNonBreakingChanges(t *testing.T) {
	result, err := compare.GraphQL(testdataDir+"base.graphql", testdataDir+"head.graphql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.NonBreaking == 0 {
		t.Error("expected at least one non-breaking change")
	}
}

func TestGraphQL_IdenticalSchemas_NoChanges(t *testing.T) {
	result, err := compare.GraphQL(testdataDir+"base.graphql", testdataDir+"base.graphql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.Total != 0 {
		t.Errorf("expected 0 changes for identical schemas, got %d", result.Summary.Total)
	}
}

func TestGraphQL_MissingBaseFile(t *testing.T) {
	_, err := compare.GraphQL("/nonexistent/base.graphql", testdataDir+"head.graphql")
	if err == nil {
		t.Fatal("expected error for missing base file")
	}
}

// --------------------------------------------------------------------------
// gRPC
// --------------------------------------------------------------------------

func TestGRPC_ReturnsResult(t *testing.T) {
	result, err := compare.GRPC(testdataDir+"base.proto", testdataDir+"head.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BaseFile != testdataDir+"base.proto" {
		t.Errorf("expected BaseFile=%q, got %q", testdataDir+"base.proto", result.BaseFile)
	}
}

func TestGRPC_DetectsBreakingChanges(t *testing.T) {
	result, err := compare.GRPC(testdataDir+"base.proto", testdataDir+"head.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.Breaking == 0 {
		t.Error("expected at least one breaking change")
	}
}

func TestGRPC_DetectsNonBreakingChanges(t *testing.T) {
	result, err := compare.GRPC(testdataDir+"base.proto", testdataDir+"head.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.NonBreaking == 0 {
		t.Error("expected at least one non-breaking change")
	}
}

func TestGRPC_IdenticalSchemas_NoChanges(t *testing.T) {
	result, err := compare.GRPC(testdataDir+"base.proto", testdataDir+"base.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.Total != 0 {
		t.Errorf("expected 0 changes for identical schemas, got %d", result.Summary.Total)
	}
}

func TestGRPC_MissingBaseFile(t *testing.T) {
	_, err := compare.GRPC("/nonexistent/base.proto", testdataDir+"head.proto")
	if err == nil {
		t.Fatal("expected error for missing base file")
	}
}
