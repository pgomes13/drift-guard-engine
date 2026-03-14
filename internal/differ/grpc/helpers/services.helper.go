// Package helpers provides internal diffing utilities for the gRPC differ.
package helpers

import "github.com/pgomes13/api-drift-engine/pkg/schema"

// IndexServices indexes a GRPCSchema's services by name.
func IndexServices(s *schema.GRPCSchema) map[string]schema.GRPCService {
	m := make(map[string]schema.GRPCService, len(s.Services))
	for _, svc := range s.Services {
		m[svc.Name] = svc
	}
	return m
}

// IndexMessages indexes a GRPCSchema's messages by name.
func IndexMessages(s *schema.GRPCSchema) map[string]schema.GRPCMessage {
	m := make(map[string]schema.GRPCMessage, len(s.Messages))
	for _, msg := range s.Messages {
		m[msg.Name] = msg
	}
	return m
}
