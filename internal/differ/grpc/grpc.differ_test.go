package grpc_test

import (
	"testing"

	differgrpc "github.com/DriftaBot/driftabot-engine/internal/differ/grpc"
	parsergrpc "github.com/DriftaBot/driftabot-engine/internal/parser/grpc"
	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

const testdataDir = "../../testdata/"

// loadFixtures parses the base/head proto fixtures used across smoke tests.
func loadFixtures(t *testing.T) (base, head *schema.GRPCSchema) {
	t.Helper()
	var err error
	base, err = parsergrpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("parse base: %v", err)
	}
	head, err = parsergrpc.Parse(testdataDir + "head.proto")
	if err != nil {
		t.Fatalf("parse head: %v", err)
	}
	return
}

// ---------- fixtures ----------

func baseSchema() *schema.GRPCSchema {
	return &schema.GRPCSchema{
		Services: []schema.GRPCService{
			{
				Name: "UserService",
				RPCs: []schema.GRPCRPC{
					{Name: "GetUser", RequestType: "GetUserRequest", ResponseType: "GetUserResponse"},
					{Name: "ListUsers", RequestType: "ListUsersRequest", ResponseType: "ListUsersResponse", ServerStreaming: true},
					{Name: "DeleteUser", RequestType: "DeleteUserRequest", ResponseType: "DeleteUserResponse"},
				},
			},
		},
		Messages: []schema.GRPCMessage{
			{
				Name: "GetUserRequest",
				Fields: []schema.GRPCField{
					{Name: "id", Type: "string", Number: 1},
				},
			},
			{
				Name: "GetUserResponse",
				Fields: []schema.GRPCField{
					{Name: "id", Type: "string", Number: 1},
					{Name: "name", Type: "string", Number: 2},
					{Name: "email", Type: "string", Number: 3},
				},
			},
		},
	}
}

func headSchema() *schema.GRPCSchema {
	return &schema.GRPCSchema{
		Services: []schema.GRPCService{
			{
				Name: "UserService",
				RPCs: []schema.GRPCRPC{
					// GetUser request type changed
					{Name: "GetUser", RequestType: "GetUserRequestV2", ResponseType: "GetUserResponse"},
					// ListUsers streaming mode changed to bidirectional
					{Name: "ListUsers", RequestType: "ListUsersRequest", ResponseType: "ListUsersResponse", ClientStreaming: true, ServerStreaming: true},
					// DeleteUser removed
					// CreateUser added
					{Name: "CreateUser", RequestType: "CreateUserRequest", ResponseType: "CreateUserResponse"},
				},
			},
			// AdminService added
			{
				Name: "AdminService",
				RPCs: []schema.GRPCRPC{
					{Name: "BanUser", RequestType: "BanUserRequest", ResponseType: "BanUserResponse"},
				},
			},
		},
		Messages: []schema.GRPCMessage{
			{
				Name: "GetUserRequest",
				Fields: []schema.GRPCField{
					{Name: "id", Type: "string", Number: 1},
				},
			},
			{
				Name: "GetUserResponse",
				Fields: []schema.GRPCField{
					{Name: "id", Type: "string", Number: 1},
					// name type changed
					{Name: "name", Type: "bytes", Number: 2},
					// email removed
					// role added
					{Name: "role", Type: "string", Number: 4},
				},
			},
		},
	}
}

func diff(t *testing.T) []schema.Change {
	t.Helper()
	return differgrpc.Diff(baseSchema(), headSchema())
}

