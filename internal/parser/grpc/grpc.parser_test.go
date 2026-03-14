package grpc_test

import (
	"testing"

	"github.com/DriftaBot/driftabot-engine/internal/parser/grpc"
	"github.com/DriftaBot/driftabot-engine/pkg/schema"
)

const testdataDir = "../../testdata/"

func TestParse_ReturnsSchema(t *testing.T) {
	s, err := grpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil schema")
	}
}

func TestParse_ServiceCount(t *testing.T) {
	s, err := grpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(s.Services))
	}
	if s.Services[0].Name != "UserService" {
		t.Errorf("expected service 'UserService', got '%s'", s.Services[0].Name)
	}
}

func TestParse_RPCCount(t *testing.T) {
	s, err := grpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc := findService(s, "UserService")
	if svc == nil {
		t.Fatal("service 'UserService' not found")
	}
	if len(svc.RPCs) != 3 {
		t.Errorf("expected 3 RPCs, got %d", len(svc.RPCs))
	}
}

func TestParse_RPCTypes(t *testing.T) {
	s, err := grpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc := findService(s, "UserService")
	rpc := findRPC(svc, "GetUser")
	if rpc == nil {
		t.Fatal("RPC 'GetUser' not found")
	}
	if rpc.RequestType != "GetUserRequest" {
		t.Errorf("expected RequestType='GetUserRequest', got '%s'", rpc.RequestType)
	}
	if rpc.ResponseType != "GetUserResponse" {
		t.Errorf("expected ResponseType='GetUserResponse', got '%s'", rpc.ResponseType)
	}
}

func TestParse_ServerStreaming(t *testing.T) {
	s, err := grpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc := findService(s, "UserService")
	rpc := findRPC(svc, "ListUsers")
	if rpc == nil {
		t.Fatal("RPC 'ListUsers' not found")
	}
	if !rpc.ServerStreaming {
		t.Error("expected ServerStreaming=true for ListUsers")
	}
	if rpc.ClientStreaming {
		t.Error("expected ClientStreaming=false for ListUsers")
	}
}

func TestParse_MessageCount(t *testing.T) {
	s, err := grpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Messages) < 5 {
		t.Errorf("expected at least 5 messages, got %d", len(s.Messages))
	}
}

func TestParse_MessageFields(t *testing.T) {
	s, err := grpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	msg := findMessage(s, "GetUserResponse")
	if msg == nil {
		t.Fatal("message 'GetUserResponse' not found")
	}
	expected := []string{"id", "name", "email"}
	for _, name := range expected {
		if findField(msg, name) == nil {
			t.Errorf("expected field '%s' on GetUserResponse", name)
		}
	}
}

func TestParse_FieldNumber(t *testing.T) {
	s, err := grpc.Parse(testdataDir + "base.proto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	msg := findMessage(s, "GetUserRequest")
	f := findField(msg, "id")
	if f == nil {
		t.Fatal("field 'id' not found on GetUserRequest")
	}
	if f.Number != 1 {
		t.Errorf("expected field number 1, got %d", f.Number)
	}
}

func TestParse_MissingFile(t *testing.T) {
	_, err := grpc.Parse("/nonexistent/path.proto")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

func findService(s *schema.GRPCSchema, name string) *schema.GRPCService {
	for i := range s.Services {
		if s.Services[i].Name == name {
			return &s.Services[i]
		}
	}
	return nil
}

func findRPC(svc *schema.GRPCService, name string) *schema.GRPCRPC {
	if svc == nil {
		return nil
	}
	for i := range svc.RPCs {
		if svc.RPCs[i].Name == name {
			return &svc.RPCs[i]
		}
	}
	return nil
}

func findMessage(s *schema.GRPCSchema, name string) *schema.GRPCMessage {
	for i := range s.Messages {
		if s.Messages[i].Name == name {
			return &s.Messages[i]
		}
	}
	return nil
}

func findField(msg *schema.GRPCMessage, name string) *schema.GRPCField {
	if msg == nil {
		return nil
	}
	for i := range msg.Fields {
		if msg.Fields[i].Name == name {
			return &msg.Fields[i]
		}
	}
	return nil
}