func findChange(changes []schema.Change, ct schema.ChangeType, locSubstr string) *schema.Change {
	for i := range changes {
		if changes[i].Type == ct && (locSubstr == "" || containsSubstr(changes[i].Location, locSubstr)) {
			return &changes[i]
		}
	}
	return nil
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ---------- service-level ----------

func TestDiffGRPC_ServiceAdded(t *testing.T) {
	c := findChange(diff(t), schema.ChangeTypeGRPCServiceAdded, "AdminService")
	if c == nil {
		t.Error("expected grpc_service_added for AdminService")
	}
}

func TestDiffGRPC_ServiceRemoved(t *testing.T) {
	// Remove the only service to verify detection
	base := &schema.GRPCSchema{
		Services: []schema.GRPCService{{Name: "OldService"}},
	}
	head := &schema.GRPCSchema{}
	changes := differgrpc.Diff(base, head)
	c := findChange(changes, schema.ChangeTypeGRPCServiceRemoved, "OldService")
	if c == nil {
		t.Error("expected grpc_service_removed for OldService")
	}
}

// ---------- RPC-level ----------

func TestDiffGRPC_RPCRemoved(t *testing.T) {
	c := findChange(diff(t), schema.ChangeTypeGRPCRPCRemoved, "DeleteUser")
	if c == nil {
		t.Error("expected grpc_rpc_removed for DeleteUser")
	}
}

func TestDiffGRPC_RPCAdded(t *testing.T) {
	c := findChange(diff(t), schema.ChangeTypeGRPCRPCAdded, "CreateUser")
	if c == nil {
		t.Error("expected grpc_rpc_added for CreateUser")
	}
}

func TestDiffGRPC_RPCRequestTypeChanged(t *testing.T) {
	c := findChange(diff(t), schema.ChangeTypeGRPCRPCRequestTypeChanged, "GetUser")
	if c == nil {
		t.Fatal("expected grpc_rpc_request_type_changed for GetUser")
	}
	if c.Before != "GetUserRequest" {
		t.Errorf("expected Before='GetUserRequest', got '%s'", c.Before)
	}
	if c.After != "GetUserRequestV2" {
		t.Errorf("expected After='GetUserRequestV2', got '%s'", c.After)
	}
}

func TestDiffGRPC_RPCStreamingChanged(t *testing.T) {
	c := findChange(diff(t), schema.ChangeTypeGRPCRPCStreamingChanged, "ListUsers")
	if c == nil {
		t.Fatal("expected grpc_rpc_streaming_changed for ListUsers")
	}
	if c.Before != "server streaming" {
		t.Errorf("expected Before='server streaming', got '%s'", c.Before)
	}
	if c.After != "bidirectional streaming" {
		t.Errorf("expected After='bidirectional streaming', got '%s'", c.After)
	}
}

// ---------- message/field-level ----------

func TestDiffGRPC_FieldRemoved(t *testing.T) {
	c := findChange(diff(t), schema.ChangeTypeGRPCFieldRemoved, "GetUserResponse.email")
	if c == nil {
		t.Error("expected grpc_field_removed for GetUserResponse.email")
	}
}

func TestDiffGRPC_FieldAdded(t *testing.T) {
	c := findChange(diff(t), schema.ChangeTypeGRPCFieldAdded, "GetUserResponse.role")
	if c == nil {
		t.Error("expected grpc_field_added for GetUserResponse.role")
	}
}

func TestDiffGRPC_FieldTypeChanged(t *testing.T) {
	c := findChange(diff(t), schema.ChangeTypeGRPCFieldTypeChanged, "GetUserResponse.name")
	if c == nil {
		t.Fatal("expected grpc_field_type_changed for GetUserResponse.name")
	}
	if c.Before != "string" || c.After != "bytes" {
		t.Errorf("expected string→bytes, got %s→%s", c.Before, c.After)
	}
}

func TestDiffGRPC_MessageAdded(t *testing.T) {
	base := &schema.GRPCSchema{}
	head := &schema.GRPCSchema{
		Messages: []schema.GRPCMessage{{Name: "NewMessage"}},
	}
	changes := differgrpc.Diff(base, head)
	c := findChange(changes, schema.ChangeTypeGRPCMessageAdded, "NewMessage")
	if c == nil {
		t.Error("expected grpc_message_added for NewMessage")
	}
}

func TestDiffGRPC_MessageRemoved(t *testing.T) {
	base := &schema.GRPCSchema{
		Messages: []schema.GRPCMessage{{Name: "OldMessage"}},
	}
	head := &schema.GRPCSchema{}
	changes := differgrpc.Diff(base, head)
	c := findChange(changes, schema.ChangeTypeGRPCMessageRemoved, "OldMessage")
	if c == nil {
		t.Error("expected grpc_message_removed for OldMessage")
	}
}

// ---------- sanity ----------

func TestDiffGRPC_IdenticalSchemas_NoChanges(t *testing.T) {
	base := baseSchema()
	changes := differgrpc.Diff(base, base)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for identical schemas, got %d: %v", len(changes), changes)
	}
}

// --------------------------------------------------------------------------
// Smoke tests — full parser → differ pipeline through .proto fixtures
// --------------------------------------------------------------------------

func TestDiffGRPC_Smoke_RPCRemoved(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgrpc.Diff(base, head)

	// DeleteUser removed in head.proto
	c := findChange(changes, schema.ChangeTypeGRPCRPCRemoved, "DeleteUser")
	if c == nil {
		t.Error("expected grpc_rpc_removed for UserService.DeleteUser")
	}
}

func TestDiffGRPC_Smoke_RPCAdded(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgrpc.Diff(base, head)

	// CreateUser added in head.proto
	c := findChange(changes, schema.ChangeTypeGRPCRPCAdded, "CreateUser")
	if c == nil {
		t.Error("expected grpc_rpc_added for UserService.CreateUser")
	}
}

func TestDiffGRPC_Smoke_RPCRequestTypeChanged(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgrpc.Diff(base, head)

	// GetUser request type: GetUserRequest → GetUserRequestV2
	c := findChange(changes, schema.ChangeTypeGRPCRPCRequestTypeChanged, "GetUser")
	if c == nil {
		t.Fatal("expected grpc_rpc_request_type_changed for UserService.GetUser")
	}
	if c.Before != "GetUserRequest" {
		t.Errorf("expected Before='GetUserRequest', got '%s'", c.Before)
	}
	if c.After != "GetUserRequestV2" {
		t.Errorf("expected After='GetUserRequestV2', got '%s'", c.After)
	}
}

func TestDiffGRPC_Smoke_RPCStreamingChanged(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgrpc.Diff(base, head)

	// ListUsers: server streaming → bidirectional streaming
	c := findChange(changes, schema.ChangeTypeGRPCRPCStreamingChanged, "ListUsers")
	if c == nil {
		t.Fatal("expected grpc_rpc_streaming_changed for UserService.ListUsers")
	}
	if c.Before != "server streaming" {
		t.Errorf("expected Before='server streaming', got '%s'", c.Before)
	}
	if c.After != "bidirectional streaming" {
		t.Errorf("expected After='bidirectional streaming', got '%s'", c.After)
	}
}

func TestDiffGRPC_Smoke_ServiceAdded(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgrpc.Diff(base, head)

	// AdminService added in head.proto
	c := findChange(changes, schema.ChangeTypeGRPCServiceAdded, "AdminService")
	if c == nil {
		t.Error("expected grpc_service_added for AdminService")
	}
}

func TestDiffGRPC_Smoke_FieldRemoved(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgrpc.Diff(base, head)

	// GetUserResponse.email removed in head.proto
	c := findChange(changes, schema.ChangeTypeGRPCFieldRemoved, "GetUserResponse.email")
	if c == nil {
		t.Error("expected grpc_field_removed for GetUserResponse.email")
	}
}

func TestDiffGRPC_Smoke_FieldAdded(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgrpc.Diff(base, head)

	// GetUserResponse.role added in head.proto
	c := findChange(changes, schema.ChangeTypeGRPCFieldAdded, "GetUserResponse.role")
	if c == nil {
		t.Error("expected grpc_field_added for GetUserResponse.role")
	}
}

func TestDiffGRPC_Smoke_TotalChanges(t *testing.T) {
	base, head := loadFixtures(t)
	changes := differgrpc.Diff(base, head)

	if len(changes) == 0 {
		t.Fatal("expected at least one change between base and head fixtures")
	}

	byType := map[schema.ChangeType]int{}
	for _, c := range changes {
		byType[c.Type]++
	}

	required := []schema.ChangeType{
		schema.ChangeTypeGRPCRPCRemoved,            // DeleteUser
		schema.ChangeTypeGRPCRPCAdded,              // CreateUser
		schema.ChangeTypeGRPCRPCRequestTypeChanged, // GetUser
		schema.ChangeTypeGRPCRPCStreamingChanged,   // ListUsers
		schema.ChangeTypeGRPCServiceAdded,          // AdminService
		schema.ChangeTypeGRPCFieldRemoved,          // GetUserResponse.email, ListUsersResponse.total
		schema.ChangeTypeGRPCFieldAdded,            // GetUserResponse.role
	}

	for _, ct := range required {
		if byType[ct] == 0 {
			t.Errorf("expected at least one change of type %s", ct)
		}
	}
}

func TestDiffGRPC_Smoke_IdenticalSchemas_NoChanges(t *testing.T) {
	base, err := parsergrpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	changes := differgrpc.Diff(base, base)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for identical schemas, got %d: %v", len(changes), changes)
	}
}
